// Package runner runs Layer source end to end (lex, parse, load, check,
// execute) and reports the outcome in a structured form shared by the CLI,
// the WebAssembly bridge used by the visual editor, and the tests.
package runner

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jeanlazarou/layer/pkg/ast"
	"github.com/jeanlazarou/layer/pkg/interpreter"
	"github.com/jeanlazarou/layer/pkg/lexer"
	"github.com/jeanlazarou/layer/pkg/parser"
)

// Issue is one error with its location in the source (Line/Column are 0
// when unknown).
type Issue struct {
	Kind    string `json:"kind"` // lexer | parser | check | runtime | internal
	Line    int    `json:"line"`
	Column  int    `json:"column"`
	Message string `json:"message"`
	Trace   string `json:"trace,omitempty"` // Layer call trace for runtime errors
}

func (is Issue) String() string {
	loc := ""
	if is.Line > 0 {
		loc = fmt.Sprintf("%d:%d: ", is.Line, is.Column)
	}
	s := fmt.Sprintf("%s%s error: %s", loc, is.Kind, is.Message)
	if is.Trace != "" {
		s += "\n    in " + is.Trace
	}
	return s
}

// Result is the outcome of running a program.
type Result struct {
	Success bool        `json:"success"`
	Value   interface{} `json:"-"`
	// Display is the human-readable result (XML for UI components).
	Display string `json:"result"`
	// Kind is "ui" when the result is a UI component tree, "none" when the
	// program has no main expression, and "value" otherwise.
	Kind   string   `json:"kind"`
	Logs   []string `json:"logs"`
	Errors []Issue  `json:"errors"`
}

// Options configures a run.
type Options struct {
	// Vars are seeded into the program's top-level scope before it runs
	// (conditions in @layers, current values of UI inputs).
	Vars map[string]interface{}
	// LogSink, when non-nil, receives log lines as they are produced.
	// Lines are always collected in Result.Logs as well.
	LogSink func(line string)
	// CheckOnly loads and checks the program without running the main
	// expression.
	CheckOnly bool
}

// Parse lexes and parses source, converting failures into issues.
func Parse(source string) (*ast.Program, []Issue) {
	tokens, err := lexer.Tokenize(source)
	if err != nil {
		return nil, Issues(err)
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return nil, Issues(err)
	}
	return program, nil
}

// Run executes source and never panics: interpreter panics are reported as
// "internal" issues (important under WebAssembly, where a panic would kill
// the runtime for every later call).
func Run(source string, opts Options) (res Result) {
	res = Result{Logs: []string{}, Errors: []Issue{}, Kind: "none"}
	defer func() {
		if r := recover(); r != nil {
			res.Success = false
			res.Errors = append(res.Errors, Issue{Kind: "internal", Message: fmt.Sprintf("interpreter panic: %v", r)})
		}
	}()

	program, issues := Parse(source)
	if issues != nil {
		res.Errors = issues
		return res
	}

	interp := interpreter.NewInterpreter()
	interp.LogSink = opts.LogSink
	for k, v := range opts.Vars {
		interp.SetVar(k, v)
	}
	defer func() { res.Logs = append(res.Logs, interp.Logs()...) }()

	if opts.CheckOnly {
		if err := interp.Load(program); err != nil {
			res.Errors = Issues(err)
			return res
		}
		res.Success = true
		return res
	}

	value, err := interp.Execute(program)
	if err != nil {
		res.Errors = Issues(err)
		return res
	}

	res.Success = true
	res.Value = value
	if program.MainExpression != nil {
		res.Display = interpreter.Show(value)
		res.Kind = "value"
		if _, ok := value.(*interpreter.UIComponent); ok {
			res.Kind = "ui"
		} else if s, ok := value.(string); ok && renderedUI.MatchString(s) {
			res.Kind = "ui" // render(component) returns the XML as a string
		}
	}
	return res
}

// DescribeResult is the outcome of Describe.
type DescribeResult struct {
	Success     bool                     `json:"success"`
	Description *interpreter.Description `json:"description,omitempty"`
	Errors      []Issue                  `json:"errors"`
}

// Describe loads source (without running the main expression) and reports
// its layer stack and how each user-defined name resolves.
func Describe(source string) (res DescribeResult) {
	res = DescribeResult{Errors: []Issue{}}
	defer func() {
		if r := recover(); r != nil {
			res.Success = false
			res.Errors = append(res.Errors, Issue{Kind: "internal", Message: fmt.Sprintf("interpreter panic: %v", r)})
		}
	}()

	program, issues := Parse(source)
	if issues != nil {
		res.Errors = issues
		return res
	}
	interp := interpreter.NewInterpreter()
	interp.LogSink = nil
	// Check errors don't prevent describing the stack: definitions are
	// already registered when the checks run.
	if err := interp.Load(program); err != nil {
		var cerr *interpreter.CheckError
		if !errors.As(err, &cerr) {
			res.Errors = Issues(err)
			return res
		}
		res.Errors = Issues(err)
	}
	d := interp.Describe(program)
	res.Description = &d
	res.Success = true
	return res
}

// renderedUI matches the XML produced by the ui layer's render().
var renderedUI = regexp.MustCompile(`^<(view|text|button|input|container)[\s/>]`)

var locatedError = regexp.MustCompile(`^(lexer|parser) error at (\d+):(\d+): (.*)$`)

// Issues converts an error from any stage into located issues.
func Issues(err error) []Issue {
	var cerr *interpreter.CheckError
	if errors.As(err, &cerr) {
		out := make([]Issue, len(cerr.Issues))
		for i, is := range cerr.Issues {
			out[i] = Issue{Kind: "check", Line: is.Line, Column: is.Column, Message: is.Message}
		}
		return out
	}

	var rerr *interpreter.RuntimeError
	if errors.As(err, &rerr) {
		msg := rerr.Message
		// Keep any context added around the runtime error (e.g. "error
		// evaluating condition for layer 'debug'").
		if full := err.Error(); !strings.HasPrefix(full, "runtime error at") {
			if idx := strings.Index(full, ": runtime error at"); idx > 0 {
				msg = full[:idx] + ": " + msg
			}
		}
		return []Issue{{Kind: "runtime", Line: rerr.Line, Column: rerr.Column, Message: msg, Trace: rerr.TraceString()}}
	}

	if m := locatedError.FindStringSubmatch(err.Error()); m != nil {
		line, _ := strconv.Atoi(m[2])
		col, _ := strconv.Atoi(m[3])
		return []Issue{{Kind: m[1], Line: line, Column: col, Message: m[4]}}
	}

	return []Issue{{Kind: "runtime", Message: err.Error()}}
}
