package btree

/*
Remarks:
--------

- 'cow' stands for copy-on-write and is used throughout the code for variables holding clones of nodes.

- We use a programming-style reminiscent of functional programming (see remarks on
  re-balancing) where it makes things easier to understand.

- A new modified incarnation of a tree always is reflected by a new tree.root.

*/

const defaultLowWaterMark uint = 3 // 2^n - 1
// high water mark includes space for +1 child link and for a stopper
var defaultHighWaterMark uint = uint(ceiling(int(defaultLowWaterMark)*2)) - 2

// Tree is an in-memory B-tree. An empty instance is usable as an empty tree, i.e.
// this is legal:
//
//     tree := btree.Tree[int,int]{}.With(1, 42)
//
// returning a tree containing a single node ⟨1⟩ associated with value 42.
//
type Tree struct {
	root          *xnode
	depth         uint
	lowWaterMark  uint
	highWaterMark uint
}

// Immutable constructs a B-tree with options, if you need any.
// Use it like this:
//
//     tree := btree.Immutable[int, string](Degree(16))
//     tree = tree.With(42, "Galaxy")
//     value, found := tree.Find(42)   // returns "Galaxy"
//
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

// Option is a type to help initializing B-trees at creation time.
type Option func(Tree) Tree

// Degree is an option to set the minimum number of children a node in the tree owns.
// The lower bound for the degree is 3.
//
// Use it like this:
//
//     tree := btree.Immutable[int, string](Degree(16))
//
func Degree(n int) Option {
	return func(tree Tree) Tree {
		low := max(2, n-1)
		tree.lowWaterMark = uint(low)
		tree.highWaterMark = uint(ceiling(int(tree.lowWaterMark)*2)) - 2
		return tree
	}
}

// --- API -------------------------------------------------------------------

// Find locates a key in a tree, if present, and returns the value associated with the key.
// If `key` is not found, the zero value for type T will be returned, together with found=false.
func (tree Tree) Find(key K) (T, bool) {
	var found bool
	var path slotPath = make([]slot, tree.depth)
	if found, path = tree.findKeyAndPath(key, path); found {
		return path.last().item().value, true
	}
	var none T
	return none, false
}

// With returns a copy of a tree with a new key inserted, which is associated with `value`.
// If an entry for key is already present in tree, the associated value will be replaced
// (in a new incarnation of the tree, nevertheless).
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

// With returns a copy of a tree with key deleted, if present, together with its associated value.
// If key is not found, tree is returned unchanged.
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
		leafSlot = slot{node: &cowLeaf, index: l.index}                    // leaf to start balancing
	}
	// balance from leaf-node upwards, starting at the leaf where we deleted an item
	tracer().Debugf("after delete: path = %v", path)
	newRoot := path.dropLast().foldR(balance(tree.lowWaterMark),
		leafSlot,
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

// --- Ext -------------------------------------------------------------------

// TreeExtension represents a B-tree as a tree and exposes some of its tree properties.
//
// This is something I'll need for using B-trees as ropes/cords in the
// future. I'm not yet sure of how to go about it in a general way, but my current
// thinking is that Extension will let clients treat a tree like a tree (in a
// controlled fashion), while the primary API of B-tree is more like a map.
//
type TreeExtension struct {
	tree Tree
	ext  Ext
}

func (tex TreeExtension) Locate(key K) Location {
	return tex.tree.locate(6, tex.ext, nil)
}

// Ext returns a tree extension for a given incarnation of a tree.
// This will wrap a client-provided Ext into an opaque TreeExtension, which then will
// manage accessing tree-properties of B-trees.
//
// Supplying nil as an ext result in returning a default type TreeExtension.
//
func (tree Tree) Ext(ext Ext) TreeExtension {
	return TreeExtension{} // TODO
}

// Location reflects a key/value pair in the B-tree, together with the node-path to it.
// A location is valid for a specific incarnation of a tree only; applying any of its methods
// on a different incarnation will result in a panic.
type Location struct {
	rootNode *xnode
	path     slotPath
	present  bool
}

// Ext (extensions) is something I'll need for using B-trees as ropes/cords in the
// future. I'm not yet sure of how to go about it in a general way, but my current
// thinking is that Extension will let clients treat a tree like a tree (in a
// controlled fashion), while the primary API of B-tree is more like a map.
type Ext interface {
	Agg(key, agg K) K
	Cmp(key, itemKey, agg K) int
}

type aggregator func(key, agg K) K
type comparator func(key, itemKey, agg K) int

type _DefaultExtension struct{}

func (dext _DefaultExtension) Agg(key, agg K) K {
	return agg + key
}

func (dext _DefaultExtension) Cmp(key, itemKey, agg K) int {
	switch {
	case key == itemKey:
		return 0
	case key < itemKey:
		return -1
	default:
		return +1
	}
}

func lagg(key, agg K) K {
	return agg + key
}

func find(key, itemKey, agg K) int {
	tracer().Debugf("find: f(key=%v, item.key=%v, agg=%v)", key, itemKey, agg)
	if key == itemKey {
		return 0
	}
	if key < itemKey {
		return -1
	}
	return +1
}

func (tree Tree) locate(key K, ext Ext, pathBuf slotPath) Location {
	path := pathBuf[:0] // we track the path to the key's slot
	if tree.root == nil {
		return Location{rootNode: tree.root, path: path, present: false}
	}
	var agg K
	var cmp, index int
	var node *xnode = tree.root // walking nodes, start search at the top
	for !node.isLeaf() {
		tracer().Debugf("finding inner node = %v", node)
		cmp, index, agg = cmpNode(node, key, agg, ext.Cmp, ext.Agg)
		path = append(path, slot{node: node, index: index})
		if cmp == 0 {
			return Location{rootNode: tree.root, path: path, present: true}
		}
		node = node.children[index]
	}
	tracer().Debugf("finding leaf node %v", node)
	cmp, index, _ = cmpNode(node, key, agg, ext.Cmp, ext.Agg)
	path = append(path, slot{node: node, index: index})
	tracer().Debugf("locate: slot path = %s", path)
	return Location{rootNode: tree.root, path: path, present: cmp == 0}
}

func cmpNode(node *xnode, key, agg K, f comparator, a aggregator) (cmp, index int, aggout K) {
	aggout = agg
	for index = 0; index < len(node.items); index++ {
		item := node.items[index]
		cmp = f(key, item.key, agg)
		tracer().Debugf("f(%v,%v,%v) -> %v | %v", key, item.key, agg, agg, cmp)
		switch {
		case cmp < 0:
			return
		case cmp == 0:
			return
		case cmp > 0:
			aggout = a(item.key, aggout)
		}
	}
	return
}
