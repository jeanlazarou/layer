package parser

import (
	"fmt"

	"github.com/jeanlazarou/layer/pkg/ast"
	"github.com/jeanlazarou/layer/pkg/lexer"
)

// Parser is a recursive descent parser for the Layer language
type Parser struct {
	tokens []lexer.Token
	pos    int
	gensym int // program-wide counter for generated parameter names (partial application)
}

// NewParser creates a new parser from a list of tokens
func NewParser(tokens []lexer.Token) *Parser {
	// Filter out comments
	filtered := make([]lexer.Token, 0, len(tokens))
	for _, t := range tokens {
		if t.Type != lexer.COMMENT {
			filtered = append(filtered, t)
		}
	}
	return &Parser{
		tokens: filtered,
		pos:    0,
	}
}

// error raises a parser error with position information
func (p *Parser) error(message string) error {
	token := p.current()
	return fmt.Errorf("parser error at %d:%d: %s", token.Line, token.Column, message)
}

// current returns the current token
func (p *Parser) current() lexer.Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return p.tokens[len(p.tokens)-1] // EOF
}

// peek looks ahead at a token
func (p *Parser) peek(offset int) lexer.Token {
	pos := p.pos + offset
	if pos < len(p.tokens) {
		return p.tokens[pos]
	}
	return p.tokens[len(p.tokens)-1] // EOF
}

// advance consumes and returns the current token
func (p *Parser) advance() lexer.Token {
	token := p.current()
	if p.pos < len(p.tokens)-1 {
		p.pos++
	}
	return token
}

// expect expects a specific token type and consumes it
func (p *Parser) expect(tokenType lexer.TokenType) (lexer.Token, error) {
	token := p.current()
	if token.Type != tokenType {
		return token, p.error(fmt.Sprintf("expected %s, got %s", tokenType, token.Type))
	}
	return p.advance(), nil
}

// match checks if current token matches any of the given types
func (p *Parser) match(types ...lexer.TokenType) bool {
	current := p.current().Type
	for _, t := range types {
		if current == t {
			return true
		}
	}
	return false
}

// Parse parses the entire program
func (p *Parser) Parse() (*ast.Program, error) {
	var layersDirective *ast.LayersDirective
	var layerDeclarations []*ast.LayerDeclaration
	var mainExpression ast.Node

	// Parse @layers directive if present
	if p.match(lexer.AT) && p.peek(1).Type == lexer.LAYERS {
		directive, err := p.parseLayersDirective()
		if err != nil {
			return nil, err
		}
		layersDirective = directive
	}

	// Parse layer declarations
	for p.match(lexer.AT) && p.peek(1).Type == lexer.LAYER {
		decl, err := p.parseLayerDeclaration()
		if err != nil {
			return nil, err
		}
		layerDeclarations = append(layerDeclarations, decl)
	}

	// Parse main expression if present
	if !p.match(lexer.EOF) {
		expr, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		mainExpression = expr
	}

	return &ast.Program{
		Base:               ast.Base{Line: 1, Column: 1},
		LayersDirective:    layersDirective,
		LayerDeclarations:  layerDeclarations,
		MainExpression:     mainExpression,
	}, nil
}

