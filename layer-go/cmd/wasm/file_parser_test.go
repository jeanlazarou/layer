package main

import (
	"testing"
)

func TestExtractEnabledLayers(t *testing.T) {
	source := `@layers [core, debug, ui]

@layer core {
	fn test() = 42
}
`
	layers := extractEnabledLayers(source)

	expected := []string{"core", "debug", "ui"}
	if len(layers) != len(expected) {
		t.Errorf("Expected %d layers, got %d", len(expected), len(layers))
	}

	for i, layer := range expected {
		if i >= len(layers) || layers[i] != layer {
			t.Errorf("Expected layer %s at position %d, got %s", layer, i, layers[i])
		}
	}
}

func TestExtractLayerCode(t *testing.T) {
	source := `@layers [core, debug]

@layer core {
	fn factorial(n) =
		match n with
		| 0 -> 1
		| n -> n * factorial(n - 1)
}

@layer debug {
	fn factorial(n) =
		let _ = log("Computing factorial of {n}") in
			core.factorial(n)
}

factorial(5)
`

	layerCode := extractLayerCode(source)

	// Check core layer
	coreCode, hasCoreCode := layerCode["core"]
	if !hasCoreCode {
		t.Error("Expected core layer code to be extracted")
	}
	if coreCode == "" {
		t.Error("Expected core layer code to be non-empty")
	}
	if len(coreCode) < 10 {
		t.Errorf("Expected core layer code to be substantial, got: %q", coreCode)
	}

	// Check debug layer
	debugCode, hasDebugCode := layerCode["debug"]
	if !hasDebugCode {
		t.Error("Expected debug layer code to be extracted")
	}
	if debugCode == "" {
		t.Error("Expected debug layer code to be non-empty")
	}

	t.Logf("Core layer code:\n%s", coreCode)
	t.Logf("Debug layer code:\n%s", debugCode)
}

func TestExtractMainExpression(t *testing.T) {
	source := `@layers [core]

@layer core {
	fn test() = 42
}

test() + 10
`

	mainExpr := extractMainExpression(source)

	if mainExpr == "" {
		t.Error("Expected main expression to be extracted")
	}

	if mainExpr != "test() + 10" {
		t.Errorf("Expected 'test() + 10', got %q", mainExpr)
	}

	t.Logf("Main expression: %s", mainExpr)
}

func TestDedentCode(t *testing.T) {
	code := `    fn factorial(n) =
        match n with
        | 0 -> 1
        | n -> n * factorial(n - 1)`

	dedented := dedentCode(code)

	expected := `fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)`

	if dedented != expected {
		t.Errorf("Dedent failed.\nExpected:\n%s\nGot:\n%s", expected, dedented)
	}
}
