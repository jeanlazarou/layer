package interpreter

import (
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/jeanlazarou/layer/pkg/ast"
	"github.com/jeanlazarou/layer/pkg/layers"
)

// DefaultMaxCallDepth bounds Layer-level recursion, so runaway recursion
// reports a normal runtime error instead of exhausting the stack — which,
// under WebAssembly, would kill the whole runtime and leave the editor
// unusable until reload. It is 100000 natively and 20000 under WebAssembly
// (see limits_js.go).
const DefaultMaxCallDepth = defaultMaxCallDepth

// Closure represents a function value.
type Closure struct {
	// Name is the function name for `fn` definitions ("" for lambdas). It
	// is used in error traces and messages.
	Name        string
	Params      []string
	Body        ast.Node
	Env         map[string]interface{}
	Interpreter *Interpreter
	// Layer is the layer the closure was defined in ("" for top-level
	// code). Used to enforce purity while the closure body executes.
	Layer string
	// Native, when set, implements the function in Go. Built-ins that
	// return functions (compose, const, trace, benchmark) use it so their
	// results are ordinary closures with a known arity.
	Native func(args ...interface{}) (interface{}, error)
}

// qualifiedName returns "layer.name" for named closures.
func (c *Closure) qualifiedName() string {
	if c.Layer == "" {
		return c.Name
	}
	return c.Layer + "." + c.Name
}

func (c *Closure) describe() string {
	if c.Name != "" {
		return c.qualifiedName()
	}
	return "function"
}

// Call invokes the closure with arguments
func (c *Closure) Call(args ...interface{}) (interface{}, error) {
	if len(args) != len(c.Params) {
		return nil, fmt.Errorf("%s expects %d argument%s, got %d", c.describe(), len(c.Params), plural(len(c.Params)), len(args))
	}

	if c.Native != nil {
		return c.Native(args...)
	}

	interp := c.Interpreter
	interp.depth++
	defer func() { interp.depth-- }()
	if interp.depth > interp.MaxCallDepth {
		return nil, fmt.Errorf("maximum call depth (%d) exceeded in %s — runaway recursion?", interp.MaxCallDepth, c.describe())
	}

	// Create new environment with parameters bound
	newEnv := make(map[string]interface{}, len(c.Env)+len(c.Params))
	for k, v := range c.Env {
		newEnv[k] = v
	}
	for i, param := range c.Params {
		newEnv[param] = args[i]
	}

	// Evaluate body in new environment, tracking which layer is executing
	// so purity rules can be enforced inside the body.
	if c.Layer != "" {
		interp.pushExecLayer(c.Layer)
		defer interp.popExecLayer()
	}
	var result interface{}
	var err error
	if stackHopEvery > 0 && interp.depth%stackHopEvery == 0 {
		result, err = interp.evalOnFreshGoroutine(c.Body, newEnv)
	} else {
		result, err = interp.Eval(c.Body, newEnv)
	}
	if err != nil && c.Name != "" {
		var rerr *RuntimeError
		if errors.As(err, &rerr) {
			rerr.addFrame(c.qualifiedName())
		}
	}
	return result, err
}

// evalOnFreshGoroutine evaluates node on a new goroutine and waits for it.
// Under WebAssembly, parking the current goroutine unwinds the browser's
// native call stack, so the new goroutine starts with an empty one. The
// interpreter is not used concurrently: the caller is blocked meanwhile.
func (i *Interpreter) evalOnFreshGoroutine(node ast.Node, env map[string]interface{}) (interface{}, error) {
	type outcome struct {
		val interface{}
		err error
	}
	ch := make(chan outcome, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- outcome{nil, fmt.Errorf("interpreter panic: %v", r)}
			}
		}()
		val, err := i.Eval(node, env)
		ch <- outcome{val, err}
	}()
	out := <-ch
	return out.val, out.err
}