// parseLayersDirective parses @layers [core, ui, debug] directive
func (p *Parser) parseLayersDirective() (*ast.LayersDirective, error) {
	start, err := p.expect(lexer.AT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.LAYERS); err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.LBRACKET); err != nil {
		return nil, err
	}

	var layers []ast.LayerEntry
	for !p.match(lexer.RBRACKET) {
		nameToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}

		// Optional condition: <name> if <expr>
		var condition ast.Node
		if p.current().Type == lexer.IDENTIFIER && p.current().Value != nil && p.current().Value.(string) == "if" {
			p.advance() // consume "if"
			condition, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}

		layers = append(layers, ast.LayerEntry{
			Name:      nameToken.Value.(string),
			Condition: condition,
		})

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}

	// Optional blend mode: blend: <mode>
	var blendMode string
	if p.current().Type == lexer.IDENTIFIER && p.current().Value != nil && p.current().Value.(string) == "blend" {
		p.advance() // consume "blend"
		if _, err := p.expect(lexer.COLON); err != nil {
			return nil, err
		}
		modeToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		blendMode = modeToken.Value.(string)
		validModes := map[string]bool{"override": true, "chain": true, "collect": true, "fallback": true}
		if !validModes[blendMode] {
			return nil, p.error(fmt.Sprintf("unknown blend mode '%s': must be override, chain, collect, or fallback", blendMode))
		}
	}

	return &ast.LayersDirective{
		Base:      ast.Base{Line: start.Line, Column: start.Column},
		Layers:    layers,
		BlendMode: blendMode,
	}, nil
}

// parseLayerDeclaration parses @layer core { ... } declaration
func (p *Parser) parseLayerDeclaration() (*ast.LayerDeclaration, error) {
	start, err := p.expect(lexer.AT)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.LAYER); err != nil {
		return nil, err
	}

	// Optional purity annotation: @layer<pure|impure|ui|data|debug> name
	var purity string
	if p.match(lexer.LT) {
		p.advance() // consume <
		purityToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		purity = purityToken.Value.(string)
		validPurity := map[string]bool{"pure": true, "impure": true, "ui": true, "data": true, "debug": true}
		if !validPurity[purity] {
			return nil, p.error(fmt.Sprintf("unknown purity annotation '%s': must be pure, impure, ui, data, or debug", purity))
		}
		if _, err := p.expect(lexer.GT); err != nil {
			return nil, err
		}
	}

	nameToken, err := p.expect(lexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.LBRACE); err != nil {
		return nil, err
	}

	var definitions []ast.Node
	for !p.match(lexer.RBRACE) {
		def, err := p.parseDefinition()
		if err != nil {
			return nil, err
		}
		definitions = append(definitions, def)
	}

	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}

	return &ast.LayerDeclaration{
		Base:        ast.Base{Line: start.Line, Column: start.Column},
		LayerName:   nameToken.Value.(string),
		Definitions: definitions,
		Purity:      purity,
	}, nil
}

// parseDefinition parses a top-level definition (function, etc.)
func (p *Parser) parseDefinition() (ast.Node, error) {
	if p.match(lexer.FN) {
		return p.parseFunctionDef()
	}
	return p.parseExpression()
}

// parseFunctionDef parses function definition: fn name(params) = body
func (p *Parser) parseFunctionDef() (*ast.FunctionDef, error) {
	start, err := p.expect(lexer.FN)
	if err != nil {
		return nil, err
	}

	nameToken, err := p.expect(lexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	// Parse parameters
	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}

	var params []string
	for !p.match(lexer.RPAREN) {
		param, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		params = append(params, param.Value.(string))

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}

	// Parse body
	body, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.FunctionDef{
		Base:   ast.Base{Line: start.Line, Column: start.Column},
		Name:   nameToken.Value.(string),
		Params: params,
		Body:   body,
	}, nil
}

// parseExpression parses an expression
func (p *Parser) parseExpression() (ast.Node, error) {
	// Let binding
	if p.match(lexer.LET) {
		return p.parseLetBinding()
	}

	// Match expression
	if p.match(lexer.MATCH) {
		return p.parseMatch()
	}

	// Lambda
	if p.match(lexer.LAMBDA) {
		return p.parseLambda()
	}

	// Binary operations and other expressions
	return p.parseBinaryOp(0)
}

// parseLetBinding parses let binding: let x = value in body
func (p *Parser) parseLetBinding() (*ast.LetBinding, error) {
	start, err := p.expect(lexer.LET)
	if err != nil {
		return nil, err
	}

	nameToken, err := p.expect(lexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.ASSIGN); err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.IN); err != nil {
		return nil, err
	}

	body, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.LetBinding{
		Base:  ast.Base{Line: start.Line, Column: start.Column},
		Name:  nameToken.Value.(string),
		Value: value,
		Body:  body,
	}, nil
}

