package btree

import (
	"fmt"
	"sort"
	"strings"
)

/*
Remarks:
--------

- 'cow' stands for copy-on-write and is used throughout the code for variables holding clones of nodes.

- We use a programming-style reminiscent of functional programming (see remarks on
  re-balancing) where it makes things easier to understand.

- A new modified incarnation of a tree always is reflected by a new tree.root.

*/

type K int         // TODO for generics: change to type parameter, ordered
type T interface{} // TODO for generics: change to type parameter, comparable

// xitem is a type for entries of the tree.
type xitem struct {
	key   K
	value T
}

// xnode is a type for tree nodes, either an internal node or a leaf.
// For leafs, children will be nil.
type xnode struct {
	items    []xitem
	children []*xnode
}

// --- Tree ------------------------------------------------------------------

func (tree Tree) shallowCloneWithRoot(node xnode) Tree {
	var newTree Tree
	newTree.depth = tree.depth
	newTree.lowWaterMark, newTree.highWaterMark = tree.lowWaterMark, tree.highWaterMark
	if newTree.lowWaterMark == 0 {
		newTree.lowWaterMark = defaultLowWaterMark
		newTree.highWaterMark = defaultHighWaterMark
	}
	newTree.root = &node
	return newTree
}

func (tree Tree) withDepth(d uint) Tree {
	t := tree.shallowCloneWithRoot(xnode{})
	t.root = tree.root
	t.depth = d
	return t
}

func (tree Tree) findKeyAndPath(key K, pathBuf slotPath) (found bool, path slotPath) {
	path = pathBuf[:0] // we track the path to the key's slot
	if tree.root == nil {
		return
	}
	var index int
	var node *xnode = tree.root // walking nodes, start search at the top
	for !node.isLeaf() {
		tracer().Debugf("finding inner node = %v", node)
		found, index = node.findSlot(key)
		path = append(path, slot{node: node, index: index})
		if found {
			return // we have an exact match
		}
		node = node.children[index]
	}
	tracer().Debugf("finding leaf node %v", node)
	found, index = node.findSlot(key)
	path = append(path, slot{node: node, index: index})
	tracer().Debugf("slot path for key = %v -> %s", key, path)
	return
}

func (tree Tree) replacing(key K, value T, path slotPath) (newTree Tree) {
	assertThat(len(path) > 0, "cannot replace item without path")
	tracer().Debugf("replace: slot path = %s", path)
	hit := path[len(path)-1] // slot where `key` lives
	item := xitem{key: key, value: value}
	cow := hit.node.withReplacedValue(item, hit.index)
	tracer().Debugf("created copy of node for replacement: %#v", cow)
	newRoot := path.dropLast().foldR(cloneSeam, slot{node: &cow, index: hit.index})
	tracer().Debugf("replace: top = %s", newRoot)
	newTree.root = newRoot.node
	return
}

// --- Node ------------------------------------------------------------------

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

// A node holds a list of keys/value-pairs, called items. Internal nodes additionally hold a list
// of children. For every modification a copy of a node is returned, modifications being
//
// - replacement of value
// - deletion of item
// - insertion of item
// - left- or right-most item cut off
//
// All of these functions are straightforward.

// withReplacedValue replaces the value at index `at` with item.value.
// It returns a cloned node, containing the new value.
func (node xnode) withReplacedValue(item xitem, at int) xnode {
	assertThat(at <= len(node.items), "given item index out of range: %d < %d", len(node.items), at)
	cow := node.clone()
	assertThat(item.key == cow.items[at].key, "attempt to replace value for different key")
	cow.items[at].value = item.value
	return cow
}

// withDeletedItem returns a clone of node, where the item at index `at` has been removed.
func (node xnode) withDeletedItem(at int) xnode {
	assertThat(at <= len(node.items), "given item index out of range: %d < %d", len(node.items), at)
	tracer().Debugf("deletion in node %s at %d", node, at)
	cow := node.clone()
	cow.items = append(cow.items[:at], cow.items[at+1:]...)
	if !cow.isLeaf() {
		cow.children = append(cow.children[:at], cow.children[at+1:]...)
	}
	tracer().Debugf("after node.delete(%v): len=%d, cap=%d -> %s", node.items[at].key,
		len(cow.items), cap(cow.items), cow)
	return cow
}

