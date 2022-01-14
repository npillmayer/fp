package btree

import (
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestCreateMock(t *testing.T) {
	tree := createMockTree()
	if tree.root == nil {
		t.Error("cannot create mock tree")
	}
}

func TestFindInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := Tree{}
	_, path := tree.findKeyAndPath("7", nil)
	if len(path) > 0 {
		t.Errorf("expected path for '7' to be nil, is %v", path)
	}
}

func TestFindKeyAndPath(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := createMockTree()
	found, path := tree.findKeyAndPath("9", nil)
	if !found {
		t.Logf("path = %v", path)
		t.Error("expected to have found item with key=9, didn't")
	}
	if len(path) != 2 {
		t.Logf("path = %v", path)
		t.Fatalf("expected length of path to be 2, is %d", len(path))
	}
	if path[1].index != 2 {
		t.Logf("path = %v", path)
		t.Errorf("expected slot to be at pos=2 of leaf, is %d", path[1].index)
	}
}

func TestFindInNode(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := (&xnode{}).add("1", "2", "3", "4", "5", "6", "7", "8", "9")
	found, at := node.findSlot("7")
	if !found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("1: expected findSlot to find 7 at position 6, didn't")
	}
	node = (&xnode{}).add("1", "2", "3", "4", "5", "6", "8", "9")
	found, at = node.findSlot("7")
	if found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("2: expected findSlot to find empty slot for 7 at position 6, didn't")
	}
	node = &xnode{}
	found, at = node.findSlot("7")
	if found || at != 0 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("3: expected empty.findSlot to find empty slot for 7 at position 0, didn't")
	}
	node = (&xnode{}).add("1", "2", "3", "4", "5", "6")
	found, at = node.findSlot("7")
	if found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("4: expected findSlot to find empty slot for 7 at final position 6, didn't")
	}
}

func TestInsertInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := Tree{}.With("7", 7)
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	if tree.depth != 1 {
		t.Logf("tree.root = %s", tree.root)
		t.Errorf("expected tree.With(…) to produce tree.depth=1, has %d", tree.depth)
	}
	if !tree.root.isLeaf() {
		t.Logf("tree.root = %s", tree.root)
		t.Error("expected tree.root to be a leaf, isn't")
	}
}

func TestInsertWith(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := Tree{}.With("7", 7)
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	tree = tree.With("3", 3)
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	t.Logf("tree = %#v", tree)
}

func TestInsertWithInLeaf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := createMockTree()
	tree = tree.With("7", "7")
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	t.Logf("tree = %#v", tree)
}

// ---------------------------------------------------------------------------

func createMockTree() Tree { // tree with values 0…9, without 7
	root := &xnode{}
	root.add("2", "5")

	child0 := &xnode{}
	child0.add("0", "1")
	root.children = append(root.children, child0)

	child1 := &xnode{}
	child1.add("3", "4")
	root.children = append(root.children, child1)

	child2 := &xnode{}
	child2.add("6", "8", "9") // 7 is missing
	root.children = append(root.children, child2)

	//return newTreeWithRoot(root, minItems)
	return Tree{
		root:          root,
		depth:         2,
		lowWaterMark:  defaultLowWaterMark,
		highWaterMark: defaultHighWaterMark * 2,
	}
}

func (node *xnode) add(keys ...K) *xnode {
	for _, key := range keys {
		node.items = append(node.items, xitem{key, key})
	}
	return node
}
