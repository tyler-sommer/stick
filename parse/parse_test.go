package parse

import "testing"

type parseTest struct {
	name string
	input string
	expected *listNode
}

func mkList(nodes []node) *listNode {
	l := newListNode()
	for _, n := range nodes {
		l.append(n)
	}

	return l
}

var parseTests = []parseTest{
	{"text", "some text", mkList([]node{newTextNode([]byte("some text"), 0)})},
}

func nodeEqual(a, b *listNode) bool {
	if (a.String() != b.String()) {
		return false
	}

	return true
}

func TestParse(t *testing.T) {
	for _, test := range parseTests {
		tree := Parse(test.input)
		if !nodeEqual(tree.root, test.expected) {
			t.Errorf("%s: got\n\t%+v\nexpected\n\t%v", test.name, tree.root, test.expected)
		}
	}
}