// withInserteditem returns a clone of node, where a new item at index `at` has been inserted.
func (node xnode) withInsertedItem(item xitem, at int) xnode {
	assertThat(at <= len(node.items), "given item index out of range: %d < %d", len(node.items), at)
	cap := max(at+1, len(node.items)+1)
	cow := node.cloneWithCapacity(cap) // copy-on-write behaviour requires cloning
	if at == len(node.items) {         // append at the end
		cow.items = append(cow.items, item)
		if !cow.isLeaf() {
			cow.children = append(cow.children, nil) // append placeholder
		}
		return cow
	}
	cow.items = append(cow.items[:at], item)
	cow.items = append(cow.items, node.items[at:]...)
	if !cow.isLeaf() {
		cow.children = append(cow.children[:at+1], nil) // insert placeholder
		cow.children = append(cow.children, node.children[at:]...)
	}
	return cow
}

// withCutRight returns a clone of node, with the rightmost item cut off.
// If the node is a inner node, the rightmost child is cut off, too.
func (node xnode) withCutRight() (xnode, xitem, *xnode) {
	assertThat(len(node.items) > 0, "attempt to cut right item from empty node")
	cow := node.clone()
	item := cow.items[len(cow.items)-1]
	cow.items = cow.items[:len(cow.items)-1]
	var rchld *xnode
	if !node.isLeaf() {
		rchld = cow.children[len(cow.children)-1]
		cow.children = cow.children[:len(cow.children)-1]
	}
	return cow, item, rchld
}

// withCutLeft returns a clone of node, with the leftmost item cut off.
// If the node is a inner node, the leftmost child is cut off, too.
func (node xnode) withCutLeft() (xnode, xitem, *xnode) {
	assertThat(len(node.items) > 0, "attempt to cut left item from empty node")
	cow := node.clone()
	item := cow.items[0]
	cow.items = cow.items[1:len(cow.items)]
	var lchld *xnode
	if !node.isLeaf() {
		lchld = cow.children[0]
		cow.children = cow.children[1:len(cow.children)]
	}
	return cow, item, lchld
}

// --------------------

func (node xnode) clone() xnode {
	return node.cloneWithCapacity(0)
}

