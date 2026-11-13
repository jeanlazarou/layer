package interpreter

import (
	"fmt"
	"sort"

	"github.com/jeanlazarou/layer/pkg/ast"
)

// Static checks run by Load after every layer definition is registered and
// before the main expression executes. They walk all function bodies — not
// just the paths a particular run happens to take — and report:
//
//   - names that no scope and no active layer defines
//   - layer-qualified references to unknown/inactive layers or missing names
//   - purity violations (a pure layer referencing an effect)
//
// Names and layers are static in Layer, so these are exact, not heuristics.
// The same purity rules are also enforced at runtime as a safety net.

// scope is an immutable linked set of locally bound names.
type scope struct {
	name   string
	parent *scope
}

func (s *scope) has(name string) bool {
	for cur := s; cur != nil; cur = cur.parent {
		if cur.name == name {
			return true
		}
	}
	return false
}

func (s *scope) with(names ...string) *scope {
	out := s
	for _, n := range names {
		out = &scope{name: n, parent: out}
	}
	return out
}

type checker struct {
	interp *Interpreter
	layer  string // layer whose code is being checked ("" = top level)
	issues []CheckIssue
	seen   map[string]bool // de-duplicates identical issues at one position
}

func (c *checker) report(node ast.Node, msg string) {
	pos := node.Pos()
	key := fmt.Sprintf("%d:%d:%s", pos.Line, pos.Column, msg)
	if c.seen[key] {
		return
	}
	c.seen[key] = true
	c.issues = append(c.issues, CheckIssue{Line: pos.Line, Column: pos.Column, Message: msg})
}

// check runs the static checks over the active layer declarations, the
// main expression, and the per-function blend modes of the directive.
func (i *Interpreter) check(program *ast.Program, decls []*ast.LayerDeclaration) error {
	c := &checker{interp: i, seen: make(map[string]bool)}
	main := program.MainExpression

	// A per-function blend mode must name a function some layer defines.
	// Inactive (hidden) layers count: hiding a layer must not turn its
	// blend entry into an error. A name nobody defines is most likely a typo.
	if program.LayersDirective != nil {
		declared := make(map[string]bool)
		for _, decl := range program.LayerDeclarations {
			for _, def := range decl.Definitions {
				if fd, ok := def.(*ast.FunctionDef); ok {
					declared[fd.Name] = true
				}
			}
		}
		for _, entry := range program.LayersDirective.BlendEntries {
			if !declared[entry.Name] && len(i.layerContext.LayersDefining(entry.Name)) == 0 {
				c.issues = append(c.issues, CheckIssue{
					Line: entry.Line, Column: entry.Column,
					Message: fmt.Sprintf("blend mode given for '%s', but no layer defines a function with that name", entry.Name),
				})
			}
		}
	}

	// Variables seeded with SetVar (conditions, UI inputs) are in scope
	// everywhere.
	var globals *scope
	for name := range i.localEnv {
		globals = globals.with(name)
	}

	for _, decl := range decls {
		c.layer = decl.LayerName
		for _, def := range decl.Definitions {
			c.walk(def, globals)
		}
	}

	c.layer = ""
	if main != nil {
		c.walk(main, globals)
	}

	if len(c.issues) == 0 {
		return nil
	}
	sort.SliceStable(c.issues, func(a, b int) bool {
		if c.issues[a].Line != c.issues[b].Line {
			return c.issues[a].Line < c.issues[b].Line
		}
		return c.issues[a].Column < c.issues[b].Column
	})
	return &CheckError{Issues: c.issues}
}

