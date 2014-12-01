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
	nodeList
)

// A list of nodes
type listNode struct {
	nodeType
	pos
	nodes []node
}

func newListNode() *listNode {
	return &listNode{nodeList, pos(0), make([]node, 0)}
}

func (l *listNode) append(n node) {
	l.nodes = append(l.nodes, n)
}

func (l *listNode) String() string {
	return fmt.Sprintf("List(%s)", l.nodes)
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

func (*textNode) Children() (children []node) {
	children = make([]node, 0)

	return
}
