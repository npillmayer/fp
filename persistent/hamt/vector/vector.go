package vector

import "fmt"

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

// Degree is an option to set the minimum number of children a node in the tree owns.
// The lower bound for the degree is 3.
//
// Use it like this:
//
//     vec := vector.Immutable[int](BitsPerLevel(5))
//
func BitsPerLevel(n int) Option {
	conf := func(p props) props {
		if n < 0 {
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

// ---------------------------------------------------------------------------

func (v Vector[T]) Get(i int) T {
	assertThat(i >= 0 && uint32(i) < v.length, fmt.Sprintf("vector index out of bounds: %d with length %d", i, v.length))
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
	//ts := v.tailSize()
	//if ts != v.degree {
	if !v.tailFull() { // just append value to tail
		tracer().Debugf("tail not full, appending %v to %v", value, v.tail)
		newTail := cloneTail(v.tail, len(v.tail)+1)
		newTail[len(newTail)-1] = value
		// newTail[ts] = value
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
	// check if the root is completely filled. Must also increment
	// shift if that's the case.
	newRoot := &vnode[T]{}
	//Object[] newRoot;
	s := v.shift
	if (v.length >> v.bits) > (1 << v.shift) {
		s += v.bits
		newRoot = emptyNode[T](v.degree)
		newRoot.children[0] = v.root
		newRoot.children[1] = newPath(v.shift, v.bits, v.degree, v.tail)
		//return new PVec(size+1, newShift, newRoot, newTail);
		v = Vector[T]{length: v.length + 1, props: v.props.withShift(s), root: v.root, tail: newTail}
		return v
	}
	// still space in root
	//newRoot = pushLeaf(shift, size-1, root, tail)
	newRoot = v.pushLeaf(v.length - 1)
	//return new PVec(size+1, shift, newRoot, newTail);
	return Vector[T]{length: v.length + 1, props: v.props, root: newRoot, tail: newTail}
}

func (v Vector[T]) pushLeaf(i uint32) *vnode[T] {
	newRoot := v.root.clone(false)
	node := newRoot
	for level := v.shift; level > 0; level -= bits {
		subidx := (i >> level) & v.mask
		child := node.children[subidx]
		// You could replace this null check with
		// ((tailOffset() - 1) ^ tailOffset() >> level) != 0
		// but we'll still have to assign node[subidx].
		// The null check should therefore be a bit faster.
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

func (v Vector[T]) pop() Vector[T] {
	assertThat(v.length > 0, "attempt to remove item from empty vector")
	if v.length == 1 {
		v = Vector[T]{props: v.props}
		v.shift = 0
		return v
	}
	if ((v.length - 1) & v.mask) > 0 {
		// This one is curious: having int ts_1 = ((size-1) & 31); and using
		// it is slower than using tail.length - 1 and newTail.length!
		//newTail = new Object[tail.length - 1]
		newTail := cloneTail(v.tail, len(v.tail)-1)
		//System.arraycopy(tail, 0, newTail, 0, newTail.length);
		//return new PVec(size-1, shift, root, newTail)
		return Vector[T]{length: v.length - 1, props: v.props, root: v.root, tail: newTail}
	}
	newTrieSize := v.length - v.degree - 1
	// special case: if new size is 32, then new root turns is null, old
	// root the tail
	if newTrieSize == 0 {
		//return new PVec(32, 0, null, root)
		v = Vector[T]{length: v.degree, props: v.props, root: nil, tail: v.root.leafs}
		v.shift = 0
		return v
	}
	// check if we can reduce the trie's height
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
	//return new PVec(size-1, lowerShift, newRoot, node)
	v = Vector[T]{length: v.length - 1, props: v.props, root: newRoot, tail: node.leafs}
	v.shift = lowerShift
	return v
}

func (v Vector[T]) popTrie() Vector[T] {
	newTrieSize := v.length - 33
	// diverges contain information on when the path diverges.
	diverges := newTrieSize ^ (newTrieSize - 1)
	var hasDiverged bool
	newRoot := v.root.clone(false)
	node := newRoot
	for level := v.shift; level > 0; level -= bits {
		subidx := (newTrieSize >> level) & v.mask
		child := node.children[subidx]
		if hasDiverged {
			node = child
		} else if (diverges >> level) != 0 {
			hasDiverged = true
			node.children[subidx] = nil
			node = child
		} else {
			child = child.clone(false)
			node.children[subidx] = child
			node = child
		}
	}
	//return new PVec(size-1, shift, newRoot, node);
	v = Vector[T]{length: v.length - 1, props: v.props, root: newRoot, tail: node.leafs}
	return v
}

func (v Vector[T]) tailOffset() uint32 {
	return (v.length - 1) &^ v.mask
}

func (v Vector[T]) tailSize() uint32 {
	return uint32(len(v.tail))
	// if v.length == 0 {
	// 	return 0
	// }
	// return ((v.length - 1) & v.mask) + 1
}

func (v Vector[T]) tailFull() bool {
	if len(v.tail) < int(v.degree) {
		tracer().Debugf("tail is not full")
		return false
	}
	tracer().Debugf("tail is full")
	return true
}
