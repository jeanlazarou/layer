// Package sourcefile splits a .layer source file into its editor-facing
// components (enabled layers, per-layer source code, main expression) using
// the real lexer and parser instead of regular expressions. Braces inside
// strings or comments never confuse the block extraction because slicing is
// driven by token positions.
package sourcefile

import (
	"strings"

	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

// File is the result of splitting a .layer source file.
type File struct {
	// EnabledLayers holds the layer names from the @layers directive, in
	// declaration order (conditional entries are included by name).
	EnabledLayers []string `json:"enabledLayers"`
	// DeclaredLayers holds the names of all @layer blocks in source order,
	// including blocks not listed in @layers (inactive layers).
	DeclaredLayers []string `json:"declaredLayers"`
	// Layers maps each @layer block name to its dedented body source.
	Layers map[string]string `json:"layers"`
	// Purity maps a layer name to its explicit annotation ("pure" or
	// "impure") when the block was written as @layer<...> name.
	Purity map[string]string `json:"purity"`
	// BlendMode is the default @layers blend mode (`blend: chain`, or the
	// `_` entry of a blend map), or "" when not specified.
	BlendMode string `json:"blendMode"`
	// BlendModes holds per-function modes from `blend: {fetch: fallback}`.
	BlendModes map[string]string `json:"blendModes"`
	// MainExpression is the source text after the last @layer block (or
	// after the @layers directive when there are no blocks).
	MainExpression string `json:"mainExpression"`
}

// Split parses source and returns its components. The source must be a
// valid Layer program; lexer or parser errors are returned as-is.
func Split(source string) (*File, error) {
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		return nil, err
	}

	program, err := parser.Parse(tokens)
	if err != nil {
		return nil, err
	}

	file := &File{
		EnabledLayers:  []string{},
		DeclaredLayers: []string{},
		Layers:         make(map[string]string),
		Purity:         make(map[string]string),
		BlendModes:     make(map[string]string),
	}

	if program.LayersDirective != nil {
		for _, entry := range program.LayersDirective.Layers {
			file.EnabledLayers = append(file.EnabledLayers, entry.Name)
		}
		file.BlendMode = program.LayersDirective.BlendMode
		for _, entry := range program.LayersDirective.BlendEntries {
			file.BlendModes[entry.Name] = entry.Mode
		}
	}
	for _, decl := range program.LayerDeclarations {
		file.DeclaredLayers = append(file.DeclaredLayers, decl.LayerName)
		switch decl.Purity {
		case "":
		case "pure":
			file.Purity[decl.LayerName] = "pure"
		default:
			file.Purity[decl.LayerName] = "impure"
		}
	}

	runes := []rune(source)
	lineStarts := runeLineStarts(runes)
	offsetOf := func(t lexer.Token) int {
		line := t.Line
		if line < 1 {
			line = 1
		}
		if line > len(lineStarts) {
			return len(runes)
		}
		offset := lineStarts[line-1] + t.Column - 1
		if offset > len(runes) {
			return len(runes)
		}
		return offset
	}

	// Walk tokens to find @layer <name> { ... } blocks and slice their
	// bodies out of the original source.
	mainStart := 0
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type != lexer.AT {
			continue
		}
		next, nameTok, brace := nextSignificant(tokens, i)
		switch next {
		case lexer.LAYERS:
			// Skip past the directive so a file with no @layer blocks
			// reports only the trailing code as its main expression.
			if end := directiveEnd(tokens, i); end > 0 {
				if pos := offsetOf(tokens[end]) + tokenWidth(tokens[end]); pos > mainStart {
					mainStart = pos
				}
			}
		case lexer.LAYER:
			if brace < 0 {
				continue
			}
			closing := matchingBrace(tokens, brace)
			if closing < 0 {
				continue
			}
			name, _ := nameTok.Value.(string)
			body := string(runes[offsetOf(tokens[brace])+1 : offsetOf(tokens[closing])])
			file.Layers[name] = dedent(body)
			if pos := offsetOf(tokens[closing]) + 1; pos > mainStart {
				mainStart = pos
			}
			i = closing
		}
	}

	if mainStart < len(runes) {
		file.MainExpression = strings.TrimSpace(string(runes[mainStart:]))
	}

	return file, nil
}

