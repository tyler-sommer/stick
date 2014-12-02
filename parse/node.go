package parse

import "fmt"

// A node is an item in the AST
type node interface {
	Type() nodeType
	Pos() pos
	String() string
}

type nodeType int

func (t nodeType) Type() nodeType {
	return t
}

type pos int

func (p pos) Pos() pos {
	return p
}

const (
	nodeText nodeType = iota
	nodeModule
	nodePrint
	nodeTag
)

// A list of nodes
type moduleNode struct {
	nodeType
	pos
	nodes []node
}

func newModuleNode() *moduleNode {
	return &moduleNode{nodeModule, pos(0), make([]node, 0)}
}

func (l *moduleNode) append(n node) {
	l.nodes = append(l.nodes, n)
}

func (l *moduleNode) String() string {
	return fmt.Sprintf("Module%s", l.nodes)
}

// A text node
type textNode struct {
	nodeType
	pos
	data []byte
}

func newTextNode(data []byte, p pos) *textNode {
	return &textNode{nodeText, p, data}
}

func (t *textNode) String() string {
	return fmt.Sprintf("Text(%s)", t.data)
}

// A print node
type printNode struct {
	nodeType
	pos
	exp expr
}

func newPrintNode(exp expr, p pos) *printNode {
	return &printNode{nodePrint, p, exp}
}

func (t *printNode) String() string {
	return fmt.Sprintf("Print(%s)", t.exp)
}

// A tag node
type tagNode struct {
	nodeType
	pos
	name *nameExpr
	body node
	attr map[string]expr
}

func newTagNode(name string, body node, attr map[string]expr, p pos) *tagNode {
	return &tagNode{nodeTag, p, newNameExpr(name), body, attr}
}

func (t *tagNode) String() string {
	return fmt.Sprintf("Tag(%s: %s)", t.name, t.body)
}
