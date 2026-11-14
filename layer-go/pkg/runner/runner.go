// Package runner runs Layer source end to end (lex, parse, load, check,
// execute) and reports the outcome in a structured form shared by the CLI,
// the WebAssembly bridge used by the visual editor, and the tests.
package runner

import (
	"encoding/json"
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
	// Kind is "app" when the main expression returns app({...}) (Display is
	// then the current view), "ui" for a UI component tree, "none" when the
	// program has no main expression, and "value" otherwise.
	Kind string `json:"kind"`
	// Model is the app's current model as JSON (Kind "app" only). Pass it
	// back in Options.Model to continue from this state.
	Model json.RawMessage `json:"model,omitempty"`
	// ModelText is the model in Layer syntax, for display.
	ModelText string   `json:"modelText,omitempty"`
	Logs      []string `json:"logs"`
	Errors    []Issue  `json:"errors"`
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

	// The options below apply when the main expression returns an app.
	//
	// Model (JSON, as returned in Result.Model) resumes from that model
	// instead of the app's init. Messages (JSON each) are then applied in
	// order, followed by Send: Layer expressions (e.g. `{step: 2}`)
	// evaluated in the program. The result is the final model and view.
	Model    string
	Messages []string
	Send     []string
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

	if app, ok := value.(*interpreter.App); ok {
		runApp(interp, app, opts, &res)
		return res
	}
	if opts.Model != "" || len(opts.Messages) > 0 || len(opts.Send) > 0 {
		res.Errors = []Issue{{Kind: "runtime", Message: "messages were given, but the main expression does not return an app (see app({init, update, view}))"}}
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

// runApp drives an app: start from init (or opts.Model), apply each message
// with update, then render view(model).
func runApp(interp *interpreter.Interpreter, app *interpreter.App, opts Options, res *Result) {
	fail := func(issues ...Issue) {
		res.Success = false
		res.Errors = append(res.Errors, issues...)
	}

	model := app.Init
	if opts.Model != "" {
		decoded, err := interpreter.DecodeData(opts.Model)
		if err != nil {
			fail(Issue{Kind: "internal", Message: "invalid model: " + err.Error()})
			return
		}
		model = decoded
	}

	var messages []interface{}
	for _, raw := range opts.Messages {
		msg, err := interpreter.DecodeData(raw)
		if err != nil {
			fail(Issue{Kind: "internal", Message: "invalid message: " + err.Error()})
			return
		}
		messages = append(messages, msg)
	}
	for _, src := range opts.Send {
		msg, err := evalExpression(interp, src)
		if err != nil {
			fail(Issue{Kind: "runtime", Message: fmt.Sprintf("in message %q: %v", src, err)})
			return
		}
		if err := interpreter.CheckData(msg, "message"); err != nil {
			fail(Issue{Kind: "runtime", Message: fmt.Sprintf("message %q must be plain data: %v", src, err)})
			return
		}
		messages = append(messages, msg)
	}

	for _, msg := range messages {
		next, err := interpreter.CallValue(app.Update, msg, model)
		if err != nil {
			fail(Issues(err)...)
			return
		}
		if err := interpreter.CheckData(next, "model"); err != nil {
			fail(Issue{Kind: "runtime", Message: fmt.Sprintf("update returned a model that is not plain data after message %s: %v", interpreter.ShowNested(msg), err)})
			return
		}
		model = next
	}

	view, err := interpreter.CallValue(app.View, model)
	if err != nil {
		fail(Issues(err)...)
		return
	}
	xml, err := renderView(view)
	if err != nil {
		fail(Issue{Kind: "runtime", Message: err.Error()})
		return
	}
	encoded, err := interpreter.EncodeData(model)
	if err != nil {
		fail(Issue{Kind: "runtime", Message: "the model must be plain data: " + err.Error()})
		return
	}

	res.Success = true
	res.Value = model
	res.Kind = "app"
	res.Display = xml
	res.Model = json.RawMessage(encoded)
	res.ModelText = interpreter.Show(model)
}

// renderView turns what an app's view returned into XML: a component, a
// list of components (e.g. from blend: {page: collect}) shown as a column,
// or XML already produced by render().
func renderView(view interface{}) (string, error) {
	switch v := view.(type) {
	case *interpreter.UIComponent:
		return v.Render(0), nil
	case []interface{}:
		for _, item := range v {
			if _, ok := item.(*interpreter.UIComponent); !ok {
				return "", fmt.Errorf("an app's view returned a list, so every item must be a UI component, but one is %s", interpreter.ShowNested(item))
			}
		}
		column := &interpreter.UIComponent{Type: "container", Props: map[string]interface{}{"layout": "vertical"}, Children: v}
		return column.Render(0), nil
	case string:
		if renderedUI.MatchString(v) {
			return v, nil
		}
	}
	return "", fmt.Errorf("an app's view must return a UI component (or a list of them), got %s", interpreter.ShowNested(view))
}

// evalExpression evaluates a standalone Layer expression in the program's
// top-level scope (used for --send messages).
func evalExpression(interp *interpreter.Interpreter, src string) (interface{}, error) {
	tokens, err := lexer.Tokenize(src)
	if err != nil {
		return nil, err
	}
	program, err := parser.Parse(tokens)
	if err != nil {
		return nil, err
	}
	if program.MainExpression == nil || program.LayersDirective != nil || len(program.LayerDeclarations) > 0 {
		return nil, fmt.Errorf("expected a single expression")
	}
	return interp.Eval(program.MainExpression, nil)
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
