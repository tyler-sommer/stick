package parse

import (
	"fmt"
	"sort"
	"strings"
)

// Node is an item in the AST
type Node interface {
	Pos() pos
	String() string
	All() []Node
}

// Type pos is an internal type used to represent the
// position of a token or node.
type pos struct {
	Line   int
	Offset int
}

func newPos(line, offset int) pos {
	return pos{line, offset}
}

// Pos returns a pos.
func (p pos) Pos() pos {
	return p
}

// String returns a string representation of a pos.
func (p pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Offset)
}

// ModuleNode represents a root node in the AST.
type ModuleNode struct {
	*BodyNode
	parent *ExtendsNode
}

func newModuleNode(nodes ...Node) *ModuleNode {
	return &ModuleNode{newBodyNode(newPos(1, 0), nodes...), nil}
}

// Parent returns the ModuleNode's parent reference, in the form of an ExtendsNode.
func (l *ModuleNode) Parent() *ExtendsNode {
	return l.parent
}

// String returns a string representation of a ModuleNode.
func (l *ModuleNode) String() string {
	return fmt.Sprintf("Module%s", l.nodes)
}

// BodyNode represents a list of nodes.
type BodyNode struct {
	pos
	nodes []Node
}

func newBodyNode(pos pos, nodes ...Node) *BodyNode {
	return &BodyNode{pos, nodes}
}

func (l *BodyNode) append(n Node) {
	l.nodes = append(l.nodes, n)
}

// String returns a string representation of a BodyNode.
func (l *BodyNode) String() string {
	return fmt.Sprintf("Body%s", l.nodes)
}

// All returns all the child Nodes in a BodyNode.
func (l *BodyNode) All() []Node {
	return l.nodes
}

// TextNode represents raw, non Stick source code, like plain HTML.
type TextNode struct {
	pos
	data string
}

func newTextNode(data string, p pos) *TextNode {
	return &TextNode{p, data}
}

// String returns a string representation of a TextNode.
func (t *TextNode) String() string {
	return fmt.Sprintf("Text(%s)", t.data)
}

// Text returns the raw text stored in a TextNode.
func (t *TextNode) Text() string {
	return t.data
}

// All returns all the child Nodes in a TextNode.
func (t *TextNode) All() []Node {
	return []Node{}
}

// PrintNode represents a print statement
type PrintNode struct {
	pos
	exp Expr
}

func newPrintNode(exp Expr, p pos) *PrintNode {
	return &PrintNode{p, exp}
}

// Expr returns the expression to be evaluated and printed.
func (t *PrintNode) Expr() Expr {
	return t.exp
}

// String returns a string representation of a PrintNode.
func (t *PrintNode) String() string {
	return fmt.Sprintf("Print(%s)", t.exp)
}

// All returns all the child Nodes in a PrintNode.
func (t *PrintNode) All() []Node {
	return []Node{t.exp}
}

// BlockNode represents a block statement
type BlockNode struct {
	pos
	name string
	body Node
}

func newBlockNode(name string, body Node, p pos) *BlockNode {
	return &BlockNode{p, name, body}
}

// Name returns the name of the BlockNode.
func (t *BlockNode) Name() string {
	return t.name
}

// Body returns the body Node of a BlockNode.
func (t *BlockNode) Body() Node {
	return t.body
}

// String returns a string representation of a BlockNode.
func (t *BlockNode) String() string {
	return fmt.Sprintf("Block(%s: %s)", t.name, t.body)
}

// All returns all the child Nodes in a BlockNode.
func (t *BlockNode) All() []Node {
	return []Node{t.body}
}

// IfNode represents an if statement
type IfNode struct {
	pos
	cond Expr
	body Node
	els  Node
}

func newIfNode(cond Expr, body Node, els Node, p pos) *IfNode {
	return &IfNode{p, cond, body, els}
}

// Cond returns the conditional expression to be evaluated in the IfNode.
func (t *IfNode) Cond() Expr {
	return t.cond
}

