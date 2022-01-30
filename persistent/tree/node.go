package tree

/*
License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2022 Norbert Pillmayer <norbert@pillmayer.com>

*/

import (
	"fmt"
)

/*
We manage a tree of mutable nodes. Each nodes carries a payload of type parameter T.
Nodes maintain a slice of children.

In the future, we may move to immutable nodes to reduce lock contention, but first let's get
some experience with this one.
*/

// Node is the base type our tree is built of.
type Node[T comparable] struct {
	Payload  T        // nodes may carry a payload of arbitrary type
	parent   *Node[T] // parent node of this node
	children chvec[T] // children nodes
	Rank     uint32   // rank is used for preserving sequence
}

// NewNode creates a new tree node with a given payload.
func NewNode[T comparable](payload T) Node[T] {
	return Node[T]{Payload: payload}
}

func (node Node[T]) clone(children []*Node[T], transient bool) Node[T] {
	return Node[T]{
		Payload:  node.Payload,
		parent:   node.parent,
		Rank:     node.Rank,
		children: node.children.clone(),
	}
}

func (node Node[T]) String() string {
	return fmt.Sprintf("(Node #ch=%d %v)", node.ChildCount(), node.Payload)
}

// AddChild inserts a new child node into the tree.
// The newly inserted node is connected to this node as its parent.
// It returns the parent node to allow for chaining.
//
// This operation is concurrency-safe.
func (node *Node[T]) AddChild(ch *Node[T]) *Node[T] {
	return node.add(ch, false)
}

func (node *Node[T]) add(ch *Node[T], transient bool) *Node[T] {
	var n *Node[T] = node
	if !transient {
		newnode := node.clone(node.children, false)
		n = &newnode
	}
	node.children = n.children.appendChild(ch)
	if ch != nil {
		n.children[len(node.children)-1].parent = node
	}
	return node
}

// SetChildAt inserts a new child node into the tree.
// The newly inserted node is connected to this node as its parent.
// The child is set at a given position in relation to other children,
// replacing the child at position i if it exists.
// It returns the parent node to allow for chaining.
//
// This operation is concurrency-safe.
//
func (node *Node[T]) ReplaceChild(i int, ch *Node[T], transient bool) *Node[T] {
	return node.replaceChild(i, ch, false)
}

func (node *Node[T]) replaceChild(i int, ch *Node[T], transient bool) *Node[T] {
	var n *Node[T] = node
	if !transient {
		newnode := node.clone(node.children, false)
		n = &newnode
	}
	n.children = n.children.replaceChild(i, ch)
	if ch != nil {
		n.children[i].parent = node
	}
	return n
}

// InsertChild creates a new node wich contains a new child ch.
// if the newly inserted node is non-nil, it is connected to this node as its parent.
// The child is set at a given position in relation to other children,
// shifting children at later positions.
// It returns the parent node to allow for chaining.
//
// This operation is concurrency-safe.
//
func (node *Node[T]) InsertChild(i int, ch *Node[T]) *Node[T] {
	return node.insertChild(i, ch, false)
}

func (node *Node[T]) insertChild(i int, ch *Node[T], transient bool) *Node[T] {
	var n *Node[T] = node
	if !transient {
		newnode := node.clone(node.children, false)
		n = &newnode
	}
	n.children = n.children.insertChildAt(i, ch)
	if ch != nil {
		n.children[i].parent = node
	}
	return n
}

// Parent returns the parent node or nil (for the root of the tree).
func (node *Node[T]) Parent() *Node[T] {
	return node.parent
}

// Isolate removes a node from its parent.
// Isolate returns the isolated node.
func (node *Node[T]) Isolate() *Node[T] {
	if node != nil && node.parent != nil {
		node.parent.children.remove(node)
	}
	return node
}

// ChildCount returns the number of children-nodes for a node
// (concurrency-safe).
func (node *Node[T]) ChildCount() int {
	return node.children.length()
}

// Child is a concurrency-safe way to get a children-node of a node.
func (node *Node[T]) Child(n int) (*Node[T], bool) {
	if n < 0 || node.children.length() <= n {
		return nil, false
	}
	ch := node.children.child(n)
	return ch, ch != nil
}

// Children returns a slice with all children of a node.
// If omitNilChildren is set, empty children aren't included in the slice
func (node *Node[T]) Children(omitNilChildren bool) []*Node[T] {
	return node.children.asSlice(omitNilChildren)
}

// IndexOfChild returns the index of a child within the list of children
// of its parent. ch may not be nil.
func (node *Node[T]) IndexOfChild(ch *Node[T]) int {
	if node.ChildCount() > 0 {
		children := node.Children(false)
		for i, child := range children {
			if ch == child {
				return i
			}
		}
	}
	return -1
}

// --- Slices of concurrency-safe sets of children ----------------------

type chvec[T comparable] []*Node[T]

func (chs chvec[T]) clone() chvec[T] {
	c := make([]*Node[T], len(chs))
	copy(c, chs)
	return c
}

// length returns the number of children slots (including empty slots).
func (chs chvec[T]) length() int {
	return len(chs)
}

// count returns the number of non-nil children.
func (chs chvec[T]) count() int {
	var n int
	for _, child := range chs {
		if child != nil {
			n++
		}
	}
	return n
}

func (chs chvec[T]) appendChild(child *Node[T]) chvec[T] {
	chs = append(chs, child)
	return chs
}

func (chs chvec[T]) replaceChild(i int, child *Node[T]) chvec[T] {
	if chs.length() <= i { // make room for child at i
		l := chs.length()
		chs = append(chs, make([]*Node[T], i-l+1)...)
	}
	chs[i] = child
	return chs
}

func (chs chvec[T]) insertChildAt(i int, child *Node[T]) chvec[T] {
	if chs.length() <= i {
		l := chs.length()
		chs = append(chs, make([]*Node[T], i-l+1)...)
	} else {
		chs = append(chs, nil)   // make room for one child
		copy(chs[i+1:], chs[i:]) // shift i+1..n
	}
	chs[i] = child
	return chs
}

func (chs chvec[T]) remove(node *Node[T]) {
	for i, ch := range chs {
		if ch == node {
			chs[i] = nil
			break
		}
	}
}

func (chs chvec[T]) child(n int) *Node[T] {
	if chs.length() == 0 || n < 0 || n >= chs.length() {
		return nil
	}
	return chs[n]
}

func (chs chvec[T]) compact() []*Node[T] {
	children := make([]*Node[T], 0, chs.count())
	for _, ch := range chs {
		if ch != nil {
			children = append(children, ch)
		}
	}
	return children
}