// parseMatch parses match expression
func (p *Parser) parseMatch() (*ast.Match, error) {
	start, err := p.expect(lexer.MATCH)
	if err != nil {
		return nil, err
	}

	value, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.WITH); err != nil {
		return nil, err
	}

	var cases []*ast.MatchCase
	for p.match(lexer.PIPE) {
		p.advance() // consume |

		pattern, err := p.parsePattern()
		if err != nil {
			return nil, err
		}

		// Optional guard: when <expr>
		var guard ast.Node
		if p.current().Type == lexer.IDENTIFIER && p.current().Value != nil && p.current().Value.(string) == "when" {
			p.advance() // consume "when"
			guard, err = p.parseExpression()
			if err != nil {
				return nil, err
			}
		}

		if _, err := p.expect(lexer.ARROW); err != nil {
			return nil, err
		}

		body, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		cases = append(cases, &ast.MatchCase{
			Base:    ast.Base{Line: pattern.Pos().Line, Column: pattern.Pos().Column},
			Pattern: pattern,
			Guard:   guard,
			Body:    body,
		})
	}

	return &ast.Match{
		Base:  ast.Base{Line: start.Line, Column: start.Column},
		Value: value,
		Cases: cases,
	}, nil
}

// parsePattern parses a pattern in a match expression
func (p *Parser) parsePattern() (ast.Node, error) {
	token := p.current()

	// Number pattern
	if p.match(lexer.NUMBER) {
		numToken := p.advance()
		return &ast.NumberPattern{
			Base:  ast.Base{Line: numToken.Line, Column: numToken.Column},
			Value: numToken.Value,
		}, nil
	}

	// String pattern
	if p.match(lexer.STRING) {
		strToken := p.advance()
		return &ast.StringPattern{
			Base:  ast.Base{Line: strToken.Line, Column: strToken.Column},
			Value: strToken.Value.(string),
		}, nil
	}

	// Boolean pattern
	if p.match(lexer.BOOLEAN) {
		boolToken := p.advance()
		return &ast.BooleanPattern{
			Base:  ast.Base{Line: boolToken.Line, Column: boolToken.Column},
			Value: boolToken.Value.(bool),
		}, nil
	}

	// List pattern
	if p.match(lexer.LBRACKET) {
		return p.parseListPattern()
	}

	// Record pattern
	if p.match(lexer.LBRACE) {
		return p.parseRecordPattern()
	}

	// Identifier pattern (variable binding or wildcard)
	if p.match(lexer.IDENTIFIER) {
		idToken := p.advance()
		return &ast.IdentifierPattern{
			Base: ast.Base{Line: idToken.Line, Column: idToken.Column},
			Name: idToken.Value.(string),
		}, nil
	}

	return nil, p.error(fmt.Sprintf("expected pattern, got %s", token.Type))
}

// parseListPattern parses list pattern: [a, b, ...rest]
func (p *Parser) parseListPattern() (*ast.ListPattern, error) {
	start, err := p.expect(lexer.LBRACKET)
	if err != nil {
		return nil, err
	}

	var elements []ast.Node
	var rest string

	for !p.match(lexer.RBRACKET) {
		if p.match(lexer.ELLIPSIS) {
			p.advance()
			restToken, err := p.expect(lexer.IDENTIFIER)
			if err != nil {
				return nil, err
			}
			rest = restToken.Value.(string)
			break
		} else {
			pattern, err := p.parsePattern()
			if err != nil {
				return nil, err
			}
			elements = append(elements, pattern)
		}

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}

	return &ast.ListPattern{
		Base:     ast.Base{Line: start.Line, Column: start.Column},
		Elements: elements,
		Rest:     rest,
	}, nil
}

