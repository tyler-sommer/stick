package parse

import (
	"fmt"
	"sort"
	"strings"
)

// Node is an item in the AST.
type Node interface {
	String() string // String representation of the Node, for debugging.
	Start() Pos     // The position of the Node in the source code.
	All() []Node    // All children of the Node.
}

// A TrimmableNode contains information on whether preceding or trailing whitespace should
// be removed when executing the template.
type TrimmableNode struct {
	TrimBefore bool // True if whitespace before the node should be removed.
	TrimAfter  bool // True if whitespace after the node should be removed.
}

// Pos is used to track line and offset in a given string.
type Pos struct {
	Line   int
	Offset int
}

// Start returns the start position of the node.
func (p Pos) Start() Pos {
	return p
}

// String returns a string representation of a pos.
func (p Pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Offset)
}

// ModuleNode represents a root node in the AST.
type ModuleNode struct {
	*BodyNode
	Parent *ExtendsNode // Parent template reference.
	Origin string       // The name where this module is originally defined.
}

// NewModuleNode returns a ModuleNode.
func NewModuleNode(name string, nodes ...Node) *ModuleNode {
	return &ModuleNode{NewBodyNode(Pos{1, 0}, nodes...), nil, name}
}

// String returns a string representation of a ModuleNode.
func (l *ModuleNode) String() string {
	return fmt.Sprintf("Module%s", l.Nodes)
}

// BodyNode represents a list of nodes.
type BodyNode struct {
	Pos
	Nodes []Node
}

// NewBodyNode returns a BodyNode.
func NewBodyNode(pos Pos, nodes ...Node) *BodyNode {
	return &BodyNode{pos, nodes}
}

// Append a Node to the BodyNode.
func (l *BodyNode) Append(n Node) {
	l.Nodes = append(l.Nodes, n)
}

// String returns a string representation of a BodyNode.
func (l *BodyNode) String() string {
	return fmt.Sprintf("Body%s", l.Nodes)
}

// All returns all the child Nodes in a BodyNode.
func (l *BodyNode) All() []Node {
	return l.Nodes
}

// TextNode represents raw, non Stick source code, like plain HTML.
type TextNode struct {
	Pos
	Data string // Textual data in the node.
}

// NewTextNode returns a TextNode.
func NewTextNode(data string, p Pos) *TextNode {
	return &TextNode{p, data}
}

// String returns a string representation of a TextNode.
func (t *TextNode) String() string {
	return fmt.Sprintf("Text(%s)", t.Data)
}

// All returns all the child Nodes in a TextNode.
func (t *TextNode) All() []Node {
	return []Node{}
}

// CommentNode represents a comment.
type CommentNode struct {
	*TextNode
	TrimmableNode
}

// NewCommentNode returns a CommentNode.
func NewCommentNode(data string, p Pos) *CommentNode {
	return &CommentNode{NewTextNode(data, p), TrimmableNode{}}
}

// PrintNode represents a print statement
type PrintNode struct {
	Pos
	TrimmableNode
	X Expr // Expression to print.
}

// NewPrintNode returns a PrintNode.
func NewPrintNode(exp Expr, p Pos) *PrintNode {
	return &PrintNode{p, TrimmableNode{}, exp}
}

// String returns a string representation of a PrintNode.
func (t *PrintNode) String() string {
	return fmt.Sprintf("Print(%s)", t.X)
}

// All returns all the child Nodes in a PrintNode.
func (t *PrintNode) All() []Node {
	return []Node{t.X}
}

// BlockNode represents a block statement
type BlockNode struct {
	Pos
	TrimmableNode
	Name   string // Name of the block.
	Body   Node   // Body of the block.
	Origin string // The name where this block is originally defined.
}

// NewBlockNode returns a BlockNode.
func NewBlockNode(name string, body Node, p Pos) *BlockNode {
	return &BlockNode{p, TrimmableNode{}, name, body, ""}
}

// String returns a string representation of a BlockNode.
func (t *BlockNode) String() string {
	return fmt.Sprintf("Block(%s: %s)", t.Name, t.Body)
}

// All returns all the child Nodes in a BlockNode.
func (t *BlockNode) All() []Node {
	return []Node{t.Body}
}

// IfNode represents an if statement
type IfNode struct {
	Pos
	TrimmableNode
	Cond Expr // Condition to test.
	Body Node // Body to evaluate if Cond is true.
	Else Node // Body if Cond is false.
}

// NewIfNode returns a IfNode.
func NewIfNode(cond Expr, body Node, els Node, p Pos) *IfNode {
	return &IfNode{p, TrimmableNode{}, cond, body, els}
}

// String returns a string representation of an IfNode.
func (t *IfNode) String() string {
	return fmt.Sprintf("If(%s: %s Else: %s)", t.Cond, t.Body, t.Else)
}

