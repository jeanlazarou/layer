package ast

// Node is the interface all AST nodes implement
type Node interface {
	Pos() Position
	String() string
}

// Position represents a position in the source code
type Position struct {
	Line   int
	Column int
}

// Base struct embedded in all AST nodes
type Base struct {
	Line   int
	Column int
}

func (b Base) Pos() Position {
	return Position{Line: b.Line, Column: b.Column}
}

// Literals

type NumberLiteral struct {
	Base
	Value interface{} // int64 or float64
}

func (n *NumberLiteral) String() string {
	return "NumberLiteral"
}

type StringLiteral struct {
	Base
	Value string
}

func (s *StringLiteral) String() string {
	return "StringLiteral"
}

type BooleanLiteral struct {
	Base
	Value bool
}

func (b *BooleanLiteral) String() string {
	return "BooleanLiteral"
}

type ListLiteral struct {
	Base
	Elements []Node
}

func (l *ListLiteral) String() string {
	return "ListLiteral"
}

type RecordLiteral struct {
	Base
	Fields map[string]Node
}

func (r *RecordLiteral) String() string {
	return "RecordLiteral"
}

type TupleLiteral struct {
	Base
	Elements []Node
}

func (t *TupleLiteral) String() string {
	return "TupleLiteral"
}

// Variables and References

type Identifier struct {
	Base
	Name string
}

func (i *Identifier) String() string {
	return "Identifier"
}

type MemberAccess struct {
	Base
	Object Node
	Member string
}

func (m *MemberAccess) String() string {
	return "MemberAccess"
}

// Functions

type Lambda struct {
	Base
	Params []string
	Body   Node
}

func (l *Lambda) String() string {
	return "Lambda"
}

type FunctionDef struct {
	Base
	Name           string
	Params         []string
	Body           Node
	TypeAnnotation interface{} // Optional, can be nil
}

func (f *FunctionDef) String() string {
	return "FunctionDef"
}

type FunctionCall struct {
	Base
	Function  Node
	Arguments []Node
}

func (f *FunctionCall) String() string {
	return "FunctionCall"
}

// InterpolatedString is a double-quoted string containing {expr} segments,
// e.g. "fib({n}) = {result}". Parts alternate between *StringLiteral (the
// literal text) and arbitrary expressions; evaluation concatenates the
// display form of every part.
type InterpolatedString struct {
	Base
	Parts []Node
}

func (s *InterpolatedString) String() string {
	return "InterpolatedString"
}

// Control Flow

// IfExpr is `if cond then a else b`. Both branches are mandatory: every
// expression must produce a value.
type IfExpr struct {
	Base
	Cond Node
	Then Node
	Else Node
}

func (i *IfExpr) String() string {
	return "IfExpr"
}

type LetBinding struct {
	Base
	Name  string
	Value Node
	Body  Node
}

func (l *LetBinding) String() string {
	return "LetBinding"
}

type MatchCase struct {
	Base
	Pattern Node
	Guard   Node // Optional: when condition (nil if absent)
	Body    Node
}

func (m *MatchCase) String() string {
	return "MatchCase"
}

type Match struct {
	Base
	Value Node
	Cases []*MatchCase
}

func (m *Match) String() string {
	return "Match"
}

// Patterns

type NumberPattern struct {
	Base
	Value interface{} // int64 or float64
}

func (n *NumberPattern) String() string {
	return "NumberPattern"
}

type StringPattern struct {
	Base
	Value string
}

func (s *StringPattern) String() string {
	return "StringPattern"
}

type BooleanPattern struct {
	Base
	Value bool
}

func (b *BooleanPattern) String() string {
	return "BooleanPattern"
}

type IdentifierPattern struct {
	Base
	Name string
}

func (i *IdentifierPattern) String() string {
	return "IdentifierPattern"
}

type ListPattern struct {
	Base
	Elements []Node
	Rest     string // For ...rest patterns (empty string if none)
}

func (l *ListPattern) String() string {
	return "ListPattern"
}

// RecordPattern matches a record (partial — extra fields are ignored).
// Each field value is itself a pattern.
type RecordPattern struct {
	Base
	Fields map[string]Node // field name → sub-pattern
}

func (r *RecordPattern) String() string {
	return "RecordPattern"
}

// IntrospectionExpr represents a runtime layer query expression.
// Op is one of "which", "inspect", "diff".
// Args holds the identifier name(s) the query operates on.
//   - @which name     → Args = ["name"]
//   - @inspect layer  → Args = ["layer"]
//   - @diff a b       → Args = ["a", "b"]
type IntrospectionExpr struct {
	Base
	Op   string
	Args []string
}

func (ie *IntrospectionExpr) String() string {
	return "IntrospectionExpr"
}

// Operators

type BinaryOp struct {
	Base
	Operator string
	Left     Node
	Right    Node
}

func (b *BinaryOp) String() string {
	return "BinaryOp"
}

type UnaryOp struct {
	Base
	Operator string
	Operand  Node
}

func (u *UnaryOp) String() string {
	return "UnaryOp"
}

// Layer-specific

type LayerDeclaration struct {
	Base
	LayerName   string
	Definitions []Node
	// Purity holds the optional annotation from @layer<pure|impure|ui|data|debug>.
	// Empty means "use the default" (custom layers are pure by default).
	Purity string
}

func (l *LayerDeclaration) String() string {
	return "LayerDeclaration"
}

// LayerEntry is a single entry in @layers: a name and an optional condition.
// When Condition is nil the layer is always enabled.
type LayerEntry struct {
	Name      string
	Condition Node // nil means unconditional
}

type LayersDirective struct {
	Base
	Layers    []LayerEntry
	BlendMode string // "override" (default), "chain", "collect", "fallback"
}

func (l *LayersDirective) String() string {
	return "LayersDirective"
}

// LayerNames returns just the names, preserving order.
func (l *LayersDirective) LayerNames() []string {
	names := make([]string, len(l.Layers))
	for i, e := range l.Layers {
		names[i] = e.Name
	}
	return names
}

// Program

type Block struct {
	Base
	Statements []Node
}

func (b *Block) String() string {
	return "Block"
}

type Program struct {
	Base
	LayersDirective   *LayersDirective
	LayerDeclarations []*LayerDeclaration
	MainExpression    Node
}

func (p *Program) String() string {
	return "Program"
}
