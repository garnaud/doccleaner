package node

import "encoding/json"
import "fmt"
import "github.com/stretchr/testify/assert"
import "testing"

type constantTransformer struct {
	changed []interface{}
}

func (c *constantTransformer) transform(value interface{}) interface{} {
	if c.changed == nil {
		c.changed = make([]interface{}, 0)
	}

	fmt.Printf("new change: %+v on existing change %+v\n", value, c.changed)
	c.changed = append(c.changed, value)
	return "xxx"
}

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
	// given
	root := Node{name: "root"}

	// test
	root.addLeaf("node1.node2.leaf", &constantTransformer{})

	// check
	assert.Len(t, root.children, 1, "root should have one child")
	assert.Equal(t, root.children[0].name, "node1", "root's child has name 'node1'")
	assert.Len(t, root.children[0].children, 1, "root's should have one child")
	assert.Equal(t, root.children[0].children[0].name, "node2", "root's child'schild has name 'node2'")
	assert.False(t, root.children[0].children[0].leaf, "node2 should be a leaf")
	assert.Len(t, root.children[0].children[0].children, 1, "node2 should have one child")
	assert.True(t, root.children[0].children[0].children[0].leaf, "node named 'leaf' should be a leaf")
}

func TestAddLeaves(t *testing.T) {
	fmt.Println("TestAddLeaves")

	// given
	root := &Node{name: "root"}

	// test
	root.addLeaf("node1.node2.leaf1", &constantTransformer{})
	root.addLeaf("node1.node2.leaf2", &constantTransformer{})
	root.addLeaf("node1.node3.leaf3", &constantTransformer{})

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

func TestTraverseRoot(t *testing.T) {
	// given
	root := &Node{name: "root"}
	transformer := &constantTransformer{}
	root.addLeaf("node1", transformer)
	root.addLeaf("node2", transformer)
	fmt.Printf("children: %+v\n", root.children[1])

	obj := make(map[string]interface{})
	obj["node1"] = "value1"
	obj["node2"] = "value2"
	obj["node3"] = "value3"

	expected := []interface{}{"value1", "value2"}

	// test
	root.traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestTraverseOneLevel(t *testing.T) {
	// given
	root := &Node{name: "root"}
	transformer := &constantTransformer{}
	root.addLeaf("node1", transformer)
	root.addLeaf("node2.leaf2", transformer)

	obj := make(map[string]interface{})
	obj["node1"] = "value1"
	leaf2 := make(map[string]interface{})
	leaf2["leaf2"] = "value2"
	obj["node2"] = leaf2
	obj["node3"] = "value3"

	expected := []interface{}{"value1", "value2"}

	// test
	root.traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestTraverseOneLevelWithArray(t *testing.T) {
	// given
	root := &Node{name: "root"}
	transformer := &constantTransformer{}
	root.addLeaf("node1", transformer)
	root.addLeaf("node2.leaf2", transformer)
	root.addLeaf("node2.leaf4", transformer)

	obj := make(map[string]interface{})
	obj["node1"] = "value1"

	leaves1 := make(map[string]interface{})
	leaves1["leaf2"] = "value2"
	leaves1["leaf3"] = "value3"
	leaves2 := make(map[string]interface{})
	leaves2["leaf2"] = "value2"
	leaves2["leaf4"] = "value4"
	leaf2 := []interface{}{leaves1, leaves2}

	obj["node2"] = leaf2
	obj["node3"] = "value3"

	expected := []interface{}{"value1", "value2", "value2", "value4"}

	// test
	root.traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestCompareJson(t *testing.T) {
	// given
	root := &Node{name: "root"}
	transformer := &constantTransformer{}
	root.addLeaf("node1", transformer)
	root.addLeaf("node2.leaf2", transformer)
	root.addLeaf("node2.leaf4", transformer)

	input := `{
	"node1":"value1",
	"node2":[{"leaf2":"value2","leaf3":"value3","leaf4":"value4"}],
	"node3":"value3"
 }`
	expected := `{
	"node1":"xxx",
	"node2":[{"leaf2":"xxx","leaf3":"value3","leaf4":"xxx"}],
	"node3":"value3"
 }`

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(input), &obj); err == nil {
		root.traverse(obj)

		// test
		output, _ := json.Marshal(obj)
		fmt.Println(string(output))
		assert.JSONEq(t, expected, string(output))

	} else {
		t.Error(err)
	}

}
