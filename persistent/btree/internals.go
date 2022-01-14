package btree

import (
	"fmt"
	"sort"
)

func (tree XTree) replacing(key K, value T, path slotPath) (newTree XTree) {
	assertThat(len(path) > 0, "cannot replace item without path")
	tracer().Debugf("btree.With: slot path = %s", path)
	hit := path[len(path)-1] // slot where `key` lives
	item := xitem{key: key, value: value}
	cow := hit.node.withReplacedValue(item, hit.index)
	tracer().Debugf("created copy of node for replacement: %#v", cow)
	newRoot := path.dropLast().foldR(cloneSeam, slot{node: &cow, index: hit.index})
	tracer().Debugf("replace: top = %s", newRoot)
	newTree.root = newRoot.node
	return
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

func cloneSeam(parent, child slot) slot {
	tracer().Debugf("seam: parent = %s, child = %s", parent, child)
	cowParent := parent.node.clone()
	cowParent.children[parent.index] = child.node
	return slot{node: &cowParent, index: parent.index}
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

func (tree XTree) findKeyAndPath(key K, pathBuf slotPath) (found bool, path slotPath) {
	path = pathBuf[:0] // we track the path to the key's slot
	if tree.root == nil {
		return
	}
	var index int
	var node *xnode = tree.root // walking nodes, start search at the top
	for !node.isLeaf() {
		tracer().Debugf("node = %v", node)
		found, index = node.findSlot(key)
		path = append(path, slot{node: node, index: index})
		if found {
			return // we have an exact match
		}
		node = node.children[index]
	}
	tracer().Debugf("node = %v", node)
	found, index = node.findSlot(key)
	path = append(path, slot{node: node, index: index})
	tracer().Debugf("slot path for key=%v -> %s", key, path)
	return
}

func (node *xnode) findSlot(key K) (bool, int) {
	items, itemcnt := node.items, len(node.items)
	k := key
	slotinx := sort.Search(itemcnt, func(i int) bool {
		return items[i].key >= k // sort.Search will find the smallest i for which this is true
	})
	tracer().Debugf("slot index ∈ %v = %d", items, slotinx)
	return slotinx < itemcnt && k == items[slotinx].key, slotinx
}

func (node xnode) withReplacedValue(item xitem, at int) xnode {
	assertThat(at <= len(node.items), "given item index out of range: %d < %d", len(node.items), at)
	cow := node.clone()
	cow.items[at].value = item.value
	return cow
}

func (node xnode) withDeletedItem(at int) xnode {
	assertThat(at+1 <= len(node.items), "no space for stopper-item: %d ≤ %d", len(node.items), at)
	cow := node.clone()
	cow.items = append(cow.items[:at], cow.items[at+1:]...) // stopper slot required!
	if !cow.isLeaf() {
		cow.children = append(cow.children[:at], cow.children[at+1:]...) // stopper slot required!
	}
	return cow
}

func (node xnode) withInsertedItem(item xitem, at int) xnode {
	assertThat(at <= len(node.items), "given item index out of range: %d < %d", len(node.items), at)
	cap := max(ceiling(at), len(node.items))
	cow := node.cloneWithCapacity(cap) // change-on-write behaviour requires copying
	if at == len(node.items) {         // append at the end
		cow.items = append(cow.items, item)
		return cow
	}
	cow.items = append(cow.items[:at], item)
	cow.items = append(cow.items, cow.items[at:]...)
	if !cow.isLeaf() {
		cow.children = append(cow.children[:at+1], nil)
		cow.children = append(cow.children, cow.children[at:]...)
	}
	return cow
}

func (node xnode) withCutRight() (xnode, xitem, *xnode) {
	assertThat(len(node.items) > 0, "attempt to cut right item from empty node")
	cow := node.clone()
	item := cow.items[len(cow.items)-1]
	rnode := cow.children[len(cow.children)-1]
	cow.items = cow.items[:len(cow.items)-1]
	cow.children = cow.children[:len(cow.children)-1]
	return cow, item, rnode
}

func (node xnode) withCutLeft() (xnode, xitem, *xnode) {
	assertThat(len(node.items) > 0, "attempt to cut left item from empty node")
	cow := node.clone()
	item := cow.items[0]
	rnode := cow.children[0]
	cow.items = cow.items[1:len(cow.items)]
	cow.children = cow.children[1:len(cow.children)]
	return cow, item, rnode
}

// splitChild splits an overfull child node.
// It is not checked if the child is indeed overfull.
// Returns a modified copy of node with 2 new children, where the left one substitues a child of node.
//
// It's legal to pass in xnode{} as node (in order to create a new Tree.root).
//
func (node xnode) splitChild(s slot) slot {
	child := s.node
	half := len(node.items) / 2
	medianxitem := child.items[half]
	siblingL := node.slice(0, half)
	siblingR := node.slice(half+1, -1)
	found, index := node.findSlot(K(medianxitem.key))
	assertThat(!found, "internal inconsistency: child has same key as parent (during split)")
	cow := node.withInsertedItem(medianxitem, index).asNonLeaf()
	cow.children[index] = &siblingL
	cow.children[index+1] = &siblingR
	return slot{node: &cow, index: index}
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
func (parent slot) merge(siblings [2]slot) slot {
	assertThat(parent.len() > 0, "attempt to extract an item from an empty parent node")
	cow := parent.node.withDeletedItem(parent.index)
	newParent := slot{node: &cow, index: parent.index}
	lsbl, rsbl := siblings[0], siblings[1] // rsbl may be slot{}, i.e. empty
	cap := ceiling(lsbl.len() + rsbl.len() + 1)
	cowch := lsbl.node.cloneWithCapacity(cap)
	assertThat(len(cowch.items) == len(lsbl.node.items), "internal inconsistency")
	cowch.items = append(cowch.items, parent.item())
	cowch.items = append(cowch.items, rsbl.items()...)
	if !cowch.isLeaf() && rsbl.len() > 0 {
		cowch.children = append(cowch.children, rsbl.node.children...)
		assertThat(len(cowch.children) == lsbl.len()+1, "internal inconsistency")
	}
	cow.children[parent.index] = &cowch // link new parent to new child
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
	return ((n + 1) >> 1) << 1
}
