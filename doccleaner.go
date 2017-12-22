package doccleaner

import (
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/davecgh/go-spew/spew"
	"gopkg.in/mgo.v2/bson"
	"io"
	"strings"
	"time"
)

// DocCleaner contains config cleaners
type DocCleaner struct {
	root     *configNode
	cleaners map[string]ValueCleaner
}

// Clean given object
func (docCleaner DocCleaner) Clean(value interface{}) (interface{}, error) {
	if docCleaner.root == nil {
		return nil, errors.New("cleaner without root node. Check the DocCleaner creation")
	}
	return docCleaner.root.clean(value)
}

// methodInfo stores method parameters to use for cleaning
type methodInfo struct {
	MethodName string        `toml:"method"`
	Args       []interface{} `toml:"args"`
	cleaner    ValueCleaner
}

// Clean object to change value
func (methodInfo *methodInfo) Clean(value interface{}) (interface{}, error) {
	return methodInfo.cleaner.Clean(value, methodInfo.Args...)
}

func newMethodInfo(cleaner ValueCleaner) methodInfo {
	return methodInfo{
		cleaner: cleaner,
	}
}

// NewDocCleaner creates a cleaner with default cleaner methods
func NewDocCleaner(configuration io.Reader) (docCleaner *DocCleaner) {
	cleaners := make(map[string]ValueCleaner)
	cleaners["set"] = &Set{}
	cleaners["nil"] = &Nil{}
	cleaners["date"] = &Date{}
	return NewDocCleanerFromConfig(configuration, cleaners)
}

// NewDocCleanerFromConfig creates a cleaner with default cleaners
func NewDocCleanerFromConfig(configuration io.Reader, cleaners map[string]ValueCleaner) (docCleaner *DocCleaner) {
	docCleaner = &DocCleaner{root: &configNode{pathItem: "root"}, cleaners: cleaners}
	var paths map[string]methodInfo
	md, err := toml.DecodeReader(configuration, &paths)
	if err != nil {
		fmt.Printf("metadata:%v\n", md)
		panic(err)
	}
	for path, methodInfo := range paths {
		if cleaner, ok := cleaners[methodInfo.MethodName]; !ok {
			panic(errors.New("can't find method " + methodInfo.MethodName))
		} else {
			methodInfo.cleaner = cleaner
		}
		docCleaner.root.addLeaf(path, methodInfo)
	}
	return docCleaner
}

// configNode for storing each path item. If a configNode is a leaf, it will contain clean method info
type configNode struct {
	pathItem   string
	leaf       bool
	children   []*configNode
	methodInfo methodInfo
}

// ValueCleaner for change a value to an other
type ValueCleaner interface {
	Clean(value interface{}, args ...interface{}) (changed interface{}, err error)
}

// addChild adds a child to the current node
func (parent *configNode) addChild(child *configNode) *configNode {
	if parent.children == nil {
		parent.children = []*configNode{child}
		return child
	}
	if ok, existingChild := parent.hasChild(child.pathItem); ok {
		// nothing todo
		return existingChild
	}
	parent.children = append(parent.children, child)
	return child
}

// hasChild return true if this node as a child with the given name
func (parent *configNode) hasChild(childName string) (ok bool, child *configNode) {
	for _, child = range parent.children {
		if childName == child.pathItem {
			return true, child
		}
	}
	return false, nil
}

// addLeaf adds a leaf in format 'node1.node2.leaf' and with the corresponding cleaner
func (parent *configNode) addLeaf(leaf string, cleaner methodInfo) (n *configNode, err error) {
	nodeNames := strings.Split(leaf, ".")
	if nodeNames == nil || len(nodeNames) == 0 {
		err = errors.New("can't split leaf " + leaf)
		return n, err
	}
	if len(nodeNames) == 1 {
		if ok, _ := parent.hasChild(leaf); !ok {
			n = parent.addChild(&configNode{pathItem: leaf, leaf: true, methodInfo: cleaner})
		}
	} else {
		n = &configNode{pathItem: nodeNames[0], leaf: false}
		n = parent.addChild(n)

		currNode := n
		for _, n := range nodeNames[1:len(nodeNames)] {
			lastNode := &configNode{pathItem: n, leaf: false}
			currNode = currNode.addChild(lastNode)
		}
		currNode.leaf = true
		currNode.methodInfo = cleaner
		n = currNode
	}
	return n, err
}

// Set type allowed to replace current value by a given value
type Set struct {
}

// Set a value method
func (s Set) Clean(value interface{}, args ...interface{}) (changed interface{}, err error) {
	if len(args) == 0 {
		return value, nil
	}
	return args[0], nil
}

// Nil type allowed to replace current value by nil
type Nil struct {
}

// Clean with nil value
func (n Nil) Clean(value interface{}, args ...interface{}) (changed interface{}, err error) {
	return nil, nil
}

// Date type allowed to replace current value by a date
type Date struct {
}

// Clean value by a date. Will use method time.Parse(args[0], args[1]).
func (d Date) Clean(value interface{}, args ...interface{}) (changed interface{}, err error) {
	if len(args) != 2 {
		return nil, nil
	}
	t, err := time.Parse(args[0].(string), args[1].(string))

	return t, err
}

// clean object and apply clean functions on leaves
func (parent *configNode) clean(obj interface{}) (objres interface{}, err error) {
	objres = obj
	switch obj.(type) {
	case []interface{}:
		for i, subobj := range obj.([]interface{}) {
			if subobj, err = parent.clean(subobj); err != nil {
				fmt.Printf("can't clean %+v\n", subobj)
			} else {
				obj.([]interface{})[i] = subobj
			}
		}
	case bson.M:
		objBson := obj.(bson.M)
		for _, child := range parent.children {
			subobj, exists := objBson[child.pathItem]
			if !exists {
				continue
			}
			if child.leaf {
				// leaf case
				objBson[child.pathItem], err = child.methodInfo.Clean(subobj)
			} else {
				newSubObj, _ := child.clean(subobj)
				objBson[child.pathItem] = newSubObj
			}
		}
	case map[string]interface{}:
		objmap := obj.(map[string]interface{})
		for _, child := range parent.children {
			subobj, exists := objmap[child.pathItem]
			if !exists {
				continue
			}
			if child.leaf {
				// leaf case
				objmap[child.pathItem], err = child.methodInfo.Clean(subobj)
			} else {
				child.clean(subobj)
			}
		}
	case []bson.M:
		for i, subobj := range obj.([]bson.M) {
			clean, err := parent.clean(subobj)
			if err != nil {
				spew.Printf("can't clean %#+v\n", subobj)
			} else {
				obj.([]bson.M)[i] = clean.(bson.M)
			}
		}
	default:
		switch len(parent.children) {
		case 0: // TODO nothing to clean
		case 1:
			fmt.Printf("no type: %T\n", obj)
			objres, err = parent.children[0].methodInfo.Clean(obj)
		default: // TODO log problem too many children
		}
	}
	return
}