// String returns a string representation
func (c *Closure) String() string {
	if c.Name != "" {
		return fmt.Sprintf("<function %s(%s)>", c.qualifiedName(), stringJoin(c.Params, ", "))
	}
	return fmt.Sprintf("<function(%s)>", stringJoin(c.Params, ", "))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// Interpreter interprets Layer programs
type Interpreter struct {
	layerContext *layers.LayerContext
	localEnv     map[string]interface{} // Local variable environment
	execLayers   []string               // Stack of layers currently executing
	depth        int                    // current Layer call depth
	// combining counts, per name, the combined (chain/collect/fallback)
	// calls currently running; see lookupWithBlend.
	combining map[string]int

	// MaxCallDepth bounds recursion; see DefaultMaxCallDepth.
	MaxCallDepth int

	// LogSink receives every line produced by log/trace/print/... as it is
	// emitted. NewInterpreter sets it to write to stderr; set it to nil to
	// only collect lines (see Logs).
	LogSink func(line string)
	logs    []string
}

// emit records a log line and forwards it to LogSink.
func (i *Interpreter) emit(line string) {
	i.logs = append(i.logs, line)
	if i.LogSink != nil {
		i.LogSink(line)
	}
}

// Logs returns every line emitted so far by log, trace, print, etc.
func (i *Interpreter) Logs() []string {
	return i.logs
}

func (i *Interpreter) pushExecLayer(layer string) {
	i.execLayers = append(i.execLayers, layer)
}

func (i *Interpreter) popExecLayer() {
	i.execLayers = i.execLayers[:len(i.execLayers)-1]
}

// executingLayer returns the layer whose code is currently running, or ""
// for top-level code (which is unrestricted).
func (i *Interpreter) executingLayer() string {
	if len(i.execLayers) == 0 {
		return ""
	}
	return i.execLayers[len(i.execLayers)-1]
}

// definingLayer returns the layer a closure created right now should belong
// to: the currently executing layer at runtime, or the layer being declared
// during the definition phase.
func (i *Interpreter) definingLayer() string {
	if layer := i.executingLayer(); layer != "" {
		return layer
	}
	return i.layerContext.GetCurrentLayer()
}

// checkPurityQualified enforces that pure layers cannot use functions from
// impure layers via explicit qualification (e.g. data.write_file).
func (i *Interpreter) checkPurityQualified(node ast.Node, targetLayer, name string) error {
	if msg := i.purityViolationQualified(i.executingLayer(), targetLayer, name); msg != "" {
		return i.error(node, msg)
	}
	return nil
}

// checkPurityUnqualified enforces purity for unqualified names. An
// unqualified name is allowed from a pure layer as long as SOME layer
// provides a pure definition: higher impure layers intercepting the call
// (open recursion) is the outer program's choice, not an effect requested
// by the pure code. A name defined only in impure layers (e.g. log) is an
// explicit effect and is rejected.
func (i *Interpreter) checkPurityUnqualified(node ast.Node, name string) error {
	if msg := i.purityViolationUnqualified(i.executingLayer(), name); msg != "" {
		return i.error(node, msg)
	}
	return nil
}

// purityViolationQualified returns a violation message, or "" when caller
// may reference targetLayer.name. Shared by the runtime and static checks.
func (i *Interpreter) purityViolationQualified(caller, targetLayer, name string) string {
	if caller == "" || !i.layerContext.IsPureLayer(caller) {
		return "" // top-level and impure layers are unrestricted
	}
	if i.layerContext.IsPureFunction(targetLayer, name) {
		return ""
	}
	return fmt.Sprintf(
		"purity violation: pure layer '%s' cannot call '%s.%s' from impure layer '%s' (mark the layer with @layer<impure> if it needs effects)",
		caller, targetLayer, name, targetLayer)
}

// purityViolationUnqualified is the unqualified counterpart of
// purityViolationQualified.
func (i *Interpreter) purityViolationUnqualified(caller, name string) string {
	if caller == "" || !i.layerContext.IsPureLayer(caller) {
		return ""
	}
	defining := i.layerContext.LayersDefining(name)
	if len(defining) == 0 {
		return "" // let undefined-name handling report it
	}
	for _, layer := range defining {
		if i.layerContext.IsPureFunction(layer, name) {
			return ""
		}
	}
	return fmt.Sprintf(
		"purity violation: pure layer '%s' cannot call '%s', which is only defined in impure layer '%s' (mark the layer with @layer<impure> if it needs effects)",
		caller, name, defining[0])
}

// NewInterpreter creates a new interpreter
func NewInterpreter() *Interpreter {
	interp := &Interpreter{
		layerContext: layers.NewLayerContext(),
		localEnv:     make(map[string]interface{}),
		combining:    make(map[string]int),
		MaxCallDepth: DefaultMaxCallDepth,
		LogSink:      func(line string) { fmt.Fprintln(os.Stderr, line) },
	}

	// Set up core layer
	interp.layerContext.AddLayer("core")
	SetupCoreLayer(interp.layerContext, interp)

	return interp
}

// EnableLayer enables a layer
func (i *Interpreter) EnableLayer(layerName string) error {
	i.layerContext.AddLayer(layerName)

	// Set up built-in layer implementations. The ui, data, and debug layers
	// are always impure (code written in them may perform effects). Some of
	// their built-ins are pure exceptions because they only build values:
	// the ui component constructors (rendering is the runtime's job) and
	// data's Option/JSON helpers.
	switch layerName {
	case "debug":
		i.layerContext.SetLayerImpure("debug")
		SetupDebugLayer(i.layerContext, i)
	case "ui":
		i.layerContext.SetLayerImpure("ui")
		for _, name := range uiPureFunctions {
			i.layerContext.MarkPureException("ui", name)
		}
		SetupUILayer(i.layerContext)
	case "data":
		i.layerContext.SetLayerImpure("data")
		for _, name := range []string{"Some", "None", "is_some", "is_none", "unwrap", "map_option", "flatmap_option", "parse_json", "to_json"} {
			i.layerContext.MarkPureException("data", name)
		}
		SetupDataLayer(i.layerContext)
	}

	return nil
}

// error builds a positioned runtime error for node.
func (i *Interpreter) error(node ast.Node, message string) error {
	pos := node.Pos()
	return &RuntimeError{Line: pos.Line, Column: pos.Column, Message: message}
}

// wrap attaches node's position to err unless err already carries one
// (errors raised deeper in the program keep their original location).
func (i *Interpreter) wrap(node ast.Node, err error) error {
	var rerr *RuntimeError
	if errors.As(err, &rerr) {
		return err
	}
	return i.error(node, err.Error())
}

// Eval evaluates an AST node
func (i *Interpreter) Eval(node ast.Node, env map[string]interface{}) (interface{}, error) {
	if env == nil {
		env = i.localEnv
	}

	switch n := node.(type) {
	// Literals
	case *ast.NumberLiteral:
		return n.Value, nil

	case *ast.StringLiteral:
		return n.Value, nil

	case *ast.BooleanLiteral:
		return n.Value, nil

	case *ast.InterpolatedString:
		var out []byte
		for _, part := range n.Parts {
			val, err := i.Eval(part, env)
			if err != nil {
				return nil, err
			}
			out = append(out, Show(val)...)
		}
		return string(out), nil

	case *ast.ListLiteral:
		elements := make([]interface{}, len(n.Elements))
		for idx, elem := range n.Elements {
			val, err := i.Eval(elem, env)
			if err != nil {
				return nil, err
			}
			elements[idx] = val
		}
		return elements, nil

	case *ast.TupleLiteral:
		elements := make([]interface{}, len(n.Elements))
		for idx, elem := range n.Elements {
			val, err := i.Eval(elem, env)
			if err != nil {
				return nil, err
			}
			elements[idx] = val
		}
		return elements, nil

	case *ast.RecordLiteral:
		record := make(map[string]interface{})
		for key, valueNode := range n.Fields {
			val, err := i.Eval(valueNode, env)
			if err != nil {
				return nil, err
			}
			record[key] = val
		}
		return record, nil

	// Identifiers
	case *ast.Identifier:
		// Look up in local environment first
		if val, found := env[n.Name]; found {
			return val, nil
		}

		if err := i.checkPurityUnqualified(node, n.Name); err != nil {
			return nil, err
		}

		// Then in layer context, applying blend mode
		return i.lookupWithBlend(node, n.Name)

	// Member access: record.field, or layer.function. A local name (a
	// parameter or let binding) shadows a layer of the same name.
	case *ast.MemberAccess:
		if ident, ok := n.Object.(*ast.Identifier); ok {
			if val, local := env[ident.Name]; local {
				return i.fieldOf(node, val, n.Member)
			}
			layerName := ident.Name
			if err := i.checkPurityQualified(node, layerName, n.Member); err != nil {
				return nil, err
			}
			val, err := i.layerContext.GetLayerValue(layerName, n.Member)
			if err != nil {
				return nil, i.error(node, err.Error())
			}
			return val, nil
		}
		obj, err := i.Eval(n.Object, env)
		if err != nil {
			return nil, err
		}
		return i.fieldOf(node, obj, n.Member)

	// Record update: {record | field: value}
	case *ast.RecordUpdate:
		base, err := i.Eval(n.Record, env)
		if err != nil {
			return nil, err
		}
		record, ok := base.(map[string]interface{})
		if !ok {
			return nil, i.error(n.Record, fmt.Sprintf("only records can be updated with {… | field: value}, got %s", typeName(base)))
		}
		updated := make(map[string]interface{}, len(record))
		for k, v := range record {
			updated[k] = v
		}
		for key, valueNode := range n.Fields {
			if _, exists := record[key]; !exists {
				return nil, i.error(valueNode, fmt.Sprintf("cannot update field '%s': the record has no such field (fields: %s)", key, fieldList(record)))
			}
			val, err := i.Eval(valueNode, env)
			if err != nil {
				return nil, err
			}
			updated[key] = val
		}
		return updated, nil

	// Functions
	case *ast.Lambda:
		envCopy := make(map[string]interface{}, len(env))
		for k, v := range env {
			envCopy[k] = v
		}
		return &Closure{
			Params:      n.Params,
			Body:        n.Body,
			Env:         envCopy,
			Interpreter: i,
			Layer:       i.definingLayer(),
		}, nil

	case *ast.FunctionDef:
		envCopy := make(map[string]interface{}, len(env))
		for k, v := range env {
			envCopy[k] = v
		}
		closure := &Closure{
			Name:        n.Name,
			Params:      n.Params,
			Body:        n.Body,
			Env:         envCopy,
			Interpreter: i,
			Layer:       i.definingLayer(),
		}
		// Store in current layer
		if err := i.layerContext.Define(n.Name, closure); err != nil {
			return nil, i.error(node, err.Error())
		}
		return closure, nil

	case *ast.FunctionCall:
		funcVal, err := i.Eval(n.Function, env)
		if err != nil {
			return nil, err
		}

		if !isCallable(funcVal) {
			return nil, i.error(node, fmt.Sprintf("not a function: %s", Show(funcVal)))
		}

		args := make([]interface{}, len(n.Arguments))
		for idx, argNode := range n.Arguments {
			val, err := i.Eval(argNode, env)
			if err != nil {
				return nil, err
			}
			args[idx] = val
		}

		result, err := callValue(funcVal, args...)
		if err != nil {
			return nil, i.wrap(node, err)
		}
		return result, nil

	// Binary operations
	case *ast.BinaryOp:
		left, err := i.Eval(n.Left, env)
		if err != nil {
			return nil, err
		}

		right, err := i.Eval(n.Right, env)
		if err != nil {
			return nil, err
		}

		// Look up operator function
		opFunc, err := i.layerContext.Lookup(n.Operator)
		if err != nil {
			return nil, i.error(node, fmt.Sprintf("unknown operator: %s", n.Operator))
		}

		result, err := callValue(opFunc, left, right)
		if err != nil {
			return nil, i.wrap(node, err)
		}
		return result, nil

	// Unary operations
	case *ast.UnaryOp:
		operand, err := i.Eval(n.Operand, env)
		if err != nil {
			return nil, err
		}

		switch n.Operator {
		case "-":
			result, err := negateNumber(operand)
			if err != nil {
				return nil, i.wrap(node, err)
			}
			return result, nil
		case "not":
			opFunc, err := i.layerContext.Lookup("not")
			if err != nil {
				return nil, i.error(node, "not operator not found")
			}
			result, err := callValue(opFunc, operand)
			if err != nil {
				return nil, i.wrap(node, err)
			}
			return result, nil
		}

		return nil, i.error(node, fmt.Sprintf("unknown unary operator: %s", n.Operator))

	// Conditional
	case *ast.IfExpr:
		cond, err := i.Eval(n.Cond, env)
		if err != nil {
			return nil, err
		}
		b, ok := cond.(bool)
		if !ok {
			return nil, i.error(n.Cond, fmt.Sprintf("if condition must be a boolean, got %s", typeName(cond)))
		}
		if b {
			return i.Eval(n.Then, env)
		}
		return i.Eval(n.Else, env)

	// Let binding
	case *ast.LetBinding:
		value, err := i.Eval(n.Value, env)
		if err != nil {
			return nil, err
		}

		newEnv := make(map[string]interface{}, len(env)+1)
		for k, v := range env {
			newEnv[k] = v
		}
		newEnv[n.Name] = value

		return i.Eval(n.Body, newEnv)

	// Pattern matching
	case *ast.Match:
		value, err := i.Eval(n.Value, env)
		if err != nil {
			return nil, err
		}

		for _, matchCase := range n.Cases {
			bindings := i.matchPattern(matchCase.Pattern, value, env)
			if bindings != nil {
				// Build env with pattern bindings for guard and body
				newEnv := make(map[string]interface{}, len(env)+len(bindings))
				for k, v := range env {
					newEnv[k] = v
				}
				for k, v := range bindings {
					newEnv[k] = v
				}

				// Evaluate guard if present
				if matchCase.Guard != nil {
					guardResult, err := i.Eval(matchCase.Guard, newEnv)
					if err != nil {
						return nil, err
					}
					guardBool, ok := guardResult.(bool)
					if !ok {
						return nil, i.error(matchCase.Guard, fmt.Sprintf("guard expression must evaluate to a boolean, got %s", typeName(guardResult)))
					}
					if !guardBool {
						continue // Guard failed, try next case
					}
				}

				return i.Eval(matchCase.Body, newEnv)
			}
		}

		return nil, i.error(node, fmt.Sprintf("no pattern matched for value: %s", ShowNested(value)))

	// Introspection expressions
	case *ast.IntrospectionExpr:
		switch n.Op {
		case "which":
			layerName := i.layerContext.WhichLayer(n.Args[0])
			if layerName == "" {
				return nil, i.error(node, fmt.Sprintf("@which: '%s' is not defined in any layer", n.Args[0]))
			}
			return layerName, nil

		case "inspect":
			names, err := i.layerContext.InspectLayer(n.Args[0])
			if err != nil {
				return nil, i.error(node, fmt.Sprintf("@inspect: %s", err))
			}
			result := make([]interface{}, len(names))
			for idx, name := range names {
				result[idx] = name
			}
			return result, nil

		case "diff":
			added, overridden, err := i.layerContext.DiffLayers(n.Args[0], n.Args[1])
			if err != nil {
				return nil, i.error(node, fmt.Sprintf("@diff: %s", err))
			}
			toIfaceSlice := func(ss []string) []interface{} {
				out := make([]interface{}, len(ss))
				for idx, s := range ss {
					out[idx] = s
				}
				return out
			}
			return map[string]interface{}{
				"added":      toIfaceSlice(added),
				"overridden": toIfaceSlice(overridden),
			}, nil
		}

	// Block
	case *ast.Block:
		var result interface{}
		for _, stmt := range n.Statements {
			val, err := i.Eval(stmt, env)
			if err != nil {
				return nil, err
			}
			result = val
		}
		return result, nil
	}

	return nil, i.error(node, fmt.Sprintf("unknown node type: %T", node))
}

// fieldOf reads record.field.
func (i *Interpreter) fieldOf(node ast.Node, val interface{}, field string) (interface{}, error) {
	record, ok := val.(map[string]interface{})
	if !ok {
		return nil, i.error(node, fmt.Sprintf("cannot read field '%s' of %s", field, typeName(val)))
	}
	v, exists := record[field]
	if !exists {
		return nil, i.error(node, fmt.Sprintf("the record has no field '%s' (fields: %s)", field, fieldList(record)))
	}
	return v, nil
}

// fieldList renders a record's field names, sorted, for messages.
func fieldList(record map[string]interface{}) string {
	if len(record) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}

// matchPattern matches a pattern against a value
// Returns a map of bindings if match succeeds, nil otherwise
func (i *Interpreter) matchPattern(pattern ast.Node, value interface{}, env map[string]interface{}) map[string]interface{} {
	switch p := pattern.(type) {
	case *ast.NumberPattern:
		if numEqual(p.Value, value) {
			return make(map[string]interface{})
		}
		return nil

	case *ast.StringPattern:
		if str, ok := value.(string); ok && p.Value == str {
			return make(map[string]interface{})
		}
		return nil

	case *ast.BooleanPattern:
		if b, ok := value.(bool); ok && p.Value == b {
			return make(map[string]interface{})
		}
		return nil

	case *ast.IdentifierPattern:
		if p.Name == "_" {
			return make(map[string]interface{}) // Wildcard, always matches
		}
		return map[string]interface{}{p.Name: value}

	case *ast.RecordPattern:
		record, ok := value.(map[string]interface{})
		if !ok {
			return nil
		}

		bindings := make(map[string]interface{})
		for fieldName, subPattern := range p.Fields {
			fieldVal, exists := record[fieldName]
			if !exists {
				return nil // required field missing
			}
			fieldBindings := i.matchPattern(subPattern, fieldVal, env)
			if fieldBindings == nil {
				return nil
			}
			for k, v := range fieldBindings {
				bindings[k] = v
			}
		}
		return bindings

	case *ast.ListPattern:
		list, ok := value.([]interface{})
		if !ok {
			return nil
		}

		bindings := make(map[string]interface{})

		if p.Rest != "" {
			// Pattern like [a, b, ...rest]
			if len(list) < len(p.Elements) {
				return nil
			}

			// Match fixed elements
			for idx, elemPattern := range p.Elements {
				elemBindings := i.matchPattern(elemPattern, list[idx], env)
				if elemBindings == nil {
					return nil
				}
				for k, v := range elemBindings {
					bindings[k] = v
				}
			}

			// Bind rest
			restValue := list[len(p.Elements):]
			bindings[p.Rest] = restValue

		} else {
			// Exact length match
			if len(list) != len(p.Elements) {
				return nil
			}

			for idx, elemPattern := range p.Elements {
				elemBindings := i.matchPattern(elemPattern, list[idx], env)
				if elemBindings == nil {
					return nil
				}
				for k, v := range elemBindings {
					bindings[k] = v
				}
			}
		}

		return bindings
	}

	return nil
}

// isCallable reports whether val can be called as a function.
func isCallable(val interface{}) bool {
	switch val.(type) {
	case *Closure, func(...interface{}) (interface{}, error):
		return true
	}
	return false
}

// CallValue calls a Layer function value (a closure or a built-in) — used
// by runtimes that drive apps (update, view).
func CallValue(fn interface{}, args ...interface{}) (interface{}, error) {
	return callValue(fn, args...)
}

// callValue invokes a value that may be a *Closure or a built-in Go function.
func callValue(val interface{}, args ...interface{}) (interface{}, error) {
	if closure, ok := val.(*Closure); ok {
		return closure.Call(args...)
	}
	if fn, ok := val.(func(...interface{}) (interface{}, error)); ok {
		return fn(args...)
	}
	return nil, fmt.Errorf("not a function: %s", Show(val))
}

// isNoneValue checks whether a value is the data layer's None option type.
func isNoneValue(val interface{}) bool {
	if m, ok := val.(map[string]interface{}); ok {
		if t, ok := m["type"].(string); ok {
			return t == "None"
		}
	}
	return false
}

// lookupWithBlend resolves a name from the layer context, applying the
// blend mode for that name (its per-function mode, or the default).
//
// override resolves to the topmost definition; unqualified calls inside it
// re-enter the top of the stack (open recursion), which is what lets a
// wrapping layer observe every recursive step.
//
// The combining modes (chain, collect, fallback) apply once per outermost
// call: while a combined call to name is running, an unqualified call to
// name from a layer that defines it resolves to that layer's own
// definition. Recursion inside a participating layer therefore stays in
// that layer, and the combination happens around it exactly once.
func (i *Interpreter) lookupWithBlend(node ast.Node, name string) (interface{}, error) {
	mode := i.layerContext.BlendModeFor(name)

	if mode == "override" {
		val, err := i.layerContext.Lookup(name)
		if err != nil {
			return nil, i.error(node, fmt.Sprintf("undefined variable: %s", name))
		}
		return val, nil
	}

	// Collect all definitions across layers (top to bottom).
	values := i.layerContext.LookupAll(name)
	if len(values) == 0 {
		return nil, i.error(node, fmt.Sprintf("undefined variable: %s", name))
	}
	if len(values) == 1 {
		return values[0], nil
	}

	// Inside a combined call to name: recursion stays in the calling layer.
	if i.combining[name] > 0 {
		if layer := i.executingLayer(); layer != "" {
			if own, err := i.layerContext.GetLayerValue(layer, name); err == nil {
				return own, nil
			}
		}
	}

	layerNames := i.layerContext.LayersDefining(name) // parallel to values
	var combine func(args ...interface{}) (interface{}, error)

	switch mode {
	case "chain":
		// Pipe result from bottom layer up through each higher layer.
		// values[0] = top layer, values[len-1] = bottom layer.
		// Partial application semantics: each upper layer is called with (result, args[1], args[2], …).
		// The first parameter is always the piped value; remaining parameters are forwarded
		// unchanged from the original call, so multi-param functions work naturally.
		combine = func(args ...interface{}) (interface{}, error) {
			result, err := callValue(values[len(values)-1], args...)
			if err != nil {
				return nil, err
			}
			for j := len(values) - 2; j >= 0; j-- {
				// Replace the first arg with the piped result; keep the rest.
				chainArgs := append([]interface{}{result}, args[1:]...)
				result, err = callValue(values[j], chainArgs...)
				if err != nil {
					return nil, err
				}
			}
			return result, nil
		}

	case "collect":
		// Call every layer that defines the name; return results as a list.
		combine = func(args ...interface{}) (interface{}, error) {
			results := make([]interface{}, 0, len(values))
			for _, val := range values {
				res, err := callValue(val, args...)
				if err != nil {
					return nil, err
				}
				results = append(results, res)
			}
			return results, nil
		}

	case "fallback":
		// Try from top layer down; return first result that is not None.
		// Priority: non-None result > None result > error.
		// A layer that errors is skipped, but the error is reported in the
		// log so a bug in (say) a cache layer never silently looks like a
		// cache miss.
		combine = func(args ...interface{}) (interface{}, error) {
			var lastNone interface{}
			var lastErr error
			for idx, val := range values {
				res, err := callValue(val, args...)
				if err != nil {
					lastErr = err
					i.emit(fmt.Sprintf("[WARN] fallback: %s.%s failed, trying the next layer: %v", layerNames[idx], name, err))
					continue
				}
				if !isNoneValue(res) {
					return res, nil
				}
				lastNone = res
			}
			// At least one layer returned None (prefer that over an error).
			if lastNone != nil {
				return lastNone, nil
			}
			// All layers errored — return the last error directly.
			return nil, lastErr
		}

	default:
		return values[0], nil
	}

	return func(args ...interface{}) (interface{}, error) {
		i.combining[name]++
		defer func() { i.combining[name]-- }()
		return combine(args...)
	}, nil
}

// applyPurityAnnotation records a layer's declared purity. Built-in layers
// have fixed purity: core is always pure; ui, data, and debug are always
// impure. Custom layers default to pure and may opt into effects with
// @layer<impure>.
func (i *Interpreter) applyPurityAnnotation(layerDecl *ast.LayerDeclaration) error {
	name := layerDecl.LayerName
	annotation := layerDecl.Purity
	if annotation == "" {
		return nil
	}

	impure := annotation != "pure"
	switch name {
	case "core":
		if impure {
			return i.error(layerDecl, "the core layer is always pure and cannot be marked impure")
		}
	case "ui", "data", "debug":
		if !impure {
			return i.error(layerDecl, fmt.Sprintf("the built-in '%s' layer is always impure and cannot be marked pure", name))
		}
	default:
		if impure {
			i.layerContext.SetLayerImpure(name)
		}
	}
	return nil
}

// SetVar seeds a variable into the interpreter's local environment.
// Call this before Execute to provide values that @layers conditions can
// reference, or (in the visual editor) the current values of UI inputs.
func (i *Interpreter) SetVar(name string, value interface{}) {
	i.localEnv[name] = value
}

// activeDeclarations returns the layer declarations that take part in the
// program. With an @layers directive, a declared layer is active only if it
// is listed (and its condition, if any, held) — a declared-but-unlisted
// layer is "hidden", like a hidden layer in an image editor. core is always
// active. Without a directive, every declared layer is active.
func activeDeclarations(program *ast.Program, enabled map[string]bool) []*ast.LayerDeclaration {
	var active []*ast.LayerDeclaration
	for _, decl := range program.LayerDeclarations {
		if program.LayersDirective != nil && decl.LayerName != "core" && !enabled[decl.LayerName] {
			continue
		}
		active = append(active, decl)
	}
	return active
}

// Load prepares a program without running its main expression: it
// activates the @layers stack, applies purity annotations, evaluates all
// layer definitions, and statically checks every function body (undefined
// names, unknown layers, purity violations).
func (i *Interpreter) Load(program *ast.Program) error {
	// Evaluate conditional @layers directives and activate layers.
	enabled := make(map[string]bool)
	if program.LayersDirective != nil {
		if program.LayersDirective.BlendMode != "" {
			i.layerContext.SetBlendMode(program.LayersDirective.BlendMode)
		}
		for _, entry := range program.LayersDirective.BlendEntries {
			i.layerContext.SetBlendModeFor(entry.Name, entry.Mode)
		}
		for _, entry := range program.LayersDirective.Layers {
			if entry.Condition != nil {
				result, err := i.Eval(entry.Condition, i.localEnv)
				if err != nil {
					return fmt.Errorf("error evaluating condition for layer '%s': %w", entry.Name, err)
				}
				active, ok := result.(bool)
				if !ok {
					return i.error(entry.Condition, fmt.Sprintf("condition for layer '%s' must evaluate to a boolean, got %s", entry.Name, typeName(result)))
				}
				if !active {
					continue
				}
			}
			enabled[entry.Name] = true
			if err := i.EnableLayer(entry.Name); err != nil {
				return err
			}
		}
	}

	decls := activeDeclarations(program, enabled)

	// Apply purity annotations before evaluating any definitions, so that
	// a layer's purity is known regardless of declaration order.
	for _, layerDecl := range decls {
		if err := i.applyPurityAnnotation(layerDecl); err != nil {
			return err
		}
	}

	// Evaluate layer declarations.
	for _, layerDecl := range decls {
		i.layerContext.SetCurrentLayer(layerDecl.LayerName)

		for _, definition := range layerDecl.Definitions {
			if _, err := i.Eval(definition, i.localEnv); err != nil {
				return err
			}
		}
	}

	// Definitions are done; top-level code runs outside any layer so it is
	// unrestricted by purity rules.
	i.layerContext.SetCurrentLayer("")

	return i.check(program, decls)
}

// Execute executes a complete Layer program
func (i *Interpreter) Execute(program *ast.Program) (interface{}, error) {
	if err := i.Load(program); err != nil {
		return nil, err
	}

	// Evaluate main expression
	if program.MainExpression != nil {
		return i.Eval(program.MainExpression, i.localEnv)
	}

	return nil, nil
}

// Helper functions

func stringJoin(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func negateNumber(val interface{}) (interface{}, error) {
	switch v := val.(type) {
	case int64:
		return -v, nil
	case float64:
		return -v, nil
	case int:
		return -int64(v), nil
	default:
		return nil, fmt.Errorf("cannot negate %s", typeName(val))
	}
}

func numEqual(a, b interface{}) bool {
	aFloat := toFloat64(a)
	bFloat := toFloat64(b)
	if aFloat == nil || bFloat == nil {
		return false
	}
	return math.Abs(*aFloat-*bFloat) < 1e-9
}

func toFloat64(val interface{}) *float64 {
	switch v := val.(type) {
	case int64:
		f := float64(v)
		return &f
	case float64:
		return &v
	case int:
		f := float64(v)
		return &f
	default:
		return nil
	}
}