func (c *checker) walk(node ast.Node, sc *scope) {
	lc := c.interp.layerContext

	switch n := node.(type) {
	case nil:
		return

	case *ast.NumberLiteral, *ast.StringLiteral, *ast.BooleanLiteral, *ast.IntrospectionExpr:
		return

	case *ast.InterpolatedString:
		for _, part := range n.Parts {
			c.walk(part, sc)
		}

	case *ast.ListLiteral:
		for _, e := range n.Elements {
			c.walk(e, sc)
		}

	case *ast.TupleLiteral:
		for _, e := range n.Elements {
			c.walk(e, sc)
		}

	case *ast.RecordLiteral:
		for _, v := range n.Fields {
			c.walk(v, sc)
		}

	case *ast.Identifier:
		if sc.has(n.Name) {
			return
		}
		if len(lc.LayersDefining(n.Name)) == 0 {
			c.report(n, c.undefinedMessage(n.Name))
			return
		}
		if msg := c.interp.purityViolationUnqualified(c.layer, n.Name); msg != "" {
			c.report(n, msg)
		}

	case *ast.MemberAccess:
		ident, ok := n.Object.(*ast.Identifier)
		if !ok {
			c.walk(n.Object, sc)
			return
		}
		if sc.has(ident.Name) {
			return // a local value; member access on it fails at runtime with a clear message
		}
		if !c.layerActive(ident.Name) {
			c.report(n, fmt.Sprintf("layer '%s' is not active (add it to @layers to use %s.%s)", ident.Name, ident.Name, n.Member))
			return
		}
		if _, err := lc.GetLayerValue(ident.Name, n.Member); err != nil {
			c.report(n, fmt.Sprintf("'%s.%s' is not defined: layer '%s' has no '%s'", ident.Name, n.Member, ident.Name, n.Member))
			return
		}
		if msg := c.interp.purityViolationQualified(c.layer, ident.Name, n.Member); msg != "" {
			c.report(n, msg)
		}

	case *ast.FunctionDef:
		c.walk(n.Body, sc.with(n.Params...))

	case *ast.Lambda:
		c.walk(n.Body, sc.with(n.Params...))

	case *ast.FunctionCall:
		c.walk(n.Function, sc)
		for _, a := range n.Arguments {
			c.walk(a, sc)
		}

	case *ast.BinaryOp:
		c.walk(n.Left, sc)
		c.walk(n.Right, sc)

	case *ast.UnaryOp:
		c.walk(n.Operand, sc)

	case *ast.IfExpr:
		c.walk(n.Cond, sc)
		c.walk(n.Then, sc)
		c.walk(n.Else, sc)

	case *ast.LetBinding:
		c.walk(n.Value, sc) // let is not recursive: the name is not in scope in its own value
		c.walk(n.Body, sc.with(n.Name))

	case *ast.Match:
		c.walk(n.Value, sc)
		for _, mc := range n.Cases {
			inner := sc.with(patternBindings(mc.Pattern)...)
			c.walk(mc.Guard, inner)
			c.walk(mc.Body, inner)
		}

	case *ast.Block:
		for _, s := range n.Statements {
			c.walk(s, sc)
		}
	}
}

func (c *checker) layerActive(name string) bool {
	for _, l := range c.interp.layerContext.GetLayers() {
		if l == name {
			return true
		}
	}
	return false
}

// undefinedMessage explains an undefined name, pointing at the built-in
// layer that provides it when that layer simply isn't enabled.
func (c *checker) undefinedMessage(name string) string {
	if layer, ok := builtinProviders[name]; ok && !c.layerActive(layer) {
		return fmt.Sprintf("undefined name '%s' — it is provided by the '%s' layer, which is not in @layers", name, layer)
	}
	return fmt.Sprintf("undefined name '%s'", name)
}

// builtinProviders maps names defined by the optional built-in layers to
// the layer that provides them, for friendlier "undefined name" messages.
var builtinProviders = map[string]string{
	"log": "debug", "trace": "debug", "inspect": "debug", "assert_equal": "debug",
	"benchmark": "debug", "breakpoint": "debug", "get_logs": "debug", "clear_logs": "debug",
	"view": "ui", "text": "ui", "button": "ui", "input": "ui", "container": "ui",
	"row": "ui", "column": "ui", "render": "ui",
	"Some": "data", "None": "data", "is_some": "data", "is_none": "data", "unwrap": "data",
	"map_option": "data", "flatmap_option": "data", "read_file": "data", "write_file": "data",
	"exists": "data", "delete": "data", "list_files": "data", "read_json": "data",
	"write_json": "data", "parse_json": "data", "to_json": "data", "get": "data", "set": "data",
	"delete_key": "data", "keys": "data", "clear": "data",
}

// patternBindings returns the names a pattern binds.
func patternBindings(p ast.Node) []string {
	switch pat := p.(type) {
	case *ast.IdentifierPattern:
		if pat.Name == "_" {
			return nil
		}
		return []string{pat.Name}
	case *ast.ListPattern:
		var names []string
		for _, e := range pat.Elements {
			names = append(names, patternBindings(e)...)
		}
		if pat.Rest != "" {
			names = append(names, pat.Rest)
		}
		return names
	case *ast.RecordPattern:
		var names []string
		for _, sub := range pat.Fields {
			names = append(names, patternBindings(sub)...)
		}
		return names
	}
	return nil
}
