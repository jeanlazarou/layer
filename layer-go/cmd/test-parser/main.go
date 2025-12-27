package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// extractEnabledLayers extracts the list of enabled layers from @layers directive
func extractEnabledLayers(source string) []string {
	re := regexp.MustCompile(`@layers\s*\[([^\]]+)\]`)
	matches := re.FindStringSubmatch(source)

	if len(matches) < 2 {
		return []string{}
	}

	layerNames := strings.Split(matches[1], ",")
	result := make([]string, 0, len(layerNames))
	for _, name := range layerNames {
		trimmed := strings.TrimSpace(name)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// extractLayerCode extracts code for each @layer block
func extractLayerCode(source string) map[string]string {
	layers := make(map[string]string)

	// Pattern to find @layer declarations
	re := regexp.MustCompile(`@layer\s+(\w+)\s*\{`)
	matches := re.FindAllStringSubmatchIndex(source, -1)

	for _, match := range matches {
		// match[0], match[1] = full match start/end
		// match[2], match[3] = layer name start/end
		layerName := source[match[2]:match[3]]
		startPos := match[1] // position after the opening {

		// Find matching closing brace by counting
		braceCount := 1
		endPos := startPos

		for braceCount > 0 && endPos < len(source) {
			if source[endPos] == '{' {
				braceCount++
			} else if source[endPos] == '}' {
				braceCount--
			}
			endPos++
		}

		if braceCount == 0 {
			// Extract layer code (excluding the closing brace)
			layerCode := source[startPos : endPos-1]
			layers[layerName] = dedentCode(layerCode)
		}
	}

	return layers
}

// extractMainExpression extracts the main expression (code after all @layer blocks)
func extractMainExpression(source string) string {
	// Find the last @layer block
	re := regexp.MustCompile(`@layer\s+\w+\s*\{`)
	matches := re.FindAllStringIndex(source, -1)

	if len(matches) == 0 {
		// No layer blocks, check for @layers directive
		layersRe := regexp.MustCompile(`@layers\s*\[[^\]]+\]`)
		if loc := layersRe.FindStringIndex(source); loc != nil {
			return strings.TrimSpace(source[loc[1]:])
		}
		return strings.TrimSpace(source)
	}

	// Find the end of the last layer block
	lastMatch := matches[len(matches)-1]
	startPos := lastMatch[1] // position after the opening {

	// Find matching closing brace
	braceCount := 1
	endPos := startPos

	for braceCount > 0 && endPos < len(source) {
		if source[endPos] == '{' {
			braceCount++
		} else if source[endPos] == '}' {
			braceCount--
		}
		endPos++
	}

	if endPos < len(source) {
		return strings.TrimSpace(source[endPos:])
	}

	return ""
}

// dedentCode removes common indentation from code
func dedentCode(code string) string {
	lines := strings.Split(code, "\n")

	// Find minimum indentation (ignoring empty lines)
	minIndent := -1
	for _, line := range lines {
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "" {
			continue // Skip empty lines
		}
		indent := len(line) - len(trimmed)
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return strings.TrimSpace(code)
	}

	// Remove the common indentation
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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-parser <file.layer>")
		os.Exit(1)
	}

	content, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		os.Exit(1)
	}

	source := string(content)

	// Test extraction functions
	enabledLayers := extractEnabledLayers(source)
	layerCode := extractLayerCode(source)
	mainExpr := extractMainExpression(source)

	fmt.Println("=== Parse Result ===")
	fmt.Printf("Enabled Layers: %v\n", enabledLayers)
	fmt.Printf("Number of layers: %d\n", len(layerCode))
	fmt.Println("\n=== Layer Code ===")
	for name, code := range layerCode {
		fmt.Printf("\nLayer '%s' (%d chars):\n%s\n", name, len(code), code)
	}
	fmt.Println("\n=== Main Expression ===")
	fmt.Println(mainExpr)

	// Verify extraction worked
	if len(layerCode) > 0 {
		fmt.Println("\n✓ SUCCESS: Layer code extracted correctly!")
		for name := range layerCode {
			if len(layerCode[name]) == 0 {
				fmt.Printf("✗ WARNING: Layer '%s' has empty code\n", name)
			}
		}
	} else {
		fmt.Println("\n✗ ERROR: No layer code extracted!")
		os.Exit(1)
	}
}