// parseRecordPattern parses record pattern: {name: n, age: a}
// Fields are partial — extra fields in the value are ignored.
// Each field value is a sub-pattern.
func (p *Parser) parseRecordPattern() (*ast.RecordPattern, error) {
	start, err := p.expect(lexer.LBRACE)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]ast.Node)
	for !p.match(lexer.RBRACE) {
		keyToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}

		if _, err := p.expect(lexer.COLON); err != nil {
			return nil, err
		}

		subPattern, err := p.parsePattern()
		if err != nil {
			return nil, err
		}

		fields[keyToken.Value.(string)] = subPattern

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}

	return &ast.RecordPattern{
		Base:   ast.Base{Line: start.Line, Column: start.Column},
		Fields: fields,
	}, nil
}

// parseLambda parses lambda: λ(x, y) -> body
func (p *Parser) parseLambda() (*ast.Lambda, error) {
	start, err := p.expect(lexer.LAMBDA)
	if err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.LPAREN); err != nil {
		return nil, err
	}

	var params []string
	for !p.match(lexer.RPAREN) {
		param, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		params = append(params, param.Value.(string))

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RPAREN); err != nil {
		return nil, err
	}

	if _, err := p.expect(lexer.ARROW); err != nil {
		return nil, err
	}

	body, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	return &ast.Lambda{
		Base:   ast.Base{Line: start.Line, Column: start.Column},
		Params: params,
		Body:   body,
	}, nil
}

// parseBinaryOp parses binary operations with precedence climbing
func (p *Parser) parseBinaryOp(precedence int) (ast.Node, error) {
	left, err := p.parseUnaryOp()
	if err != nil {
		return nil, err
	}

	for {
		opToken := p.current()
		opPrecedence := p.getPrecedence(opToken.Type)

		// Stop if not a binary operator (precedence 0) or precedence too low
		if opPrecedence == 0 || opPrecedence < precedence {
			break
		}

		operator := p.advance()
		right, err := p.parseBinaryOp(opPrecedence + 1)
		if err != nil {
			return nil, err
		}

		left = &ast.BinaryOp{
			Base:     ast.Base{Line: operator.Line, Column: operator.Column},
			Operator: operator.Value.(string),
			Left:     left,
			Right:    right,
		}
	}

	return left, nil
}

// parseUnaryOp parses unary operations
func (p *Parser) parseUnaryOp() (ast.Node, error) {
	if p.match(lexer.MINUS, lexer.NOT) {
		opToken := p.advance()
		operand, err := p.parseUnaryOp()
		if err != nil {
			return nil, err
		}

		return &ast.UnaryOp{
			Base:     ast.Base{Line: opToken.Line, Column: opToken.Column},
			Operator: opToken.Value.(string),
			Operand:  operand,
		}, nil
	}

	return p.parsePostfix()
}

