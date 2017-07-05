package node

import "fmt"
import "github.com/stretchr/testify/assert"
import "testing"

//
func TestAddChild(t *testing.T) {
	node := Node{name: "parent"}

	if node.name != "parent" {
		t.Error("name field is not filled")
	}

	child := &Node{name: "child"}
	node.addChild(child)

	if len(node.children) != 1 {
		t.Errorf("child is not correctly add to its parent. Expected children 1 but got %d", len(node.children))
	}

}
func TestAddOneLeaf(t *testing.T) {
	fmt.Println("TestAddOneLeaf")
	//given
	root := Node{name: "root"}

	// test
	root.addLeaf("node1.node2.leaf", "%s", "avalue")

	// check
	assert.Len(t, root.children, 1, "root should have one child")
	assert.Equal(t, root.children[0].name, "node1", "root's child has name 'node1'")
	assert.Len(t, root.children[0].children, 1, "root's should have one child")
	assert.Equal(t, root.children[0].children[0].name, "node2", "root's child'schild has name 'node2'")
	assert.False(t, root.children[0].children[0].leaf, "node2 should be a leaf")
	assert.Len(t, root.children[0].children[0].children, 1, "node2 should have one child")
	assert.True(t, root.children[0].children[0].children[0].leaf, "node named 'leaf' should be a leaf")
	assert.Equal(t, root.children[0].children[0].children[0].method, "avalue", "method of the leaf should be 'avalue'")
	assert.Equal(t, root.children[0].children[0].children[0].format, "%s", "format of the leaf should be '%s'")
}

func TestAddLeaves(t *testing.T) {
	fmt.Println("TestAddLeaves")

	//given
	root := &Node{name: "root"}

	// test
	root.addLeaf("node1.node2.leaf1", "%s", "method1")
	root.addLeaf("node1.node2.leaf2", "%s", "method2")
	root.addLeaf("node1.node3.leaf3", "%s", "method3")

	// check
	assert.Len(t, root.children, 1, "root should have one child")
	assert.Equal(t, root.children[0].name, "node1")
	assert.Equal(t, root.children[0].children[0].name, "node2")
	node1 := root.children[0]
	assert.Len(t, node1.children, 2, "node1 should have two children")
	node2 := root.children[0].children[0]
	node3 := root.children[0].children[1]
	assert.Len(t, node2.children, 2, "node2 should have two leaves")
	assert.Len(t, node3.children, 1, "node3 should have one leaf")
}
