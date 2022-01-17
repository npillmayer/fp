package btree

import (
	"fmt"
	"strconv"
	"strings"
)

/*
For various operations on B-trees we manage a path, consisting of an array of slots.
A slot is a tree-node, together with an index into the keys/children held by that node.

Example, given a tree

	Tree(depth=2 ⊥2 ⊤6)
	.
	└── [2,5]
		├── [0,1]
		├── [3,4]
		└── [6,8,9]

To reach key '8' in the rightmost leaf at level 2, we use a path

	[  [2 5]:2   [6 8 9]:1  ]

A slot for a node with n keys may have index ∈ 0 … n
*/

// --- Slot ------------------------------------------------------------------

// slot holds a step of a path.
type slot struct {
	node  *xnode
	index int
}

func (s slot) String() string {
	return strconv.Itoa(s.index) + "@" + s.node.String()
}

// replaceItem replaces the item of a node the slot points to, i.e. the item at s.index
func (s slot) replaceItem(item xitem) xitem {
	assertThat(s.index < len(s.node.items), "internal inconsistency: item index overflow")
	old := s.node.items[s.index]
	s.node.items[s.index] = item
	return old
}

// leftSibling returns the left sibling of an item within a node, or an empty slot.
func (s slot) leftSibling(child slot) slot {
	if s.node == nil || len(s.node.children) == 0 || s.index == 0 {
		return slot{}
	}
	assertThat(s.index <= len(s.node.children), "internal inconsistency: item index overflow")
	lsib := s.node.children[s.index-1]
	tracer().Debugf("left sibling of %s = %s, index in parent is %d", child, lsib, s.index-1)
	return slot{node: lsib, index: len(lsib.items)}
}

// rightSibling returns the right sibling of an item within a node, or an empty slot.
func (s slot) rightSibling(child slot) slot {
	if s.node == nil || len(s.node.children) == 0 || s.index >= len(s.node.children)-1 {
		return slot{}
	}
	rsib := s.node.children[s.index+1]
	tracer().Debugf("right sibling of %s = %s, index in parent is %d", child, rsib, s.index+1)
	return slot{node: rsib, index: len(rsib.items)}
}

// mergeinfo is an ad-hoc tuple for merging tree nodes. It points to the parent node, together
// with its two child nodes to be merged.
type mergeinfo struct {
	parent slot
	left   slot
	right  slot
}

// siblings2 returns child and a sibling (either left or right) as a correctly ordered pair.
// If child is an only child, a pair with an empty right sibling will be returned.
func (s slot) siblings2(child slot) mergeinfo {
	assertThat(!s.node.isLeaf(), "attempt to find siblings for leaf")
	assertThat(s.index < len(s.node.children), "internal inconsistency: child index overflow")
	tracer().Debugf("siblings2: parent %s has %d children", s, len(s.node.children))
	mi := mergeinfo{parent: s}
	sbl := s.leftSibling(child)
	if sbl.node != nil {
		mi.left, mi.right = sbl, child
		mi.parent.index--
	} else { // no left sibling available
		sbl = s.rightSibling(child)
		mi.left, mi.right = child, sbl
	}
	assertThat(mi.left.node != nil, "sibling-pair needs to have non-empty left sibling")
	return mi
}

// item returns the item in a slot.
func (s slot) item() xitem {
	return s.node.items[s.index]
}

// items returns a slice of items contained in s.node. If s is an empty slot (no node
// contained), a valid zero-length slice is returned (i.e., making it safe to call
// `s.items()`` for empty slots).
func (s slot) items() []xitem {
	if s.node == nil {
		return []xitem{}
	}
	return s.node.items
}

func (s slot) len() int {
	if s.node == nil {
		return 0
	}
	return len(s.node.items)
}

// underfull is a convenience-wrapper around s.node.underfull(…)
// It will return true if slot s is empty (i.e., its node is nil).
func (s slot) underfull(lowWaterMark uint) bool {
	if s.node == nil {
		return true
	}
	return s.node.underfull(lowWaterMark)
}

// --- Path ------------------------------------------------------------------

// slotPath is a list of slots, denoting the path to an item.
type slotPath []slot

func (path slotPath) String() string {
	var sb = strings.Builder{}
	sb.WriteRune('[')
	for _, s := range path {
		sb.WriteString(fmt.Sprintf("⟨%s⟩", s))
	}
	sb.WriteRune(']')
	return sb.String()
}

// last returns the last slot of a path.
func (path slotPath) last() slot {
	if len(path) == 0 {
		return slot{}
	}
	return path[len(path)-1]
}

// foldR applies function f on pairs (parent,child) of slots of path.
// Application starts from the right ('R'), which corresponds to the bottom-most item of the path
// (often a leaf of the tree). zero is an element to apply as `child` in the rightmost call
// of f(parent,child). If path is empty, zero will be returned, otherwise the value returned from
// the final call to f will be returned.
func (path slotPath) foldR(f func(slot, slot) slot, zero slot) slot {
	if len(path) == 0 {
		return zero
	}
	r := zero
	for i := len(path) - 1; i >= 0; i-- {
		r = f(path[i], r)
	}
	return r
}

// dropLast returns a slice of a path, omitting the last slot.
func (path slotPath) dropLast() slotPath {
	if len(path) == 0 {
		return path
	}
	return path[:len(path)-1]
}
