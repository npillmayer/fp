package vector

import (
	"fmt"

	"github.com/npillmayer/fp/maybe"
)

type Vector[T any] struct {
	props
	length uint32
	tail   []T
	root   *vnode[T]
}

func Immutable[T any](opts ...Option) Vector[T] {
	v := Vector[T]{}
	for _, option := range opts {
		v.props = option.config(v.props)
	}
	return v
}

// Option is a type to help initializing vectors at creation time.
type Option struct {
	config func(props) props
}

// DegreeExponent is an option to indirectyl set the degree of the underlying tree for a vector.
// The degree of the tree will be 2^exp. Accepted exponents are [1…5]; default is 3, i.e.
// a degree of 8.
//
// Use it like this:
//
//     vec := vector.Immutable[int](DegreeExponent(5))
//
func DegreeExponent(n int) Option {
	conf := func(p props) props {
		if n <= 0 {
			n = 2
		} else if n > 5 {
			n = 5
		}
		p = props{bits: uint32(n)}
		p.degree = 1 << p.bits
		p.mask = p.degree - 1
		return p
	}
	return Option{config: conf}
}

// --- API -------------------------------------------------------------------

func (v Vector[T]) Len() int {
	return int(v.length)
}

func (v Vector[T]) Last() maybe.Maybe[T] {
	if len(v.tail) == 0 {
		return maybe.Nothing[T]()
	}
	return maybe.Just(v.tail[len(v.tail)-1])
}

func (v Vector[T]) Get(i int) T {
	assertThat(i >= 0 && uint32(i) < v.length, fmt.Sprintf("vector index out of bounds: %d with length %d", i, v.length))
	v.props = v.props.init()
	if uint32(i) >= v.tailOffset() {
		return v.tail[uint32(i)&v.mask]
	}
	node := v.root
	for level := v.shift; level > 0; level -= v.bits {
		node = node.children[(uint32(i)>>level)&v.mask]
	}
	return node.leafs[uint32(i)&v.mask]
}

func (v Vector[T]) Set(i int, value T) Vector[T] {
	assertThat(i >= 0 && uint32(i) < v.length, fmt.Sprintf("vector index out of bounds: %d with length %d", i, v.length))
	v.props = v.props.init()
	if uint32(i) >= v.tailOffset() {
		newTail := cloneTail(v.tail, len(v.tail))
		newTail[uint32(i)&v.mask] = value
		return Vector[T]{length: v.length, props: v.props, root: v.root, tail: newTail}
	}
	newRoot := v.root.clone(false)
	node := newRoot
	for level := v.shift; level > 0; level -= v.bits {
		subidx := (uint32(i) >> level) & v.mask
		child := node.children[subidx]
		child = child.clone(false)
		node.children[subidx] = child
		node = child
	}
	node.leafs[uint32(i)&v.mask] = value
	return Vector[T]{length: v.length, props: v.props, root: newRoot, tail: v.tail}
}

func (v Vector[T]) Push(value T) Vector[T] {
	v.props = v.props.init()
	if !v.tailFull() { // just append value to tail
		tracer().Debugf("tail not full, appending %v to %v", value, v.tail)
		newTail := cloneTail(v.tail, len(v.tail)+1)
		newTail[len(newTail)-1] = value
		return Vector[T]{length: v.length + 1, props: v.props, root: v.root, tail: newTail}
	}
	// tail is full ⇒ have to move tail into tree
	newTail := []T{value}
	assertThat(v.length >= v.degree, "inconsistency: vector.length expected to be > degree")
	if v.length == v.degree { // if old size = degree ⇒ tail becomes new root
		assertThat(v.root == nil, "inconsistency: vector.root expected to be nil")
		leaf := newLeaf(v.tail)
		return Vector[T]{length: v.length + 1, props: v.props.withShift(0), root: leaf, tail: newTail}
	}
	// check for root is full ⇒ increment shift
	newRoot := &vnode[T]{}
	s := v.shift
	if (v.length >> v.bits) > (1 << v.shift) {
		s += v.bits
		newRoot = emptyNode[T](v.degree)
		newRoot.children[0] = v.root
		newRoot.children[1] = newPath(v.shift, v.bits, v.degree, v.tail)
		tracer().Debugf("created new vector tail %v", newTail)
		v = Vector[T]{length: v.length + 1, props: v.props.withShift(s), root: newRoot, tail: newTail}
		return v
	}
	// still space in root
	newRoot = v.pushLeaf(v.length - 1)
	return Vector[T]{length: v.length + 1, props: v.props, root: newRoot, tail: newTail}
}

func (v Vector[T]) pushLeaf(i uint32) *vnode[T] {
	newRoot := v.root.clone(false)
	node := newRoot
	for level := v.shift; level > 0; level -= bits {
		subidx := (i >> level) & v.mask
		child := node.children[subidx]
		if child == nil {
			node.children[subidx] = newPath(level-5, v.bits, v.degree, v.tail)
			return newRoot
		}
		child = child.clone(false)
		node.children[subidx] = child
		node = child
	}
	node.children[(i>>5)&v.mask] = newLeaf(v.tail)
	return newRoot
}

func (v Vector[T]) Pop() Vector[T] {
	assertThat(v.length > 0, "attempt to remove item from empty vector")
	v.props = v.props.init()
	if v.length == 1 {
		v = Vector[T]{props: v.props}
		v.shift = 0
		return v
	}
	if ((v.length - 1) & v.mask) > 0 {
		newTail := cloneTail(v.tail, len(v.tail)-1)
		return Vector[T]{length: v.length - 1, props: v.props, root: v.root, tail: newTail}
	}
	newTrieSize := v.length - v.degree - 1 // new trie size minus length of tail
	if newTrieSize == 0 {                  // root vanishes into tail
		v = Vector[T]{length: v.degree, props: v.props, root: nil, tail: v.root.leafs}
		v.shift = 0
		return v
	}
	if newTrieSize == 1<<v.shift { // can lower the height
		return v.lowerTrie()
	}
	return v.popTrie()
}

func (v Vector[T]) lowerTrie() Vector[T] {
	lowerShift := v.shift - v.bits
	newRoot := v.root.children[0]
	// find new tail
	node := v.root.children[1]
	for level := lowerShift; level > 0; level -= bits {
		node = node.children[0]
	}
	v = Vector[T]{length: v.length - 1, props: v.props, root: newRoot, tail: node.leafs}
	v.shift = lowerShift
	return v
}

func (v Vector[T]) popTrie() Vector[T] {
	newTrieSize := v.length - v.degree - 1
	forkPoint := newTrieSize ^ (newTrieSize - 1) // where does the node-path fork?
	var forked bool
	newRoot := v.root.clone(false)
	node := newRoot
	for level := v.shift; level > 0; level -= bits {
		subidx := (newTrieSize >> level) & v.mask
		child := node.children[subidx]
		switch {
		case forked:
			node = child
		case (forkPoint >> level) != 0:
			forked = true
			node.children[subidx] = nil
			node = child
		default:
			child = child.clone(false)
			node.children[subidx] = child
			node = child
		}
	}
	v = Vector[T]{length: v.length - 1, props: v.props, root: newRoot, tail: node.leafs}
	return v
}

func (v Vector[T]) tailOffset() uint32 {
	return (v.length - 1) &^ v.mask
}

func (v Vector[T]) tailSize() uint32 {
	return uint32(len(v.tail))
}

func (v Vector[T]) tailFull() bool {
	if len(v.tail) < int(v.degree) {
		tracer().Debugf("tail is not full: %v", v.tail)
		return false
	}
	tracer().Debugf("tail is full: %v", v.tail)
	return true
}
