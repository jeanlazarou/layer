package interpreter

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/yourusername/layer/pkg/ast"
	"github.com/yourusername/layer/pkg/layers"
)

// Closure represents a function closure
type Closure struct {
	Params      []string
	Body        ast.Node
	Env         map[string]interface{}
	Interpreter *Interpreter
}

// Call invokes the closure with arguments
func (c *Closure) Call(args ...interface{}) (interface{}, error) {
	if len(args) != len(c.Params) {
		return nil, fmt.Errorf("expected %d arguments, got %d", len(c.Params), len(args))
	}

	// Special handling for const function
	if c.Body == nil && c.Env["__const_value"] != nil {
		return c.Env["__const_value"], nil
	}

	// Special handling for compose function
	if c.Body == nil && c.Env["__compose_f"] != nil && c.Env["__compose_g"] != nil {
		f := c.Env["__compose_f"].(*Closure)
		g := c.Env["__compose_g"].(*Closure)

		// Call g(x)
		gResult, err := g.Call(args[0])
		if err != nil {
			return nil, err
		}

		// Call f(g(x))
		return f.Call(gResult)
	}

	// Special handling for trace function
	if c.Body == nil && c.Env["__trace_name"] != nil && c.Env["__trace_fn"] != nil {
		name := fmt.Sprintf("%v", c.Env["__trace_name"])
		fn := c.Env["__trace_fn"].(*Closure)

		// Log function call
		argsStr := ""
		for i, arg := range args {
			if i > 0 {
				argsStr += ", "
			}
			argsStr += fmt.Sprintf("%v", arg)
		}
		fmt.Fprintf(os.Stderr, "[TRACE] %s(%s) called\n", name, argsStr)

		// Call the wrapped function
		result, err := fn.Call(args...)
		if err != nil {
			return nil, err
		}

		// Log function result
		fmt.Fprintf(os.Stderr, "[TRACE] %s(%s) = %v\n", name, argsStr, result)

		return result, nil
	}

	// Special handling for benchmark function
	if c.Body == nil && c.Env["__benchmark_name"] != nil && c.Env["__benchmark_fn"] != nil {
		name := fmt.Sprintf("%v", c.Env["__benchmark_name"])
		fn := c.Env["__benchmark_fn"].(*Closure)

		// Benchmark the function
		start := time.Now()
		result, err := fn.Call(args...)
		elapsed := time.Since(start)

		if err != nil {
			return nil, err
		}

		// Log benchmark result
		fmt.Fprintf(os.Stderr, "[BENCH] %s took %.3fms\n", name, float64(elapsed.Microseconds())/1000.0)

		return result, nil
	}

	// Create new environment with parameters bound
	newEnv := make(map[string]interface{})
	for k, v := range c.Env {
		newEnv[k] = v
	}

	for i, param := range c.Params {
		newEnv[param] = args[i]
	}

	// Evaluate body in new environment
	return c.Interpreter.Eval(c.Body, newEnv)
}

// String returns a string representation
func (c *Closure) String() string {
	return fmt.Sprintf("<function(%s)>", stringJoin(c.Params, ", "))
}

// Interpreter interprets Layer programs
type Interpreter struct {
	layerContext *layers.LayerContext
	localEnv     map[string]interface{} // Local variable environment
}

// NewInterpreter creates a new interpreter
func NewInterpreter() *Interpreter {
	interp := &Interpreter{
		layerContext: layers.NewLayerContext(),
		localEnv:     make(map[string]interface{}),
	}

	// Set up core layer
	interp.layerContext.AddLayer("core")
	SetupCoreLayer(interp.layerContext)

	return interp
}

// EnableLayer enables a layer
func (i *Interpreter) EnableLayer(layerName string) error {
	i.layerContext.AddLayer(layerName)

	// Set up built-in layer implementations
	switch layerName {
	case "debug":
		SetupDebugLayer(i.layerContext)
	case "ui":
		SetupUILayer(i.layerContext)
	case "data":
		SetupDataLayer(i.layerContext)
	}

	return nil
}

