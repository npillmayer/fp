package vector

import (
	"fmt"
	"strings"
)

const (
	bits   uint32 = 5 // will produce nodes with degree  2 ^ 5 = 32
	degree uint32 = 1 << bits
	mask   uint32 = degree - 1
)

type props struct {
	bits   uint32 // number of bits to use per level
	degree uint32 // degree is always 2 ^ bits
	mask   uint32 // mask is degree - 1, i.e. a bit pattern with trailing 1s of length 'bits'
	shift  uint32 // we do not store h(v), but rather bits*h(v)

}

func (p props) withShift(shift uint32) props {
	p.shift = shift
	return p
}

type vnode[T any] struct {
	children []*vnode[T]
	leafs    []T
}

func emptyNode[T any](k uint32) *vnode[T] {
	return &vnode[T]{
		children: make([]*vnode[T], int(k)),
	}
}
func newLeaf[T any](tail []T) *vnode[T] {
	l := make([]T, len(tail))
	if tail != nil {
		copy(l, tail)
	}
	return &vnode[T]{leafs: l}
}

func (node vnode[T]) clone(extend bool) *vnode[T] {
	ext := 0
	if extend {
		ext = 1
	}
	n := &vnode[T]{}
	if node.leafs != nil {
		n.leafs = make([]T, len(node.leafs), len(node.leafs)+ext)
		copy(n.leafs, node.leafs)
	}
	if node.children != nil {
		n.children = make([]*vnode[T], len(node.children), len(node.children)+ext)
		copy(n.children, node.children)
	}
	return n
}

func cloneTail[T any](tail []T, l int) []T {
	var newTail []T
	newTail = make([]T, l)
	if tail != nil {
		copy(newTail, tail[:min(l, len(tail))])
	}
	return newTail
}

func newPath[T any](levels, bits, k uint32, tail []T) *vnode[T] {
	topNode := emptyNode[T](k)
	topNode.children[0] = newLeaf(tail)
	for level := levels; level > 0; level -= bits {
		newTop := emptyNode[T](k)
		newTop.children[0] = topNode
		topNode = newTop
	}
	return topNode
}

func (node vnode[T]) String() string {
	b := strings.Builder{}
	b.WriteByte('[')
	if node.leafs != nil {
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

// ---------------------------------------------------------------------------

func assertThat(that bool, msg string, msgargs ...interface{}) {
	if !that {
		msg = fmt.Sprintf("persistent.vector: "+msg, msgargs...)
		panic(msg)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