// Body returns the body of the "true" branch of an IfNode.
func (t *IfNode) Body() Node {
	return t.body
}

// Else returns the body of the "false" branch of an IfNode.
func (t *IfNode) Else() Node {
	return t.els
}

// String returns a string representation of an IfNode.
func (t *IfNode) String() string {
	return fmt.Sprintf("If(%s: %s Else: %s)", t.cond, t.body, t.els)
}

// All returns all the child Nodes in a IfNode.
func (t *IfNode) All() []Node {
	return []Node{t.cond, t.body, t.els}
}

// ExtendsNode represents an extends statement
type ExtendsNode struct {
	pos
	tplRef Expr
}

func newExtendsNode(tplRef Expr, p pos) *ExtendsNode {
	return &ExtendsNode{p, tplRef}
}

// TemplateRef returns the expression containing a template name.
func (t *ExtendsNode) TemplateRef() Expr {
	return t.tplRef
}

// String returns a string representation of an ExtendsNode.
func (t *ExtendsNode) String() string {
	return fmt.Sprintf("Extends(%s)", t.tplRef)
}

// All returns all the child Nodes in a ExtendsNode.
func (t *ExtendsNode) All() []Node {
	return []Node{t.tplRef}
}

// ForNode represents a for loop construct.
type ForNode struct {
	pos
	key  string
	val  string
	expr Expr
	body Node
	els  Node
}

func newForNode(k, v string, expr Expr, body, els Node, p pos) *ForNode {
	return &ForNode{p, k, v, expr, body, els}
}

// String returns a string representation of a ForNode.
func (t *ForNode) String() string {
	return fmt.Sprintf("For(%s, %s in %s: %s else %s)", t.key, t.val, t.expr, t.body, t.els)
}

// All returns all the child Nodes in a ForNode.
func (t *ForNode) All() []Node {
	return []Node{t.expr, t.body, t.els}
}

// Key returns the name of the key variable during iteration. Key will be an empty
// string if no key should be set.
func (t *ForNode) Key() string {
	return t.key
}

// Val returns the name of the value variable during iteration.
func (t *ForNode) Val() string {
	return t.val
}

// Expr returns the expression to be evaluated and iterated over.
func (t *ForNode) Expr() Expr {
	return t.expr
}

// Body returns the body Node to be used during iteration.
func (t *ForNode) Body() Node {
	return t.body
}

// Else returns the Node to be used if no iteration was performed.
func (t *ForNode) Else() Node {
	return t.els
}

// IncludeNode is an include statement.
type IncludeNode struct {
	pos
	tmpl Expr
	with Expr
	only bool
}

func newIncludeNode(tmpl Expr, with Expr, only bool, pos pos) *IncludeNode {
	return &IncludeNode{pos, tmpl, with, only}
}

// String returns a string representation of an IncludeNode.
func (t *IncludeNode) String() string {
	return fmt.Sprintf("Include(%s with %s %v)", t.tmpl, t.with, t.only)
}

// All returns all the child Nodes in a IncludeNode.
func (t *IncludeNode) All() []Node {
	return []Node{t.tmpl, t.with}
}

// Tpl returns an expression containing a template name.
func (t *IncludeNode) Tpl() Expr {
	return t.tmpl
}

// With returns an expression to be evaluated and used as the
// included template's context.
func (t *IncludeNode) With() Expr {
	return t.with
}

// Only returns true if only the context specified in the IncludeNode's
// With expression should be available to the included template.
func (t *IncludeNode) Only() bool {
	return t.only
}

// EmbedNode is a special include statement.
type EmbedNode struct {
	*IncludeNode
	blockRefs map[string]*BlockNode
}

func newEmbedNode(tmpl Expr, with Expr, only bool, blocks map[string]*BlockNode, pos pos) *EmbedNode {
	return &EmbedNode{newIncludeNode(tmpl, with, only, pos), blocks}
}

