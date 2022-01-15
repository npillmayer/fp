package btree

import (
	"testing"

	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestTreeCreateMock(t *testing.T) {
	tree := createMockTree()
	if tree.root == nil {
		t.Error("cannot create mock tree")
	}
}

func TestTreeFindInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := Tree{}
	_, path := tree.findKeyAndPath("7", nil)
	if len(path) > 0 {
		t.Errorf("expected path for '7' to be nil, is %v", path)
	}
}

func TestTreeFindKeyAndPath(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
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

func TestTreeInsertInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
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

func TestTreeInsertWith(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
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

func TestTreeInsertWithInLeaf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
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
