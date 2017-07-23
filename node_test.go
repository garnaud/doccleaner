package node

import "encoding/json"
import "fmt"
import "github.com/stretchr/testify/assert"
import "strings"
import "testing"

type constantTransformer struct {
	changed []interface{}
}

func (c *constantTransformer) transform(value interface{}) (changed interface{}, err error) {
	if c.changed == nil {
		c.changed = make([]interface{}, 0)
	}

	fmt.Printf("new change: %+v on existing change %+v\n", value, c.changed)
	c.changed = append(c.changed, value)
	switch value.(type) {
	default:
		return "xxx", nil
	case int, float64:
		return 1234, nil
	}
}

//
func TestAddChild(t *testing.T) {
	n := node{name: "parent"}

	if n.name != "parent" {
		t.Error("name field is not filled")
	}

	child := &node{name: "child"}
	n.addChild(child)

	if len(n.children) != 1 {
		t.Errorf("child is not correctly add to its parent. Expected children 1 but got %d", len(n.children))
	}

}
func TestAddOneLeaf(t *testing.T) {
	fmt.Println("TestAddOneLeaf")
	// given
	root := node{name: "root"}

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
	root := &node{name: "root"}

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
	root := &node{name: "root"}
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
	root.Traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestTraverseOneLevel(t *testing.T) {
	// given
	root := &node{name: "root"}
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
	root.Traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestTraverseOneLevelWithArray(t *testing.T) {
	// given
	root := &node{name: "root"}
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
	root.Traverse(obj)

	// check
	assert.Equal(t, expected, transformer.changed)
}

func TestCompareJson(t *testing.T) {
	// given
	root := &node{name: "root"}
	transformer := &constantTransformer{}
	root.addLeaf("node1", transformer)
	root.addLeaf("node2.leaf2", transformer)
	root.addLeaf("node2.leaf4", transformer)
	root.addLeaf("node3.node31.node311.node3111.leaf32", transformer)
	root.addLeaf("node4", transformer)
	root.addLeaf("node5.node51.node511.leaf5", transformer)

	input := `{
	 "node1":"value1",
	 "node2":[{"leaf2":"value2","leaf3":"value3","leaf4":"value4"}],
	 "node3":[{"node31":[{"node311":{"node3111":[{"leaf31":"abc"},{"leaf32":"cde"}]}}]}],
	 "node4":666,
	 "node5":[{"node51":{"node511":{"leaf5":5111}}}]
  }`
	expected := `{
	 "node1":"xxx",
	 "node2":[{"leaf2":"xxx","leaf3":"value3","leaf4":"xxx"}],
	 "node3":[{"node31":[{"node311":{"node3111":[{"leaf31":"abc"},{"leaf32":"xxx"}]}}]}],
	 "node4":1234,
	 "node5":[{"node51":{"node511":{"leaf5":1234}}}]
  }`

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(input), &obj); err == nil {
		// test
		root.Traverse(obj)

		// check
		output, _ := json.Marshal(obj)
		assert.JSONEq(t, expected, string(output))
	} else {
		t.Error(err)
	}
}

func TestTraverseFromConfiguration(t *testing.T) {
	// given
	config := `
	leaf1=constantTransfo
	node2.leaf2 =constantTransfo
	node2.leaf4= constantTransfo
  node3.node31.node311.node3111.leaf32 = constantTransfo
	leaf4=constantTransfo
	node5.node51.node511.leaf5=constantTransfo
	`
	transformers := make(map[string]Transformer)
	transformers["constantTransfo"] = &constantTransformer{}
	traveler := NewTraveler(strings.NewReader(config), transformers)

	input := `{
	 "leaf1":"value1",
	 "node2":[{"leaf2":"value2","leaf3":"value3","leaf4":"value4"}],
	 "node3":[{"node31":[{"node311":{"node3111":[{"leaf31":"abc"},{"leaf32":"cde"}]}}]}],
	 "leaf4":666,
	 "node5":[{"node51":{"node511":{"leaf5":5111}}}]
  }`
	expected := `{
	 "leaf1":"xxx",
	 "node2":[{"leaf2":"xxx","leaf3":"value3","leaf4":"xxx"}],
	 "node3":[{"node31":[{"node311":{"node3111":[{"leaf31":"abc"},{"leaf32":"xxx"}]}}]}],
	 "leaf4":1234,
	 "node5":[{"node51":{"node511":{"leaf5":1234}}}]
  }`

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(input), &obj); err == nil {
		// test
		traveler.Traverse(obj)

		// check
		assert.NotNil(t, traveler.root)
		assert.Len(t, traveler.root.children, 5)
		output, _ := json.Marshal(obj)
		assert.JSONEq(t, expected, string(output))
	} else {
		t.Error(err)
	}

}