// String returns a string representation of an EmbedNode.
func (t *EmbedNode) String() string {
	return fmt.Sprintf("Embed(%s with %s %v: %v)", t.tmpl, t.with, t.only, t.blockRefs)
}

// All returns all the child Nodes in a EmbedNode.
func (t *EmbedNode) All() []Node {
	r := t.IncludeNode.All()
	for _, blk := range t.blockRefs {
		r = append(r, blk)
	}
	return r
}

// Blocks returns all blocks to be used when embedding the template.
func (t *EmbedNode) Blocks() map[string]*BlockNode {
	return t.blockRefs
}

// A UseNode represents the inclusion of blocks from another template.
// It is also possible to specify aliases for the imported blocks to avoid naming conflicts.
//	{% use '::blocks.html.twig' with main as base_main, left as base_left %}
type UseNode struct {
	pos
	tmpl    Expr
	aliases map[string]string
}

func newUseNode(tmpl Expr, aliases map[string]string, pos pos) *UseNode {
	return &UseNode{pos, tmpl, aliases}
}

// String returns a string representation of a UseNode.
func (t *UseNode) String() string {
	if l := len(t.aliases); l > 0 {
		keys := make([]string, l)
		i := 0
		for orig := range t.aliases {
			keys[i] = orig
			i++
		}
		sort.Strings(keys)
		res := make([]string, l)
		for i, orig := range keys {
			res[i] = orig + ": " + t.aliases[orig]
		}
		return fmt.Sprintf("Use(%s with %s)", t.tmpl, strings.Join(res, ", "))
	}
	return fmt.Sprintf("Use(%s)", t.tmpl)
}

// All returns all the child Nodes in a UseNode.
func (t *UseNode) All() []Node {
	return []Node{t.tmpl}
}

// Tpl returns an expression containing the template name.
func (t *UseNode) Tpl() Expr {
	return t.tmpl
}

// Aliases returns a map of the specified block aliases.
func (t *UseNode) Aliases() map[string]string {
	return t.aliases
}

// SetNode is a set operation on the given varName.
type SetNode struct {
	pos
	varName string
	expr    Expr
}

func newSetNode(varName string, expr Expr, pos pos) *SetNode {
	return &SetNode{pos, varName, expr}
}

// VarName returns the name of the variable to set.
func (t *SetNode) VarName() string {
	return t.varName
}

// Expr returns the right-hand expression in the set statement.
func (t *SetNode) Expr() Expr {
	return t.expr
}

// String returns a string representation of an SetNode.
func (t *SetNode) String() string {
	return fmt.Sprintf("Set(%s = %v)", t.varName, t.expr)
}

// All returns all the child Nodes in a SetNode.
func (t *SetNode) All() []Node {
	return []Node{t.expr}
}

// DoNode simply executes the expression it contains.
type DoNode struct {
	pos
	expr Expr
}

func newDoNode(expr Expr, pos pos) *DoNode {
	return &DoNode{pos, expr}
}

// Expr returns the expression in the do statement.
func (t *DoNode) Expr() Expr {
	return t.expr
}

// String returns a string representation of an DoNode.
func (t *DoNode) String() string {
	return fmt.Sprintf("Do(%v)", t.expr)
}

// All returns all the child Nodes in a DoNode.
func (t *DoNode) All() []Node {
	return []Node{t.expr}
}

// FilterNode represents a block of filtered data.
type FilterNode struct {
	pos
	filters []string
	body    Node
}

func newFilterNode(filters []string, body Node, p pos) *FilterNode {
	return &FilterNode{p, filters, body}
}

// String returns a string representation of a FilterNode.
func (t *FilterNode) String() string {
	return fmt.Sprintf("Filter (%s): %s", strings.Join(t.filters, "|"), t.body)
}

// All returns all the child Nodes in a FilterNode.
func (t *FilterNode) All() []Node {
	return []Node{t.body}
}

func (t *FilterNode) Filters() []string {
	return t.filters
}

// Body returns the body Node to be filtered.
func (t *FilterNode) Body() Node {
	return t.body
}
