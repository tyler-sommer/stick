package parse

import "fmt"

// A node is an item in the AST
type Node interface {
	Type() nodeType
	Pos() pos
	String() string
}

type nodeType int

func (t nodeType) Type() nodeType {
	return t
}

type pos struct {
	Line int
	Offset int
}

func newPos(line, offset int) pos {
	return pos{line, offset}
}

func (p pos) Pos() pos {
	return p
}

func (p pos) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Offset)
}

const (
	nodeText nodeType = iota
	nodeModule
	nodePrint
	nodeBlock
	nodeIf
)

// ModuleNode represents a set of nodes.
type ModuleNode struct {
	nodeType
	pos
	parent *Node
	nodes  []Node
}

func newModuleNode() *ModuleNode {
	return &ModuleNode{nodeModule, newPos(1, 0), nil, make([]Node, 0)}
}

func (l *ModuleNode) append(n Node) {
	l.nodes = append(l.nodes, n)
}

func (l *ModuleNode) String() string {
	return fmt.Sprintf("Module%s", l.nodes)
}

func (l *ModuleNode) Children() []Node {
	return l.nodes
}

// TextNode represents raw, non Stick source code, like plain HTML.
type TextNode struct {
	nodeType
	pos
	data string
}

func newTextNode(data string, p pos) *TextNode {
	return &TextNode{nodeText, p, data}
}

func (t *TextNode) String() string {
	return fmt.Sprintf("Text(%s)", t.data)
}

func (t *TextNode) Text() string {
	return t.data
}

// PrintNode represents a print statement
type PrintNode struct {
	nodeType
	pos
	exp expr
}

func newPrintNode(exp expr, p pos) *PrintNode {
	return &PrintNode{nodePrint, p, exp}
}

func (t *PrintNode) String() string {
	return fmt.Sprintf("Print(%s)", t.exp)
}

// BlockNode represents a block statement
type BlockNode struct {
	nodeType
	pos
	name string
	body Node
}

func newBlockNode(name string, body Node, p pos) *BlockNode {
	return &BlockNode{nodeBlock, p, name, body}
}

func (t *BlockNode) String() string {
	return fmt.Sprintf("Block(%s: %s)", t.name, t.body)
}

// IfNode represents an if statement
type IfNode struct {
	nodeType
	pos
	cond expr
	body Node
	els  Node
}

func newIfNode(cond expr, body Node, els Node, p pos) *IfNode {
	return &IfNode{nodeIf, p, cond, body, els}
}

func (t *IfNode) String() string {
	return fmt.Sprintf("If(%s: %s Else: %s)", t.cond, t.body, t.els)
}
