package interpreter

import (
	"testing"
)

func TestRecordPatternBasic(t *testing.T) {
	// Bind fields to names
	source := `
	@layers [core]
	@layer core {
		fn greet(person) =
			match person with
			| {name: n} -> str_concat("Hello, ", n)
	}
	greet({name: "Alice", age: 30})
	`
	result := mustExecute(t, source)
	if result != "Hello, Alice" {
		t.Errorf("expected \"Hello, Alice\", got %v", result)
	}
}

func TestRecordPatternMultipleFields(t *testing.T) {
	// Destructure multiple fields in one pattern
	source := `
	@layers [core]
	@layer core {
		fn describe(p) =
			match p with
			| {name: n, age: a} -> str_concat(n, str_concat(" is ", str_concat(to_string(a), " years old")))
	}
	describe({name: "Bob", age: 25})
	`
	result := mustExecute(t, source)
	if result != "Bob is 25 years old" {
		t.Errorf("expected \"Bob is 25 years old\", got %v", result)
	}
}

func TestRecordPatternPartial(t *testing.T) {
	// Pattern only mentions some fields — extra fields in the value are ignored
	source := `
	match {x: 1, y: 2, z: 3} with
	| {x: a} -> a
	`
	result := mustExecute(t, source)
	if result != int64(1) {
		t.Errorf("expected 1, got %v", result)
	}
}

func TestRecordPatternMissingFieldNoMatch(t *testing.T) {
	// Pattern requires a field not present in the value → falls through to next case
	source := `
	match {x: 1} with
	| {y: v} -> "has y"
	| _      -> "no y"
	`
	result := mustExecute(t, source)
	if result != "no y" {
		t.Errorf("expected \"no y\", got %v", result)
	}
}

func TestRecordPatternWithGuard(t *testing.T) {
	// Record pattern + guard clause
	source := `
	@layers [core]
	@layer core {
		fn classify(person) =
			match person with
			| {age: a} when a < 18 -> "minor"
			| {age: a} when a < 65 -> "adult"
			| {age: a}             -> "senior"
	}
	classify({name: "Carol", age: 70})
	`
	result := mustExecute(t, source)
	if result != "senior" {
		t.Errorf("expected \"senior\", got %v", result)
	}
}

func TestRecordPatternNestedLiteral(t *testing.T) {
	// Field value matched against a literal pattern
	source := `
	match {status: "active", id: 42} with
	| {status: "active", id: n} -> n
	| {status: _}               -> 0
	`
	result := mustExecute(t, source)
	if result != int64(42) {
		t.Errorf("expected 42, got %v", result)
	}
}

func TestRecordPatternNotARecord(t *testing.T) {
	// Record pattern does not match a non-record value → falls through
	source := `
	match 42 with
	| {x: v} -> "record"
	| _      -> "not a record"
	`
	result := mustExecute(t, source)
	if result != "not a record" {
		t.Errorf("expected \"not a record\", got %v", result)
	}
}