// nextSignificant inspects the tokens after an AT token and, for @layer
// blocks, locates the name token and the opening brace. Comment tokens are
// skipped.
func nextSignificant(tokens []lexer.Token, atIndex int) (lexer.TokenType, lexer.Token, int) {
	i := skipComments(tokens, atIndex+1)
	if i >= len(tokens) {
		return lexer.EOF, lexer.Token{}, -1
	}
	kind := tokens[i].Type
	if kind != lexer.LAYER {
		return kind, lexer.Token{}, -1
	}

	nameIdx := skipComments(tokens, i+1)
	// Optional purity annotation: @layer<impure> name
	if nameIdx+2 < len(tokens) && tokens[nameIdx].Type == lexer.LT &&
		tokens[nameIdx+1].Type == lexer.IDENTIFIER && tokens[nameIdx+2].Type == lexer.GT {
		nameIdx = skipComments(tokens, nameIdx+3)
	}
	if nameIdx >= len(tokens) || tokens[nameIdx].Type != lexer.IDENTIFIER {
		return kind, lexer.Token{}, -1
	}

	braceIdx := skipComments(tokens, nameIdx+1)
	if braceIdx >= len(tokens) || tokens[braceIdx].Type != lexer.LBRACE {
		return kind, tokens[nameIdx], -1
	}

	return kind, tokens[nameIdx], braceIdx
}

// directiveEnd returns the index of the last token belonging to the
// @layers [...] directive (including an optional trailing blend clause:
// `blend: mode` or `blend: {name: mode, ...}`), or -1 if the directive is
// malformed.
func directiveEnd(tokens []lexer.Token, atIndex int) int {
	i := skipComments(tokens, atIndex+1) // LAYERS
	i = skipComments(tokens, i+1)        // expect LBRACKET
	if i >= len(tokens) || tokens[i].Type != lexer.LBRACKET {
		return -1
	}

	depth := 0
	for ; i < len(tokens); i++ {
		switch tokens[i].Type {
		case lexer.LBRACKET:
			depth++
		case lexer.RBRACKET:
			depth--
			if depth == 0 {
				end := i
				// Optional: blend : <mode>
				j := skipComments(tokens, end+1)
				if j < len(tokens) && tokens[j].Type == lexer.IDENTIFIER {
					if name, _ := tokens[j].Value.(string); name == "blend" {
						k := skipComments(tokens, j+1)
						if k < len(tokens) && tokens[k].Type == lexer.COLON {
							m := skipComments(tokens, k+1)
							if m < len(tokens) && tokens[m].Type == lexer.IDENTIFIER {
								end = m
							} else if m < len(tokens) && tokens[m].Type == lexer.LBRACE {
								if closing := matchingBrace(tokens, m); closing > 0 {
									end = closing
								}
							}
						}
					}
				}
				return end
			}
		}
	}
	return -1
}

// matchingBrace returns the index of the RBRACE that closes the LBRACE at
// openIndex, or -1 when unbalanced.
func matchingBrace(tokens []lexer.Token, openIndex int) int {
	depth := 0
	for i := openIndex; i < len(tokens); i++ {
		switch tokens[i].Type {
		case lexer.LBRACE:
			depth++
		case lexer.RBRACE:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func skipComments(tokens []lexer.Token, i int) int {
	for i < len(tokens) && tokens[i].Type == lexer.COMMENT {
		i++
	}
	return i
}

// tokenWidth returns the width in runes of a token's source text. Only
// needed for the directive's final token (RBRACKET or a blend-mode
// identifier).
func tokenWidth(t lexer.Token) int {
	if s, ok := t.Value.(string); ok {
		return len([]rune(s))
	}
	return 1
}

// runeLineStarts returns the rune index at which each line starts.
func runeLineStarts(runes []rune) []int {
	starts := []int{0}
	for i, r := range runes {
		if r == '\n' {
			starts = append(starts, i+1)
		}
	}
	return starts
}

// dedent removes the common leading indentation from a code block.
func dedent(code string) string {
	lines := strings.Split(code, "\n")

	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.TrimSpace(code)
	}

	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if len(line) >= minIndent {
			result = append(result, line[minIndent:])
		} else {
			result = append(result, line)
		}
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}
