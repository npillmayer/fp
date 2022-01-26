package vector

import "fmt"

type slot[T any] struct {
	inx  int
	node *vnode[T]
}

func (s slot[T]) String() string {
	return fmt.Sprintf("%d@%s", s.inx, s.node)
}

func (s slot[T]) clone() slot[T] {
	var node *vnode[T]
	if s.node.leaf {
		n := leafs[T](cap(s.node.leafs))
		copy(n.leafs, s.node.leafs)
		node = &n
	} else {
		n := nodes[T](cap(s.node.children))
		copy(n.children, s.node.children)
		node = &n
	}
	return slot[T]{
		inx:  s.inx,
		node: node,
	}
}

// --- Path ------------------------------------------------------------------

// slotPath is a list of slots, denoting the path to a leaf slot.
type slotPath[T any] []slot[T]

func (path slotPath[T]) clone() slotPath[T] {
	var p slotPath[T]
	for _, s := range path {
		p = append(p, s.clone())
	}
	return p
}

func (path slotPath[T]) last() slot[T] {
	if len(path) == 0 {
		return slot[T]{}
	}
	return path[len(path)-1]
}

func (path slotPath[T]) dropLast() slotPath[T] {
	assertThat(!path.empty(), "attempt to drop last slot from empty slot-path")
	path = path[:len(path)-1]
	return path
}

func (path slotPath[T]) empty() bool {
	return len(path) == 0
}

// foldR applies function f on pairs (parent,child) of slots of path.
// Application starts from the right ('R'), which corresponds to the bottom-most item of the path
// (often a leaf of the tree). zero is an element to apply as `child` in the rightmost call
// of f(parent,child). If path is empty, zero will be returned, otherwise the value returned from
// the final call to f will be returned.
func (path slotPath[T]) foldR(f func(slot[T], slot[T]) slot[T], zero slot[T]) slot[T] {
	if len(path) == 0 {
		return zero
	}
	r := zero
	for i := len(path) - 1; i >= 0; i-- {
		r = f(path[i], r)
	}
	return r
}
