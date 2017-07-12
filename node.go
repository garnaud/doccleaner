package node

import (
	"errors"
	"fmt"
	"strings"
)

type Node struct {
	name        string
	leaf        bool
	children    []*Node
	format      string
	method      string
	transformer *Transformer
}

type Transformer interface {
	transform(value interface{}) interface{}
}

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

func (parent *Node) hasChild(childName string) (ok bool, child *Node) {
	for _, child = range parent.children {
		if childName == child.name {
			return true, child
		}
	}
	return false, nil
}

func (parent *Node) addLeaf(leaf string, transformer Transformer) (node *Node, err error) {
	nodeNames := strings.Split(leaf, ".")
	if nodeNames == nil || len(nodeNames) == 0 {
		err = errors.New("can't split leaf " + leaf)
		return node, err
	}
	if len(nodeNames) == 1 {
		if ok, _ := parent.hasChild(leaf); !ok {
			parent.addChild(&Node{name: leaf, leaf: true})
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
		node.leaf = true
		node.transformer = &transformer
		fmt.Printf("add transformer %+v to child %s -> %+v\n", transformer, node.name, node)
	}
	return node, err
}

func (parent *Node) traverse(obj map[string]interface{}) (result string, err error) {
	result = ""
	for _, child := range parent.children {
		fmt.Printf("child: %+v\n", child)
		if value, ok := obj[child.name]; ok {
			if child.leaf {
				if child.transformer == nil {
					panic(errors.New("transformer nil: " + child.name))
				}
				v := (*child.transformer).transform(value)
				fmt.Printf("v = %+v\n", v)
			}
		}
	}
	return result, nil
}
