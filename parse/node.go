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
	nodeBlock
	nodeIf
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
	data string
}

func newTextNode(data string, p pos) *textNode {
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

// A block node
type blockNode struct {
	nodeType
	pos
	name expr
	body node
}

func newBlockNode(name expr, body node, p pos) *blockNode {
	return &blockNode{nodeBlock, p, name, body}
}

func (t *blockNode) String() string {
	return fmt.Sprintf("Block(%s: %s)", t.name, t.body)
}

// An if node
type ifNode struct {
	nodeType
	pos
	cond expr
	body node
	els  node
}

func newIfNode(cond expr, body node, els node, p pos) *ifNode {
	return &ifNode{nodeIf, p, cond, body, els}
}

func (t *ifNode) String() string {
	return fmt.Sprintf("If(%s: %s Else: %s)", t.cond, t.body, t.els)
}