// All returns all the child Nodes in a IfNode.
func (t *IfNode) All() []Node {
	return []Node{t.Cond, t.Body, t.Else}
}

// ExtendsNode represents an extends statement
type ExtendsNode struct {
	Pos
	TrimmableNode
	Tpl Expr // Name of the template being extended.
}

// NewExtendsNode returns a ExtendsNode.
func NewExtendsNode(tplRef Expr, p Pos) *ExtendsNode {
	return &ExtendsNode{p, TrimmableNode{}, tplRef}
}

// String returns a string representation of an ExtendsNode.
func (t *ExtendsNode) String() string {
	return fmt.Sprintf("Extends(%s)", t.Tpl)
}

// All returns all the child Nodes in a ExtendsNode.
func (t *ExtendsNode) All() []Node {
	return []Node{t.Tpl}
}

// ForNode represents a for loop construct.
type ForNode struct {
	Pos
	TrimmableNode
	Key  string // Name of key variable, or empty string.
	Val  string // Name of val variable.
	X    Expr   // Expression to iterate over.
	Body Node   // Body of the for loop.
	Else Node   // Body of the else section if X is empty.
}

// NewForNode returns a ForNode.
func NewForNode(k, v string, expr Expr, body, els Node, p Pos) *ForNode {
	return &ForNode{p, TrimmableNode{}, k, v, expr, body, els}
}

// String returns a string representation of a ForNode.
func (t *ForNode) String() string {
	return fmt.Sprintf("For(%s, %s in %s: %s else %s)", t.Key, t.Val, t.X, t.Body, t.Else)
}

// All returns all the child Nodes in a ForNode.
func (t *ForNode) All() []Node {
	return []Node{t.X, t.Body, t.Else}
}

// IncludeNode is an include statement.
type IncludeNode struct {
	Pos
	TrimmableNode
	Tpl  Expr // Expression evaluating to the name of the template to include.
	With Expr // Explicit list of variables to include in the included template.
	Only bool // If true, only vars defined in With will be passed.
}

// NewIncludeNode returns a IncludeNode.
func NewIncludeNode(tmpl Expr, with Expr, only bool, pos Pos) *IncludeNode {
	return &IncludeNode{pos, TrimmableNode{}, tmpl, with, only}
}

// String returns a string representation of an IncludeNode.
func (t *IncludeNode) String() string {
	return fmt.Sprintf("Include(%s with %s %v)", t.Tpl, t.With, t.Only)
}

// All returns all the child Nodes in a IncludeNode.
func (t *IncludeNode) All() []Node {
	return []Node{t.Tpl, t.With}
}

// EmbedNode is a special include statement.
type EmbedNode struct {
	*IncludeNode
	Blocks map[string]*BlockNode // Blocks inside the embed body.
}

// NewEmbedNode returns a EmbedNode.
func NewEmbedNode(tmpl Expr, with Expr, only bool, blocks map[string]*BlockNode, pos Pos) *EmbedNode {
	return &EmbedNode{NewIncludeNode(tmpl, with, only, pos), blocks}
}

// String returns a string representation of an EmbedNode.
func (t *EmbedNode) String() string {
	return fmt.Sprintf("Embed(%s with %s %v: %v)", t.Tpl, t.With, t.Only, t.Blocks)
}

// All returns all the child Nodes in a EmbedNode.
func (t *EmbedNode) All() []Node {
	r := t.IncludeNode.All()
	for _, blk := range t.Blocks {
		r = append(r, blk)
	}
	return r
}

// A UseNode represents the inclusion of blocks from another template.
// It is also possible to specify aliases for the imported blocks to avoid naming conflicts.
//	{% use '::blocks.html.twig' with main as base_main, left as base_left %}
type UseNode struct {
	Pos
	TrimmableNode
	Tpl     Expr              // Evaluates to the name of the template to include.
	Aliases map[string]string // Aliases for included block names, if any.
}

// NewUseNode returns a UseNode.
func NewUseNode(tpl Expr, aliases map[string]string, pos Pos) *UseNode {
	return &UseNode{pos, TrimmableNode{}, tpl, aliases}
}

// String returns a string representation of a UseNode.
func (t *UseNode) String() string {
	if l := len(t.Aliases); l > 0 {
		keys := make([]string, l)
		i := 0
		for orig := range t.Aliases {
			keys[i] = orig
			i++
		}
		sort.Strings(keys)
		res := make([]string, l)
		for i, orig := range keys {
			res[i] = orig + ": " + t.Aliases[orig]
		}
		return fmt.Sprintf("Use(%s with %s)", t.Tpl, strings.Join(res, ", "))
	}
	return fmt.Sprintf("Use(%s)", t.Tpl)
}