// parsePostfix parses postfix operations (function calls, member access)
func (p *Parser) parsePostfix() (ast.Node, error) {
	expr, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}

	for {
		// Function call
		if p.match(lexer.LPAREN) {
			callLine, callCol := expr.Pos().Line, expr.Pos().Column
			p.advance()
			var arguments []ast.Node
			var holeNames []string // gensym names assigned to holes in this call, in order

			for !p.match(lexer.RPAREN) {
				// Partial application hole: f(a, ?, b) → fn(<gensym>) = f(a, <gensym>, b)
				// Generated names start with \x00, which the lexer cannot produce from source,
				// so they are structurally guaranteed to never conflict with user identifiers.
				if p.match(lexer.QUESTION) {
					holeToken := p.advance()
					paramName := fmt.Sprintf("\x00h%d", p.gensym)
					p.gensym++
					holeNames = append(holeNames, paramName)
					arguments = append(arguments, &ast.Identifier{
						Base: ast.Base{Line: holeToken.Line, Column: holeToken.Column},
						Name: paramName,
					})
				} else {
					arg, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					arguments = append(arguments, arg)
				}

				if p.match(lexer.COMMA) {
					p.advance()
				}
			}

			if _, err := p.expect(lexer.RPAREN); err != nil {
				return nil, err
			}

			callNode := &ast.FunctionCall{
				Base:      ast.Base{Line: callLine, Column: callCol},
				Function:  expr,
				Arguments: arguments,
			}

			if len(holeNames) > 0 {
				// Desugar into a lambda: f(a, ?, b) → fn(\x00h42) = f(a, \x00h42, b)
				expr = &ast.Lambda{
					Base:   ast.Base{Line: callLine, Column: callCol},
					Params: holeNames,
					Body:   callNode,
				}
			} else {
				expr = callNode
			}

		} else if p.match(lexer.DOT) {
			// Member access
			p.advance()
			memberToken, err := p.expect(lexer.IDENTIFIER)
			if err != nil {
				return nil, err
			}

			expr = &ast.MemberAccess{
				Base:   ast.Base{Line: expr.Pos().Line, Column: expr.Pos().Column},
				Object: expr,
				Member: memberToken.Value.(string),
			}

		} else {
			break
		}
	}

	return expr, nil
}

// parsePrimary parses primary expressions (literals, identifiers, parenthesized)
func (p *Parser) parsePrimary() (ast.Node, error) {
	token := p.current()

	// Number literal
	if p.match(lexer.NUMBER) {
		numToken := p.advance()
		return &ast.NumberLiteral{
			Base:  ast.Base{Line: numToken.Line, Column: numToken.Column},
			Value: numToken.Value,
		}, nil
	}

	// String literal
	if p.match(lexer.STRING) {
		strToken := p.advance()
		return &ast.StringLiteral{
			Base:  ast.Base{Line: strToken.Line, Column: strToken.Column},
			Value: strToken.Value.(string),
		}, nil
	}

	// Boolean literal
	if p.match(lexer.BOOLEAN) {
		boolToken := p.advance()
		return &ast.BooleanLiteral{
			Base:  ast.Base{Line: boolToken.Line, Column: boolToken.Column},
			Value: boolToken.Value.(bool),
		}, nil
	}

	// List literal
	if p.match(lexer.LBRACKET) {
		return p.parseListLiteral()
	}

	// Record literal
	if p.match(lexer.LBRACE) {
		return p.parseRecordLiteral()
	}

	// Identifier
	if p.match(lexer.IDENTIFIER) {
		idToken := p.advance()
		return &ast.Identifier{
			Base: ast.Base{Line: idToken.Line, Column: idToken.Column},
			Name: idToken.Value.(string),
		}, nil
	}

	// Parenthesized expression or tuple
	if p.match(lexer.LPAREN) {
		p.advance()

		// Empty tuple
		if p.match(lexer.RPAREN) {
			p.advance()
			return &ast.TupleLiteral{
				Base:     ast.Base{Line: token.Line, Column: token.Column},
				Elements: []ast.Node{},
			}, nil
		}

		first, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		// Tuple
		if p.match(lexer.COMMA) {
			elements := []ast.Node{first}
			for p.match(lexer.COMMA) {
				p.advance()
				if !p.match(lexer.RPAREN) {
					elem, err := p.parseExpression()
					if err != nil {
						return nil, err
					}
					elements = append(elements, elem)
				}
			}

			if _, err := p.expect(lexer.RPAREN); err != nil {
				return nil, err
			}

			return &ast.TupleLiteral{
				Base:     ast.Base{Line: token.Line, Column: token.Column},
				Elements: elements,
			}, nil
		}

		// Parenthesized expression
		if _, err := p.expect(lexer.RPAREN); err != nil {
			return nil, err
		}
		return first, nil
	}

	// Introspection expressions: @which name | @inspect layer | @diff layerA layerB
	if p.match(lexer.AT) {
		next := p.peek(1)
		if next.Type == lexer.IDENTIFIER {
			op := next.Value.(string)
			if op == "which" || op == "inspect" || op == "diff" {
				return p.parseIntrospectionExpr()
			}
		}
	}

	return nil, p.error(fmt.Sprintf("unexpected token: %s", token.Type))
}

