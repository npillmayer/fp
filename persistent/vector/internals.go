package vector

import (
	"fmt"
	"strings"
)

type bucket[T any] []T

func (b bucket[T]) at(i int) T {
	return b[i]
}

// vnode represents node in the tree a vector is made of. An empty vnode represents
// a bucket of leafs (initially zero).
type vnode[T any] struct {
	leaf     bool
	children []*vnode[T]
	leafs    bucket[T]
}

func nodes[T any](n int) vnode[T] {
	return vnode[T]{
		children: make([]*vnode[T], n),
	}
}

func leafs[T any](n int) vnode[T] {
	return vnode[T]{
		leaf:  true,
		leafs: make([]T, n),
	}
}

func (node vnode[T]) String() string {
	b := strings.Builder{}
	b.WriteByte('[')
	if node.leaf {
		for i, l := range node.leafs {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(fmt.Sprintf("%v", l))
		}
	} else {
		for i, c := range node.children {
			if i > 0 {
				b.WriteByte(',')
			}
			if c == nil {
				b.WriteByte('_')
			} else {
				b.WriteString("▪︎")
			}
		}
	}
	b.WriteByte(']')
	return b.String()
}

func cloneSeam[T any](parent, child slot[T]) slot[T] {
	assertThat(parent.node != nil, "inconsistency: parent of a child is never nil")
	if parent.node == nil {
		return child
	}
	assertThat(!parent.node.leaf, "inconsistency: parent of a child is never a leaf")
	newp := parent.clone()
	newp.node.children[parent.inx] = child.node
	return newp
}

func chain[T any](parent, child slot[T]) slot[T] {
	assertThat(parent.node != nil, "inconsistency: parent of a child is never nil")
	assertThat(!parent.node.leaf, "inconsistency: parent of a child is never a leaf")
	parent.node.children[parent.inx] = child.node
	return parent
}

func (node *vnode[T]) foldLeafs(f func(T, T) T, zero T) T {
	r := zero
	for _, l := range node.leafs {
		r = f(zero, l)
	}
	return r
}

func (v Vector[T]) lastSlot(path slotPath[T]) slotPath[T] {
	path = path[:0]
	n := v.head
	for n != nil {
		if n.leaf {
			l := len(n.leafs)
			path = append(path, slot[T]{inx: l - 1, node: n})
			n = nil
		} else {
			c := n.children
			assertThat(len(c) > 0, "attempt to get last child of uninitialized inner node")
			path = append(path, slot[T]{inx: len(c) - 1})
			n = c[len(c)-1]
		}
	}
	return path
}

func (node *vnode[T]) last() slot[T] {
	assertThat(node != nil, "attempt to get last slot from an uninitialized node")
	if node.leaf {
		return slot[T]{inx: len(node.leafs) - 1, node: node}
	}
	return slot[T]{inx: len(node.children) - 1, node: node}
}

func (s slot[T]) full() bool {
	assertThat(s.node != nil, "node in slot may never be unintialized")
	if s.node.leaf {
		return s.inx == cap(s.node.leafs)-1
	}
	return s.inx == cap(s.node.children)-1
}

// --- Helpers ---------------------------------------------------------------

func assertThat(that bool, msg string, msgargs ...interface{}) {
	if !that {
		msg = fmt.Sprintf("vector: "+msg, msgargs...)
		panic(msg)
	}
}
