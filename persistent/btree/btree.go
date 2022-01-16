package btree

/*
TODOs:

- manage sizes of node.items and node.children correctly
- check for safe handling of empty nodes, especially root
- manage XTree.depth correctly after split or merge
- replace K and T with generic types for Go 1.18
*/

const defaultLowWaterMark uint = 3 // 2^n - 1
// high water mark includes space for +1 child link and for a stopper
var defaultHighWaterMark uint = uint(ceiling(int(defaultLowWaterMark)*2)) - 2

type Tree struct {
	root          *xnode
	depth         uint
	lowWaterMark  uint
	highWaterMark uint
}

func Immutable(opts ...Option) Tree {
	tree := Tree{
		lowWaterMark:  defaultLowWaterMark,
		highWaterMark: defaultHighWaterMark,
	}
	for _, option := range opts {
		tree = option(tree)
	}
	return tree
}

type Option func(Tree) Tree

func LowWaterMark(n int) Option {
	return func(tree Tree) Tree {
		nodesize := ceiling(max(3, n))
		tree.lowWaterMark = uint(nodesize<<2) - 1
		tree.highWaterMark = uint(nodesize) - 2
		return tree
	}
}

// --- API -------------------------------------------------------------------

func (tree Tree) Find(key K) (T, bool) {
	var found bool
	var path slotPath = make([]slot, tree.depth)
	if found, path = tree.findKeyAndPath(key, path); found {
		return path.last().item().value, true
	}
	var none T
	return none, false
}

func (tree Tree) With(key K, value T) Tree {
	var path slotPath = make([]slot, tree.depth)
	var found bool
	if found, path = tree.findKeyAndPath(key, path); found {
		if path.last().item().value == value {
			return tree // no need for modification
		}
		return tree.replacing(key, value, path) // otherwise copy with replaced value
	}
	tracer().Debugf("insert: slot path = %s", path)
	item := xitem{key, value}
	if tree.root == nil { // virgin tree => insert first node and return
		return tree.shallowCloneWithRoot(xnode{}.withInsertedItem(item, 0)).withDepth(1)
	}
	leafSlot := path.last()
	assertThat(leafSlot.node.isLeaf(), "attempt to insert item at non-leaf")
	cow := leafSlot.node.withInsertedItem(item, leafSlot.index) // copy-on-write
	tracer().Debugf("insert: created copy of (leaf + key@%d) = %s", leafSlot.index, cow)
	newRoot := path.dropLast().foldR(splitAndClone(tree.highWaterMark),
		slot{node: &cow, index: leafSlot.index},
	)
	tracer().Debugf("insert: new root = %s", newRoot)
	if newRoot.node.overfull(tree.highWaterMark) {
		newRoot = xnode{}.splitChild(newRoot)
		tree.depth++ // miss-use of tree for intermediate storage of new depth
	}
	return tree.shallowCloneWithRoot(*newRoot.node)
}

func (tree Tree) WithDeleted(key K) Tree {
	var path slotPath = make([]slot, tree.depth)
	var found bool
	if found, path = tree.findKeyAndPath(key, path); !found {
		return tree // no need for modification
	}
	tracer().Debugf("deletion: slot path = %s", path)
	del := path.last()
	var cowLeaf xnode
	var leafSlot slot
	if del.node.isLeaf() {
		cow := del.node.withDeletedItem(del.index) // copy-on-write
		tracer().Debugf("created copy of leaf w/out deleted item: %v", cow.items)
		leafSlot = slot{node: &cow, index: del.index}
	} else { // for inner node:
		// swap item with rightmost item of left subtree or leftmost item of right subtree
		cow := del.node.clone()                                            // cow is clone of inner node
		path[len(path)-1].node = &cow                                      // remember clone in path
		leafItem, leafPath := del.stealPredOrSucc(path, tree.lowWaterMark) // from left or right subtree
		cow.items[del.index] = leafItem                                    // insert stolen item
		l := leafPath.last()                                               //
		cowLeaf = l.node.withDeletedItem(l.index)                          // remove stolen item from leaf
		path = leafPath                                                    // continue with path from root to leaf
		leafSlot = slot{node: &cowLeaf, index: l.index}
	}
	// balance from leaf-node upwards, starting at the leaf where we deleted an item
	tracer().Debugf("after delete: path = %v", path)
	newRoot := path.dropLast().foldR(balance(tree.lowWaterMark),
		leafSlot,
		//slot{node: &cowLeaf, index: del.index},
	)
	tracer().Debugf("deletion: new root = %s", newRoot)
	newTree := tree.shallowCloneWithRoot(*newRoot.node)
	switch { // catch border cases where root is empty after deletion
	case newRoot.len() == 0 && newRoot.node.children[0] != nil:
		newTree.root = newRoot.node.children[0]
		newTree.depth--
	case newRoot.len() == 0 && newRoot.node.isLeaf():
		newTree.root = nil
		newTree.depth = 0
	}
	return newTree
}