// error raises an interpreter error
func (i *Interpreter) error(node ast.Node, message string) error {
	pos := node.Pos()
	return fmt.Errorf("runtime error at %d:%d: %s", pos.Line, pos.Column, message)
}

// Eval evaluates an AST node
func (i *Interpreter) Eval(node ast.Node, env map[string]interface{}) (interface{}, error) {
	if env == nil {
		env = i.localEnv
	}

	// Literals
	switch n := node.(type) {
	case *ast.NumberLiteral:
		return n.Value, nil

	case *ast.StringLiteral:
		return n.Value, nil

	case *ast.BooleanLiteral:
		return n.Value, nil

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

		// Then in layer context, applying blend mode
		return i.lookupWithBlend(node, n.Name)

	// Member access (layer.function)
	case *ast.MemberAccess:
		// For now, assume object is a layer name
		if ident, ok := n.Object.(*ast.Identifier); ok {
			layerName := ident.Name
			val, err := i.layerContext.GetLayerValue(layerName, n.Member)
			if err != nil {
				return nil, i.error(node, err.Error())
			}
			return val, nil
		}
		return nil, i.error(node, "member access only supported on layer names currently")

	// Functions
	case *ast.Lambda:
		envCopy := make(map[string]interface{})
		for k, v := range env {
			envCopy[k] = v
		}
		return &Closure{
			Params:      n.Params,
			Body:        n.Body,
			Env:         envCopy,
			Interpreter: i,
		}, nil

	case *ast.FunctionDef:
		envCopy := make(map[string]interface{})
		for k, v := range env {
			envCopy[k] = v
		}
		closure := &Closure{
			Params:      n.Params,
			Body:        n.Body,
			Env:         envCopy,
			Interpreter: i,
		}
		// Store in current layer
		i.layerContext.Define(n.Name, closure)
		return closure, nil

	case *ast.FunctionCall:
		funcVal, err := i.Eval(n.Function, env)
		if err != nil {
			return nil, err
		}

		args := make([]interface{}, len(n.Arguments))
		for idx, argNode := range n.Arguments {
			val, err := i.Eval(argNode, env)
			if err != nil {
				return nil, err
			}
			args[idx] = val
		}

		// Handle closures
		if closure, ok := funcVal.(*Closure); ok {
			return closure.Call(args...)
		}

		// Handle built-in functions (Go functions)
		if fn, ok := funcVal.(func(...interface{}) (interface{}, error)); ok {
			return fn(args...)
		}

		return nil, i.error(node, fmt.Sprintf("not a function: %v", funcVal))

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

		if fn, ok := opFunc.(func(...interface{}) (interface{}, error)); ok {
			return fn(left, right)
		}

		return nil, i.error(node, fmt.Sprintf("operator %s is not callable", n.Operator))

	// Unary operations
	case *ast.UnaryOp:
		operand, err := i.Eval(n.Operand, env)
		if err != nil {
			return nil, err
		}

		if n.Operator == "-" {
			return negateNumber(operand)
		} else if n.Operator == "not" {
			opFunc, err := i.layerContext.Lookup("not")
			if err != nil {
				return nil, i.error(node, "not operator not found")
			}
			if fn, ok := opFunc.(func(...interface{}) (interface{}, error)); ok {
				return fn(operand)
			}
		}

		return nil, i.error(node, fmt.Sprintf("unknown unary operator: %s", n.Operator))

	// Let binding
	case *ast.LetBinding:
		value, err := i.Eval(n.Value, env)
		if err != nil {
			return nil, err
		}

		newEnv := make(map[string]interface{})
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
				newEnv := make(map[string]interface{})
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
						return nil, i.error(matchCase.Guard, fmt.Sprintf("guard expression must evaluate to a boolean, got %T", guardResult))
					}
					if !guardBool {
						continue // Guard failed, try next case
					}
				}

				return i.Eval(matchCase.Body, newEnv)
			}
		}

		return nil, i.error(node, fmt.Sprintf("no pattern matched for value: %v", value))

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

