package node

import "encoding/json"
import "fmt"
import "github.com/stretchr/testify/assert"
import "strings"
import "testing"

type constantValueCleaner struct {
	changed []interface{}
}

func (c *constantValueCleaner) clean(value interface{}) (changed interface{}, err error) {
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
	root.addLeaf("node1.node2.leaf", &constantValueCleaner{})

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
	root.addLeaf("node1.node2.leaf1", &constantValueCleaner{})
	root.addLeaf("node1.node2.leaf2", &constantValueCleaner{})
	root.addLeaf("node1.node3.leaf3", &constantValueCleaner{})

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

func TestCleanRoot(t *testing.T) {
	// given
	root := &node{name: "root"}
	cleaner := &constantValueCleaner{}
	root.addLeaf("node1", cleaner)
	root.addLeaf("node2", cleaner)
	fmt.Printf("children: %+v\n", root.children[1])

	obj := make(map[string]interface{})
	obj["node1"] = "value1"
	obj["node2"] = "value2"
	obj["node3"] = "value3"

	expected := []interface{}{"value1", "value2"}

	// test
	root.Clean(obj)

	// check
	assert.Equal(t, expected, cleaner.changed)
}

func TestCleanOneLevel(t *testing.T) {
	// given
	root := &node{name: "root"}
	cleaner := &constantValueCleaner{}
	root.addLeaf("node1", cleaner)
	root.addLeaf("node2.leaf2", cleaner)

	obj := make(map[string]interface{})
	obj["node1"] = "value1"
	leaf2 := make(map[string]interface{})
	leaf2["leaf2"] = "value2"
	obj["node2"] = leaf2
	obj["node3"] = "value3"

	expected := []interface{}{"value1", "value2"}

	// test
	root.Clean(obj)

	// check
	assert.Equal(t, expected, cleaner.changed)
}

func TestCleanOneLevelWithArray(t *testing.T) {
	// given
	root := &node{name: "root"}
	cleaner := &constantValueCleaner{}
	root.addLeaf("node1", cleaner)
	root.addLeaf("node2.leaf2", cleaner)
	root.addLeaf("node2.leaf4", cleaner)

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
	root.Clean(obj)

	// check
	assert.Equal(t, expected, cleaner.changed)
}

func TestCleanFromConfiguration(t *testing.T) {
	// given
	config := `
	leaf1=constantTransfo
	node2.leaf2 =constantTransfo
	node2.leaf4= constantTransfo
  node3.node31.node311.node3111.leaf32 = constantTransfo
	leaf4=constantTransfo
	node5.node51.node511.leaf5=constantTransfo
	`
	cleaners := make(map[string]ValueCleaner)
	cleaners["constantTransfo"] = &constantValueCleaner{}
	jsonCleaner := NewJsonCleaner(strings.NewReader(config), cleaners)

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
		jsonCleaner.Clean(obj)

		// check
		assert.NotNil(t, jsonCleaner.root)
		assert.Len(t, jsonCleaner.root.children, 5)
		output, _ := json.Marshal(obj)
		assert.JSONEq(t, expected, string(output))
	} else {
		t.Error(err)
	}

}