// parseIntrospectionExpr parses @which name | @inspect layer | @diff a b
func (p *Parser) parseIntrospectionExpr() (*ast.IntrospectionExpr, error) {
	atToken, err := p.expect(lexer.AT)
	if err != nil {
		return nil, err
	}

	opToken, err := p.expect(lexer.IDENTIFIER)
	if err != nil {
		return nil, err
	}
	op := opToken.Value.(string)

	var args []string
	switch op {
	case "which", "inspect":
		// single identifier argument
		argToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		args = []string{argToken.Value.(string)}
	case "diff":
		// two identifier arguments
		aToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		bToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}
		args = []string{aToken.Value.(string), bToken.Value.(string)}
	default:
		return nil, p.error(fmt.Sprintf("unknown introspection op '%s': expected which, inspect, or diff", op))
	}

	return &ast.IntrospectionExpr{
		Base: ast.Base{Line: atToken.Line, Column: atToken.Column},
		Op:   op,
		Args: args,
	}, nil
}

// parseListLiteral parses list literal: [1, 2, 3]
func (p *Parser) parseListLiteral() (*ast.ListLiteral, error) {
	start, err := p.expect(lexer.LBRACKET)
	if err != nil {
		return nil, err
	}

	var elements []ast.Node
	for !p.match(lexer.RBRACKET) {
		elem, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		elements = append(elements, elem)

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RBRACKET); err != nil {
		return nil, err
	}

	return &ast.ListLiteral{
		Base:     ast.Base{Line: start.Line, Column: start.Column},
		Elements: elements,
	}, nil
}

// parseRecordLiteral parses record literal: {name: "Layer", version: 1}
func (p *Parser) parseRecordLiteral() (*ast.RecordLiteral, error) {
	start, err := p.expect(lexer.LBRACE)
	if err != nil {
		return nil, err
	}

	fields := make(map[string]ast.Node)
	for !p.match(lexer.RBRACE) {
		keyToken, err := p.expect(lexer.IDENTIFIER)
		if err != nil {
			return nil, err
		}

		if _, err := p.expect(lexer.COLON); err != nil {
			return nil, err
		}

		value, err := p.parseExpression()
		if err != nil {
			return nil, err
		}

		fields[keyToken.Value.(string)] = value

		if p.match(lexer.COMMA) {
			p.advance()
		}
	}

	if _, err := p.expect(lexer.RBRACE); err != nil {
		return nil, err
	}

	return &ast.RecordLiteral{
		Base:   ast.Base{Line: start.Line, Column: start.Column},
		Fields: fields,
	}, nil
}

// getPrecedence returns operator precedence
func (p *Parser) getPrecedence(tokenType lexer.TokenType) int {
	precedenceTable := map[lexer.TokenType]int{
		lexer.OR:      1,
		lexer.AND:     2,
		lexer.EQ:      3,
		lexer.NEQ:     3,
		lexer.LT:      4,
		lexer.GT:      4,
		lexer.LTE:     4,
		lexer.GTE:     4,
		lexer.PLUS:    5,
		lexer.MINUS:   5,
		lexer.STAR:    6,
		lexer.SLASH:   6,
		lexer.PERCENT: 6,
		lexer.CARET:   7,
	}

	if prec, ok := precedenceTable[tokenType]; ok {
		return prec
	}
	return 0
}

// Parse is a convenience function to parse tokens
func Parse(tokens []lexer.Token) (*ast.Program, error) {
	parser := NewParser(tokens)
	return parser.Parse()
}
