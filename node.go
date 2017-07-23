package node

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Node for storing path object to transform
type Node struct {
	name        string
	leaf        bool
	children    []*Node
	format      string
	method      string
	transformer *Transformer
}

// Transformer for change a value to an other
type Transformer interface {
	transform(value interface{}) (changed interface{}, err error)
}

// addChild adds a child to the current node
func (parent *Node) addChild(child *Node) *Node {
	fmt.Printf("try to add node = %+v\n", child.name)
	if parent.children == nil {
		parent.children = []*Node{child}
		fmt.Printf("add  first child %s to parent %s\n", child.name, parent.name)
		return child
	}
	if ok, existingChild := parent.hasChild(child.name); ok {
		// nothing todo
		fmt.Printf("nothing for adding  %s to %s\n", child.name, parent.name)
		return existingChild
	}
	parent.children = append(parent.children, child)
	fmt.Printf("add child %s to parent %s\n", child.name, parent.name)
	return child
}

// hasChild return true if this node as a child with the given name
func (parent *Node) hasChild(childName string) (ok bool, child *Node) {
	for _, child = range parent.children {
		if childName == child.name {
			return true, child
		}
	}
	return false, nil
}

// addLeaf adds a leaf in format 'node1.node2.leaf' and with the corresponding transformer
func (parent *Node) addLeaf(leaf string, transformer Transformer) (node *Node, err error) {
	nodeNames := strings.Split(leaf, ".")
	if nodeNames == nil || len(nodeNames) == 0 {
		err = errors.New("can't split leaf " + leaf)
		return node, err
	}
	if len(nodeNames) == 1 {
		if ok, _ := parent.hasChild(leaf); !ok {
			node = parent.addChild(&Node{name: leaf, leaf: true, transformer: &transformer})
		}
	} else {
		node = &Node{name: nodeNames[0], leaf: false}
		fmt.Printf("parent = %+v\n", node.name)
		node = parent.addChild(node)

		currNode := node
		for _, n := range nodeNames[1:len(nodeNames)] {
			fmt.Printf("iterate n = %+v\n", n)
			lastNode := &Node{name: n, leaf: false}
			currNode = currNode.addChild(lastNode)
		}
		currNode.leaf = true
		currNode.transformer = &transformer
		node = currNode
		fmt.Printf("add transformer %+v to child %s -> %+v\n", transformer, node.name, node)
	}
	return node, err
}

// traverse object and apply transform functions on leaves
func (parent *Node) traverse(obj map[string]interface{}) (err error) {
	for _, child := range parent.children {
		if value, ok := obj[child.name]; ok {
			if child.leaf {
				obj[child.name], err = (*child.transformer).transform(value)
			} else {
				switch value.(type) {
				default:
					//  TODO logs...
				case map[string]interface{}:
					child.traverse(value.(map[string]interface{}))
				case []interface{}:
					for _, cvalue := range value.([]interface{}) {
						child.traverse(cvalue.(map[string]interface{}))
					}
				}
			}
		}
	}
	return err
}

type Traveler struct {
	root *Node
}

func NewTraveler(configuration io.Reader, transformers map[string]Transformer) (traveler *Traveler) {
	traveler = &Traveler{root: &Node{name: "root"}}

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

func (traveler *Traveler) Traverse(obj map[string]interface{}) (err error) {
	return traveler.root.traverse(obj)
}
