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

// Children returns all the child Nodes in a BodyNode.
func (l *BodyNode) Children() []Node {
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

// Tpl returns an expression containing the template name.
func (t *UseNode) Tpl() Expr {
	return t.tmpl
}

// Aliases returns a map of the specified block aliases.
func (t *UseNode) Aliases() map[string]string {
	return t.aliases
}
