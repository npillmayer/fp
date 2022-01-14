package btree

import (
	"fmt"
	"strings"
)

/*
TODOs:

- manage sizes of node.items and node.children correctly
- check for safe handling of empty nodes, especially root
- manage XTree.depth correctly after split or merge
- replace K and T with generic types for Go 1.18
*/

const defaultLowWaterMark uint = 3                              // 2^n - 1
const defaultHighWaterMark uint = (defaultLowWaterMark+1)*2 - 2 // includes space for stopper

type K string      // TODO change to type parameter, ordered
type T interface{} // TODO change to type parameter, comparable

type xitem struct {
	key   K
	value T
}

// ---------------------------------------------------------------------------

type xnode struct {
	items    []xitem
	children []*xnode
}

func (node xnode) clone() xnode {
	return node.cloneWithCapacity(0)
}

func (node xnode) cloneWithCapacity(cap int) xnode {
	itemcnt := len(node.items)
	n := xnode{}
	if itemcnt == 0 && cap <= 0 {
		return n
	}
	if cap <= itemcnt { // there must always be room for itemcnt + 1
		cap = ceiling(itemcnt)
	}
	n.items = make([]xitem, itemcnt, cap)
	copy(n.items, node.items)
	if !node.isLeaf() {
		n.children = make([]*xnode, itemcnt+1, cap)
		copy(n.children, node.children)
	}
	return n
}

// asNonLeaf asserts that a node is not a leaf. Returns a copy with an empty children-slice
// allocated, if none present.
func (node xnode) asNonLeaf() xnode {
	if !node.isLeaf() {
		return node
	}
	return xnode{
		items:    node.items,
		children: make([]*xnode, 0, cap(node.items)),
	}
}

// slice returns node[from:to]. if to == -1, it will be replaced by the length of `node.items`.
func (node xnode) slice(from, to int) xnode {
	if to < 0 {
		to = len(node.items)
	}
	if to-from <= 0 {
		return xnode{}
	}
	size := from - to
	s := xnode{items: make([]xitem, size, ceiling(size))}
	copy(s.items, node.items[from:to])
	if len(node.children) > 0 {
		s.children = make([]*xnode, size, ceiling(size))
		copy(s.children, node.children[from:to])
	}
	return s
}

func (node xnode) String() string {
	if node.items == nil {
		return "[]"
	}
	sb := strings.Builder{}
	sb.WriteRune('[')
	for i, item := range node.items {
		if i > 0 {
			sb.WriteRune(',')
		}
		sb.WriteString(fmt.Sprintf("%v", item.key))
	}
	sb.WriteRune(']')
	return sb.String()
}

func (node xnode) isLeaf() bool {
	return len(node.children) == 0
}

func (node xnode) overfull(highWater uint) bool {
	return len(node.items) > int(highWater)
}

func (node xnode) underfull(lowWater uint) bool {
	return len(node.items) < int(lowWater)
}

// ---------------------------------------------------------------------------

// Tree{} should be a valid btree
type XTree struct {
	root          *xnode // TODO root needs special treatment when cloning!
	depth         uint
	lowWaterMark  uint
	highWaterMark uint
}

func (tree XTree) shallowCloneWithRoot(node xnode) XTree {
	var newTree XTree
	newTree.lowWaterMark, newTree.highWaterMark = tree.lowWaterMark, tree.highWaterMark
	if newTree.lowWaterMark == 0 {
		newTree.lowWaterMark = defaultLowWaterMark
		newTree.highWaterMark = defaultHighWaterMark
	}
	newTree.root = &node
	return newTree
}

// --- API -------------------------------------------------------------------

func (tree XTree) Find(key K) (bool, T) {
	var found bool
	var path slotPath = make([]slot, tree.depth)
	if found, path = tree.findKeyAndPath(key, path); found {
		return true, path.last().item().value
	}
	var none T
	return false, none
}

func (tree XTree) With(key K, value T) (newTree XTree) {
	var path slotPath = make([]slot, tree.depth)
	var found bool
	if found, path = tree.findKeyAndPath(key, path); found {
		if path.last().item().value == value {
			return tree // no need for modification
		}
		return tree.replacing(key, value, path) // otherwise copy with replaced value
	}
	tracer().Debugf("btree.With: slot path = %s", path)
	item := xitem{key, value}
	if tree.root == nil { // virgin tree => insert first node and return
		return tree.shallowCloneWithRoot(xnode{}.withInsertedItem(item, 0))
	}
	leafSlot := path.last()
	assertThat(leafSlot.node.isLeaf(), "attempt to insert item at non-leaf")
	cow := leafSlot.node.withInsertedItem(item, leafSlot.index) // copy-on-write
	tracer().Debugf("created copy of bottom node: %#v", cow)
	newRoot := path.dropLast().foldR(splitAndClone(tree.highWaterMark),
		slot{node: &cow, index: leafSlot.index},
	)
	tracer().Debugf("with: top = %s", newRoot)
	if newRoot.node.overfull(tree.highWaterMark) {
		newRoot = xnode{}.splitChild(newRoot)
	}
	newTree.root = newRoot.node
	return
}

func (tree XTree) WithDeleted(key K) XTree {
	var path slotPath = make([]slot, tree.depth)
	var found bool
	if found, path = tree.findKeyAndPath(key, path); !found {
		return tree // no need for modification
	}
	tracer().Debugf("btree.WithDeleted: slot path = %s", path)
	del := path.last()
	cow := del.node.withDeletedItem(del.index) // copy-on-write
	tracer().Debugf("created copy of node w/out deleted item: %#v", cow)
	newRoot := path.dropLast().foldR(balance(tree.lowWaterMark),
		slot{node: &cow, index: del.index},
	)
	tracer().Debugf("with: top = %s", newRoot)
	switch { // catch border cases where root is empty after deletion
	case newRoot.len() == 0 && newRoot.node.children[0] != nil:
		newRoot = slot{node: newRoot.node.children[0], index: 0}
	case newRoot.len() == 0 && newRoot.node.isLeaf():
		newRoot.node = nil
	}
	return tree.shallowCloneWithRoot(*newRoot.node)
}
