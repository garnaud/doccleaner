package node

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Traveler for traverse object and modify values
type Traveler struct {
	root *node
}

// NewTraveler create a new traveler
func NewTraveler(configuration io.Reader, transformers map[string]Transformer) (traveler *Traveler) {
	traveler = &Traveler{root: &node{name: "root"}}

	scanner := bufio.NewScanner(configuration)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		splitted := strings.Split(line, "=")
		traveler.root.addLeaf(strings.TrimSpace(splitted[0]), transformers[strings.TrimSpace(splitted[1])])
	}
	return
}

// Traverse object to change value
func (traveler *Traveler) Traverse(obj map[string]interface{}) (err error) {
	return traveler.root.Traverse(obj)
}

// node for storing path object to transform
type node struct {
	name        string
	leaf        bool
	children    []*node
	format      string
	method      string
	transformer *Transformer
}

// Transformer for change a value to an other
type Transformer interface {
	transform(value interface{}) (changed interface{}, err error)
}

// addChild adds a child to the current node
func (parent *node) addChild(child *node) *node {
	if parent.children == nil {
		parent.children = []*node{child}
		return child
	}
	if ok, existingChild := parent.hasChild(child.name); ok {
		// nothing todo
		return existingChild
	}
	parent.children = append(parent.children, child)
	return child
}

// hasChild return true if this node as a child with the given name
func (parent *node) hasChild(childName string) (ok bool, child *node) {
	for _, child = range parent.children {
		if childName == child.name {
			return true, child
		}
	}
	return false, nil
}

// addLeaf adds a leaf in format 'node1.node2.leaf' and with the corresponding transformer
func (parent *node) addLeaf(leaf string, transformer Transformer) (n *node, err error) {
	nodeNames := strings.Split(leaf, ".")
	if nodeNames == nil || len(nodeNames) == 0 {
		err = errors.New("can't split leaf " + leaf)
		return n, err
	}
	if len(nodeNames) == 1 {
		if ok, _ := parent.hasChild(leaf); !ok {
			n = parent.addChild(&node{name: leaf, leaf: true, transformer: &transformer})
		}
	} else {
		n = &node{name: nodeNames[0], leaf: false}
		n = parent.addChild(n)

		currNode := n
		for _, n := range nodeNames[1:len(nodeNames)] {
			lastNode := &node{name: n, leaf: false}
			currNode = currNode.addChild(lastNode)
		}
		currNode.leaf = true
		currNode.transformer = &transformer
		n = currNode
	}
	return n, err
}

// Traverse object and apply transform functions on leaves
func (parent *node) Traverse(obj map[string]interface{}) (err error) {
	for _, child := range parent.children {
		if value, ok := obj[child.name]; ok {
			if child.leaf {
				obj[child.name], err = (*child.transformer).transform(value)
			} else {
				switch value.(type) {
				default:
					//  TODO logs...
				case map[string]interface{}:
					child.Traverse(value.(map[string]interface{}))
				case []interface{}:
					for _, cvalue := range value.([]interface{}) {
						child.Traverse(cvalue.(map[string]interface{}))
					}
				}
			}
		}
	}
	return err
}
