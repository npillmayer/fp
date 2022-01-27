package vector

type Vector[T any] struct {
	head       *vnode[T]
	depth      int
	len        int
	bucketSize int
}

// --- API -------------------------------------------------------------------

func (v Vector[T]) Append(x T) Vector[T] {
	var pathBuffer slotPath[T] = make([]slot[T], 0, 4)
	path := v.lastSlot(pathBuffer)
	if path.empty() {
		w := v.shallowClone()
		leafs := leafs[T](v.bucketSize)
		w.head = &leafs
		w.len = 1
		return w
	}
	assertThat(path.last().node.leaf, "last element of non-empty slot path must always be leaf")
	assertThat(path.last().node.leafs != nil, "last element of non-empty slot path must be initialized")
	var w Vector[T]
	if path.last().full() {
		//w, path = v.withAddedBucket(path)
	} else {
		w = v.shallowClone()
	}
	b := path.last()
	b.node.leafs = append(b.node.leafs, x)
	top := path.dropLast().foldR(cloneSeam[T], b)
	w.head = top.node
	w.len++
	return w
}

// ---------------------------------------------------------------------------

func (v Vector[T]) findPath(i int, pathBuf slotPath[T]) (bool, int, slotPath[T]) {
	path := pathBuf[:0]
	if v.head == nil {
		return false, i, path
	}
	n, d, k, j := v.head, v.depth, v.bucketSize, i
	for n != nil && !n.leaf {
		tracer().Debugf("entering inner node %v cap=%d", n, p(k, d))
		pp := p(k, d-1)
		inx := j / pp
		j = j % pp
		path = append(path, slot[T]{inx: inx, node: n})
		n = n.children[inx]
		d--
		tracer().Debugf("loop: pp=%d, j=%d, inx=%d, d=%d, node=%v", pp, j, inx, d, n)
	}
	if n == nil {
		return false, j, path // no entry found
	}
	assertThat(n.leaf, "inconsistency detected: should have initialized leaf")
	path = append(path, slot[T]{inx: j, node: n})
	return true, j, path
}

func (v Vector[T]) adjustCapacity(i int) Vector[T] {
	d, k := v.depth, v.bucketSize
	vcap := p(k, d)   // capacity of v depending on depth d
	h := v.head       // may be nil
	var node vnode[T] // (parent) node to create
	for vcap <= i {
		if h == nil {
			d++
			vcap = p(k, d)
			continue
		}
		if d == 0 {
			node = leafs[T](k)
		} else {
			node = nodes[T](k)
			node.children[0] = h // must always be leftmost child
		}
		h = &node
		d++
		vcap = p(k, d)
		tracer().Debugf("created inner node %v cap=%d", node, vcap)
	}
	w := v.shallowClone()
	w.depth = d
	w.head = h
	return w
}

func (v Vector[T]) location(i int, pathBuf slotPath[T]) (Vector[T], slotPath[T]) {
	w := v.adjustCapacity(i)
	path := pathBuf
	_, j, path := w.findPath(i, nil)
	if len(path) > 0 {
		path = path.clone()
	}
	k, d := v.bucketSize, w.depth-len(path)
	tracer().Debugf("after find: path=%v, d=%d", path, d)
	for d > 1 {
		pp := p(k, d-1)
		inx := j / pp
		j = j % pp
		c := make([]*vnode[T], k)
		node := vnode[T]{children: c}
		path = append(path, slot[T]{inx: inx, node: &node})
		d--
	}
	leaf := leafs[T](k)
	s := slot[T]{inx: j, node: &leaf}
	path = append(path, s)
	w.head = path.dropLast().foldR(chain[T], s).node
	tracer().Debugf("after locate: path=%v, depth=%d", path, w.depth)
	return w, path
}

func (v Vector[T]) shallowClone() Vector[T] {
	return Vector[T]{
		len:        v.len,
		bucketSize: v.bucketSize,
	}
}

func p(a, b int) int {
	if b == 0 {
		return 0
	}
	r := a
	for b > 1 {
		r *= a
		b--
	}
	return r
}