func (node xnode) cloneWithCapacity(cap int) xnode {
	itemcnt := len(node.items)
	n := xnode{}
	if itemcnt == 0 && cap <= 0 {
		return n
	}
	if cap < itemcnt {
		cap = itemcnt
	}
	if cap == 0 {
		return n
	}
	cap = ceiling(cap) // there must always be room for itemcnt + 2
	assertThat(cap > itemcnt, "cap has to be ceiling(itemcnt)[%d] > itemcnt[%d]", cap, itemcnt)
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
	size := to - from
	s := xnode{items: make([]xitem, size, ceiling(size))}
	copy(s.items, node.items[from:to])
	if len(node.children) > 0 {
		s.children = make([]*xnode, size, ceiling(size))
		copy(s.children, node.children[from:to])
	}
	return s
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

// findSlot searches a key within the items of node.
// Returns the correct index for key, and found=true, if found exactly.
func (node *xnode) findSlot(key K) (bool, int) {
	items, itemcnt := node.items, len(node.items)
	k := key
	slotinx := sort.Search(itemcnt, func(i int) bool {
		return items[i].key >= k // sort.Search will find the smallest i for which this is true
	})
	//tracer().Debugf("slot index ∈ %v = %d", items, slotinx)
	return slotinx < itemcnt && k == items[slotinx].key, slotinx
}

// --- Splitting and balancing -----------------------------------------------

/*
B-trees need to be re-balanced after a modification leaves the tree in a state, where
a tree-property is violated. For example, an insertion may procuduce an item-count for a node
which exceeds the high water mark, making a split of this node necessary.

We never do proactive re-balancing (which is described by some books about B-trees), but rather
re-balance after modification.

With immuatable persistent trees we need to make copies of all of the nodes affected by a
“modification”. This usually creates clones for a path of nodes, starting at the root and
ending at a leaf. Most of the children pointed to by each node will be unmodified, resulting
in incarnations of the tree sharing most of the nodes.
*/

// splitChild splits an overfull child node.
// It is not checked if the child is indeed overfull.
// Returns a modified copy of node with 2 new children, where the left one substitues a child of node.
//
// It's legal to pass in xnode{} as node (in order to create a new Tree.root).
//
func (node xnode) splitChild(ch slot) slot {
	child := ch.node
	half := len(child.items) / 2
	miditem := child.items[half] // find the median item to split at
	siblingL := child.slice(0, half)
	siblingR := child.slice(half+1, -1)
	tracer().Debugf("split: med = %v, len(L) = %d, len(R) = %d", miditem, len(siblingL.items), len(siblingR.items))
	found, index := node.findSlot(miditem.key)
	assertThat(!found, "internal inconsistency: child has same key as parent (during split)")
	cow := node.withInsertedItem(miditem, index).asNonLeaf()
	tracer().Debugf("split: parent is now %s", cow)
	cow.children[index] = &siblingL
	cow.children[index+1] = &siblingR
	return slot{node: &cow, index: index}
}

func cloneSeam(parent, child slot) slot {
	tracer().Debugf("seam: parent = %s, child = %s", parent, child)
	cowParent := parent.node.clone()
	cowParent.children[parent.index] = child.node
	return slot{node: &cowParent, index: parent.index}
}

func splitAndClone(highWaterMark uint) func(slot, slot) slot {
	return func(parent, child slot) slot {
		tracer().Debugf("split&propagate: parent = %s, child = %s", parent, child)
		if child.node.overfull(highWaterMark) {
			tracer().Debugf("child is overfull: %v", child)
			return parent.node.splitChild(child)
		}
		return cloneSeam(parent, child)
	}
}

func balance(lowWaterMark uint) func(slot, slot) slot {
	return func(parent, child slot) slot {
		tracer().Debugf("balance: parent = %s, child = %s", parent, child)
		if child.node.underfull(lowWaterMark) {
			tracer().Debugf("child is underfull: %v", child)
			return parent.balance(child, lowWaterMark)
		}
		return cloneSeam(parent, child)
	}
}

func (parent slot) balance(child slot, lowWaterMark uint) slot {
	assertThat(len(parent.node.children) > 0, "attempt to balance parent w/ zero children")
	if !parent.leftSibling(child).underfull(lowWaterMark + 1) {
		// steal item from left sibling ⇒ rotate right
		return parent.rotateRight(parent.leftSibling(child), child)
	} else if !parent.rightSibling(child).underfull(lowWaterMark + 1) {
		// steal item from right sibling ⇒ rotate left
		return parent.rotateLeft(child, parent.rightSibling(child))
	}
	// steal item from parent and merge with a sibling
	return parent.merge(parent.siblings2(child))
}

// merge steals an item from parent and merges child with a sibling.
// Returns a new parent which may be underfull or even empty (in case of parent being root).
//
// siblings is the pair of slots to merge. child is one of this pair, and we need it to
// know which item of the parent to extract.
func (parent slot) merge(mi mergeinfo) slot {
	assertThat(parent.len() > 0, "attempt to extract an item from an empty parent node")
	assertThat(parent.node == mi.parent.node, "internal inconsistency")
	tracer().Debugf("merge: parent = %s", mi.parent)
	tracer().Debugf("       sibling L = %s", mi.left)
	tracer().Debugf("       sibling R = %s", mi.right)
	cow := parent.node.withDeletedItem(mi.parent.index)
	newParent := slot{node: &cow, index: mi.parent.index}
	//lsbl, rsbl := siblings[0], siblings[1] // rsbl may be slot{}, i.e. empty
	lsbl, rsbl := mi.left, mi.right // mi.right may be slot{}, i.e. empty
	cap := lsbl.len() + rsbl.len() + 1
	cowch := lsbl.node.cloneWithCapacity(cap)
	assertThat(len(cowch.items) == len(lsbl.node.items), "internal inconsistency")
	cowch.items = append(cowch.items, mi.parent.item())
	cowch.items = append(cowch.items, rsbl.items()...)
	if !cowch.isLeaf() && rsbl.len() > 0 {
		cowch.children = append(cowch.children, rsbl.node.children...)
		assertThat(len(cowch.children) == lsbl.len()+1, "internal inconsistency")
	}
	cow.children[mi.parent.index] = &cowch // link new parent to new child
	return newParent
}

func (parent slot) rotateRight(lsbl, rsbl slot) slot {
	cow := parent.node.clone()
	newParent := slot{node: &cow, index: parent.index}
	// cut rightmost item from left sibling
	cowlsbl, lsblxitem, grandChild := lsbl.node.withCutRight()
	// replace parent item with item from left sibling
	parentxitem := newParent.replaceItem(lsblxitem)
	// insert parent item as leftmost item in child
	cowrsbl := rsbl.node.withInsertedItem(parentxitem, 0)
	if !cowrsbl.isLeaf() {
		assertThat(len(cowlsbl.children) == len(cowlsbl.items)+1, "insertion logic failed")
		cowrsbl.children[0] = grandChild
	}
	// link new children of parent/cow
	cow.children[parent.index] = &cowlsbl
	cow.children[parent.index+1] = &cowrsbl
	return newParent
}

func (parent slot) rotateLeft(lsbl, rsbl slot) slot {
	cow := parent.node.clone()
	newParent := slot{node: &cow, index: parent.index}
	// cut leftmost item from right sibling
	cowrsbl, rsblxitem, grandChild := rsbl.node.withCutLeft()
	// replace parent item with item from right sibling
	parentxitem := newParent.replaceItem(rsblxitem)
	// insert parent item as rightmost item in child
	cowlsbl := lsbl.node.withInsertedItem(parentxitem, len(lsbl.node.items))
	if !cowlsbl.isLeaf() {
		assertThat(len(cowlsbl.children) == len(cowlsbl.items)+1, "insertion logic failed")
		cowlsbl.children[len(cowlsbl.items)] = grandChild
	}
	// link new children of parent/cow
	cow.children[parent.index] = &cowlsbl
	cow.children[parent.index+1] = &cowrsbl
	return newParent
}

// stealPredOrSucc searches the left and/or right sub-tree of s, returning a path to a
// leaf node, which is either the predecessor or the successor of s.
//
// Not pure: modifies pathBuf. pathBuf may not be invalid, but rather must be a buffer
// from an earlier call to `findKeyAndPath(…)`.
func (s slot) stealPredOrSucc(pathBuf slotPath, lowWaterMark uint) (item xitem, path slotPath) {
	assertThat(pathBuf != nil && len(pathBuf) > 0 && cap(pathBuf) > len(pathBuf), "invalid path buffer")
	tracer().Debugf("parent = %s, path.last = %s", s, pathBuf.last())
	//assertThat(pathBuf.last().node == s.node, "need path with parent as last node")
	path = pathBuf
	pinx := len(path) - 1
	var found bool
	found, path = s.findSucc(path)
	if found && len(path.last().node.items) > int(lowWaterMark) {
		path[pinx].index++
	} else {
		path = pathBuf
		path = s.findPred(path)
	}
	tracer().Debugf("slot path to steal -> %s", path)
	return path.last().item(), path
}

// Not pure: modifies pathBuf.
func (s slot) findPred(pathBuf slotPath) slotPath {
	path := pathBuf
	node := s.node.children[s.index]
	for !node.isLeaf() {
		tracer().Debugf("find pred: visiting inner node = %v", node)
		path = append(path, slot{node: node, index: len(node.items)})
		node = node.children[len(node.children)-1]
		assertThat(node != nil, "right-most child of inner node is missing")
	}
	tracer().Debugf("find pred: visiting leaf node = %v", node)
	path = append(path, slot{node: node, index: len(node.items) - 1})
	tracer().Debugf("slot path for pred -> %s", path)
	return path
}

// Not pure: modifies pathBuf.
func (s slot) findSucc(pathBuf slotPath) (bool, slotPath) {
	assertThat(s.index < len(s.node.items), "inner node has no right child")
	path := pathBuf
	assertThat(len(s.node.children) >= s.index+1, "right-most child of inner node is missing")
	node := s.node.children[s.index+1]
	for !node.isLeaf() {
		tracer().Debugf("find succ: visiting inner node = %v", node)
		path = append(path, slot{node: node, index: 0})
		node = node.children[0]
	}
	tracer().Debugf("find succ: visiting leaf node = %v", node)
	path = append(path, slot{node: node, index: 0})
	tracer().Debugf("slot path for succ -> %s", path)
	return true, path
}

// --- Helpers ---------------------------------------------------------------

func assertThat(that bool, msg string, msgargs ...interface{}) {
	if !that {
		msg = fmt.Sprintf("btree: "+msg, msgargs...)
		panic(msg)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ceiling(n int) int {
	if n <= 0 {
		return 0
	}
	// we need N=n+2 entries, but start the algorithm with N=n-1 => N=n+1
	n = n + 1
	for n&(n-1) > 0 { // do till only one bit is left
		n = n & (n - 1) // unset rightmost bit
	} // `n` is now a power of two (less than `n`)
	return n << 1 // return next power of 2
}