// All returns all the child Nodes in a UseNode.
func (t *UseNode) All() []Node {
	return []Node{t.Tpl}
}

// SetNode is a set operation on the given varName.
type SetNode struct {
	Pos
	TrimmableNode
	Name string // Name of the var to set.
	X    Expr   // Value of the var.
}

// NewSetNode returns a SetNode.
func NewSetNode(varName string, expr Expr, pos Pos) *SetNode {
	return &SetNode{pos, TrimmableNode{}, varName, expr}
}

// String returns a string representation of an SetNode.
func (t *SetNode) String() string {
	return fmt.Sprintf("Set(%s = %v)", t.Name, t.X)
}

// All returns all the child Nodes in a SetNode.
func (t *SetNode) All() []Node {
	return []Node{t.X}
}

// DoNode simply executes the expression it contains.
type DoNode struct {
	Pos
	TrimmableNode
	X Expr // The expression to evaluate.
}

// NewDoNode returns a DoNode.
func NewDoNode(expr Expr, pos Pos) *DoNode {
	return &DoNode{pos, TrimmableNode{}, expr}
}

// String returns a string representation of an DoNode.
func (t *DoNode) String() string {
	return fmt.Sprintf("Do(%v)", t.X)
}

// All returns all the child Nodes in a DoNode.
func (t *DoNode) All() []Node {
	return []Node{t.X}
}

// FilterNode represents a block of filtered data.
type FilterNode struct {
	Pos
	TrimmableNode
	Filters []string // Filters to apply to Body.
	Body    Node     // Body of the filter tag.
}

// NewFilterNode creates a FilterNode.
func NewFilterNode(filters []string, body Node, p Pos) *FilterNode {
	return &FilterNode{p, TrimmableNode{}, filters, body}
}

// String returns a string representation of a FilterNode.
func (t *FilterNode) String() string {
	return fmt.Sprintf("Filter (%s): %s", strings.Join(t.Filters, "|"), t.Body)
}

// All returns all the child Nodes in a FilterNode.
func (t *FilterNode) All() []Node {
	return []Node{t.Body}
}

// MacroNode represents a reusable macro.
type MacroNode struct {
	Pos
	TrimmableNode
	Name   string    // Name of the macro.
	Args   []string  // Args the macro receives.
	Body   *BodyNode // Body of the macro.
	Origin string    // The name where this macro is originally defined.
}

// NewMacroNode returns a MacroNode.
func NewMacroNode(name string, args []string, body *BodyNode, p Pos) *MacroNode {
	return &MacroNode{p, TrimmableNode{}, name, args, body, ""}
}

// String returns a string representation of a MacroNode.
func (t *MacroNode) String() string {
	return fmt.Sprintf("Macro %s(%s): %s", t.Name, strings.Join(t.Args, ", "), t.Body)
}

// All returns all the child Nodes in a MacroNode.
func (t *MacroNode) All() []Node {
	return []Node{t.Body}
}

// ImportNode represents importing macros from another template.
type ImportNode struct {
	Pos
	TrimmableNode
	Tpl   Expr   // Evaluates to the name of the template to include.
	Alias string // Name of the var to be used as the base for any macros.
}

// NewImportNode returns a ImportNode.
func NewImportNode(tpl Expr, alias string, p Pos) *ImportNode {
	return &ImportNode{p, TrimmableNode{}, tpl, alias}
}

// String returns a string representation of a ImportNode.
func (t *ImportNode) String() string {
	return fmt.Sprintf("Import (%s as %s)", t.Tpl, t.Alias)
}

// All returns all the child Nodes in a ImportNode.
func (t *ImportNode) All() []Node {
	return []Node{t.Tpl}
}

// FromNode represents an alternative form of importing macros.
type FromNode struct {
	Pos
	TrimmableNode
	Tpl     Expr              // Evaluates to the name of the template to include.
	Imports map[string]string // Imports to fetch from the included template.
}

// NewFromNode returns a FromNode.
func NewFromNode(tpl Expr, imports map[string]string, p Pos) *FromNode {
	return &FromNode{p, TrimmableNode{}, tpl, imports}
}

// String returns a string representation of a FromNode.
func (t *FromNode) String() string {
	keys := make([]string, len(t.Imports))
	i := 0
	for orig := range t.Imports {
		keys[i] = orig
		i++
	}
	sort.Strings(keys)
	res := make([]string, len(t.Imports))
	for i, orig := range keys {
		if orig == t.Imports[orig] {
			res[i] = orig
		} else {
			res[i] = orig + " as " + t.Imports[orig]
		}
	}
	return fmt.Sprintf("From %s import %s", t.Tpl, strings.Join(res, ", "))
}

// All returns all the child Nodes in a FromNode.
func (t *FromNode) All() []Node {
	return []Node{t.Tpl}
}