// callValue invokes a value that may be a *Closure or a built-in Go function.
func callValue(val interface{}, args ...interface{}) (interface{}, error) {
	if closure, ok := val.(*Closure); ok {
		return closure.Call(args...)
	}
	if fn, ok := val.(func(...interface{}) (interface{}, error)); ok {
		return fn(args...)
	}
	return nil, fmt.Errorf("not callable: %T", val)
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

// lookupWithBlend resolves a name from the layer context, applying the active blend mode.
func (i *Interpreter) lookupWithBlend(node ast.Node, name string) (interface{}, error) {
	mode := i.layerContext.GetBlendMode()

	if mode == "" || mode == "override" {
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

	switch mode {
	case "chain":
		// Pipe result from bottom layer up through each higher layer.
		// values[0] = top layer, values[len-1] = bottom layer.
		// Partial application semantics: each upper layer is called with (result, args[1], args[2], …).
		// The first parameter is always the piped value; remaining parameters are forwarded
		// unchanged from the original call, so multi-param functions work naturally.
		return func(args ...interface{}) (interface{}, error) {
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
		}, nil

	case "collect":
		// Call every layer that defines the name; return results as a list.
		return func(args ...interface{}) (interface{}, error) {
			results := make([]interface{}, 0, len(values))
			for _, val := range values {
				res, err := callValue(val, args...)
				if err != nil {
					return nil, err
				}
				results = append(results, res)
			}
			return results, nil
		}, nil

	case "fallback":
		// Try from top layer down; return first result that is not None.
		// Priority: non-None result > None result > error.
		return func(args ...interface{}) (interface{}, error) {
			var lastNone interface{}
			var lastErr error
			for _, val := range values {
				res, err := callValue(val, args...)
				if err != nil {
					lastErr = err
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
		}, nil
	}

	return values[0], nil
}

// SetVar seeds a variable into the interpreter's local environment.
// Call this before Execute to provide values that @layers conditions can reference.
func (i *Interpreter) SetVar(name string, value interface{}) {
	i.localEnv[name] = value
}

// Execute executes a complete Layer program
func (i *Interpreter) Execute(program *ast.Program) (interface{}, error) {
	// Evaluate conditional @layers directives and activate layers.
	// disabled tracks layers whose conditions evaluated to false — their
	// @layer declaration blocks are skipped to prevent them sneaking into
	// layerOrder via an internal AddLayer call inside Define.
	disabled := make(map[string]bool)
	if program.LayersDirective != nil {
		if program.LayersDirective.BlendMode != "" {
			i.layerContext.SetBlendMode(program.LayersDirective.BlendMode)
		}
		for _, entry := range program.LayersDirective.Layers {
			if entry.Condition != nil {
				result, err := i.Eval(entry.Condition, i.localEnv)
				if err != nil {
					return nil, fmt.Errorf("error evaluating condition for layer '%s': %w", entry.Name, err)
				}
				active, ok := result.(bool)
				if !ok {
					return nil, fmt.Errorf("condition for layer '%s' must evaluate to a boolean, got %T", entry.Name, result)
				}
				if !active {
					disabled[entry.Name] = true
					continue
				}
			}
			if err := i.EnableLayer(entry.Name); err != nil {
				return nil, err
			}
		}
	}

	// Evaluate layer declarations, skipping disabled layers entirely.
	for _, layerDecl := range program.LayerDeclarations {
		if disabled[layerDecl.LayerName] {
			continue
		}
		i.layerContext.SetCurrentLayer(layerDecl.LayerName)

		for _, definition := range layerDecl.Definitions {
			if _, err := i.Eval(definition, i.localEnv); err != nil {
				return nil, err
			}
		}
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
		return nil, fmt.Errorf("cannot negate non-number: %T", val)
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
