package sourcefile

import (
	"strings"
	"testing"
)

func TestSplitEnabledLayers(t *testing.T) {
	source := `@layers [core, debug, ui]

@layer core {
	fn test() = 42
}

test()
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	expected := []string{"core", "debug", "ui"}
	if len(file.EnabledLayers) != len(expected) {
		t.Fatalf("Expected %d layers, got %d", len(expected), len(file.EnabledLayers))
	}
	for i, layer := range expected {
		if file.EnabledLayers[i] != layer {
			t.Errorf("Expected layer %s at position %d, got %s", layer, i, file.EnabledLayers[i])
		}
	}
}

func TestSplitLayerCode(t *testing.T) {
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
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	coreCode, hasCoreCode := file.Layers["core"]
	if !hasCoreCode || len(coreCode) < 10 {
		t.Errorf("Expected substantial core layer code, got: %q", coreCode)
	}
	if !strings.Contains(coreCode, "fn factorial(n)") {
		t.Errorf("Core layer code missing function definition: %q", coreCode)
	}

	debugCode, hasDebugCode := file.Layers["debug"]
	if !hasDebugCode || debugCode == "" {
		t.Errorf("Expected debug layer code, got: %q", debugCode)
	}
}

func TestSplitMainExpression(t *testing.T) {
	source := `@layers [core]

@layer core {
	fn test() = 42
}

test() + 10
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if file.MainExpression != "test() + 10" {
		t.Errorf("Expected 'test() + 10', got %q", file.MainExpression)
	}
}

func TestSplitBracesInStringsAndComments(t *testing.T) {
	source := `@layers [core]

@layer core {
	# a comment with a stray brace } should not close the block
	fn describe() = "a string with braces { } inside"
}

describe()
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	coreCode := file.Layers["core"]
	if !strings.Contains(coreCode, `"a string with braces { } inside"`) {
		t.Errorf("Layer body truncated by brace inside string/comment: %q", coreCode)
	}
	if file.MainExpression != "describe()" {
		t.Errorf("Expected 'describe()', got %q", file.MainExpression)
	}
}

func TestSplitNestedBraces(t *testing.T) {
	source := `@layers [core]

@layer core {
	fn person() = { name: "Layer", tags: { a: 1 } }
}

person()
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if !strings.Contains(file.Layers["core"], `tags: { a: 1 }`) {
		t.Errorf("Nested record braces broke extraction: %q", file.Layers["core"])
	}
	if file.MainExpression != "person()" {
		t.Errorf("Expected 'person()', got %q", file.MainExpression)
	}
}

func TestSplitConditionalLayersAndBlendMode(t *testing.T) {
	source := `@layers [core, debug if env == "development"] blend: chain

@layer core {
	fn test() = 42
}

@layer debug {
	fn test() = core.test()
}

test()
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	expected := []string{"core", "debug"}
	if len(file.EnabledLayers) != len(expected) {
		t.Fatalf("Expected layers %v, got %v", expected, file.EnabledLayers)
	}
	for i, layer := range expected {
		if file.EnabledLayers[i] != layer {
			t.Errorf("Expected layer %s at position %d, got %s", layer, i, file.EnabledLayers[i])
		}
	}
	if file.MainExpression != "test()" {
		t.Errorf("Expected 'test()', got %q", file.MainExpression)
	}
}

func TestSplitNoLayerBlocks(t *testing.T) {
	source := `@layers [core] blend: override

1 + 2
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if file.MainExpression != "1 + 2" {
		t.Errorf("Expected '1 + 2', got %q", file.MainExpression)
	}
}

func TestSplitUnicodeSource(t *testing.T) {
	source := `@layers [core]

@layer core {
	fn double(x) = λ(y) -> y * 2
}

double(1)
`
	file, err := Split(source)
	if err != nil {
		t.Fatalf("Split failed: %v", err)
	}

	if !strings.Contains(file.Layers["core"], "λ(y) -> y * 2") {
		t.Errorf("Unicode lambda broke position slicing: %q", file.Layers["core"])
	}
}

func TestSplitInvalidSourceReturnsError(t *testing.T) {
	if _, err := Split("@layer core {"); err == nil {
		t.Error("Expected an error for unterminated layer block")
	}
}

func TestDedent(t *testing.T) {
	code := `    fn factorial(n) =
        match n with
        | 0 -> 1
        | n -> n * factorial(n - 1)`

	expected := `fn factorial(n) =
    match n with
    | 0 -> 1
    | n -> n * factorial(n - 1)`

	if got := dedent(code); got != expected {
		t.Errorf("Dedent failed.\nExpected:\n%s\nGot:\n%s", expected, got)
	}
}

func TestSplitPurityBlendAndInactiveLayers(t *testing.T) {
	file, err := Split(`@layers [core, cache] blend: fallback

@layer<impure> cache {
    fn fetch(k) = k
}

@layer<pure> core {
    fn id2(x) = x
}

@layer hidden {
    fn h() = 1
}

fetch("a")`)
	if err != nil {
		t.Fatal(err)
	}
	if file.BlendMode != "fallback" {
		t.Errorf("blend mode %q", file.BlendMode)
	}
	if file.Layers["cache"] != "fn fetch(k) = k" {
		t.Errorf("impure layer body not extracted: %q", file.Layers["cache"])
	}
	if file.Purity["cache"] != "impure" || file.Purity["core"] != "pure" {
		t.Errorf("purity %v", file.Purity)
	}
	if strings.Join(file.DeclaredLayers, ",") != "cache,core,hidden" {
		t.Errorf("declared layers %v", file.DeclaredLayers)
	}
	if strings.Join(file.EnabledLayers, ",") != "core,cache" {
		t.Errorf("enabled layers %v", file.EnabledLayers)
	}
	if file.MainExpression != `fetch("a")` {
		t.Errorf("main %q", file.MainExpression)
	}
}

func TestSplitBlendMap(t *testing.T) {
	file, err := Split(`@layers [core, cache] blend: {_: chain, fetch: fallback}
@layer core {
    fn fetch(k) = k
}
fetch(1)`)
	if err != nil {
		t.Fatal(err)
	}
	if file.BlendMode != "chain" || file.BlendModes["fetch"] != "fallback" || len(file.BlendModes) != 1 {
		t.Errorf("blend %q %v", file.BlendMode, file.BlendModes)
	}
	if file.MainExpression != "fetch(1)" {
		t.Errorf("main %q", file.MainExpression)
	}

	// No @layer blocks: the blend map must not leak into the main expression.
	file, err = Split("@layers [core] blend: {_: collect}\n1 + 1")
	if err != nil {
		t.Fatal(err)
	}
	if file.MainExpression != "1 + 1" {
		t.Errorf("main %q", file.MainExpression)
	}
}
