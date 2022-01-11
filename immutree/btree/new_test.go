package btree

import (
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestCreateMock(t *testing.T) {
	b := createMockTree()
	if b == nil {
		t.Error("cannot create mock tree")
	}
}

func TestFindKeyAndPath(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	b := createMockTree()
	found, path := b.findKeyAndPath("7", nil)
	if !found {
		t.Logf("path = %v", path)
		t.Error("expected to have found item with key=7, didn't")
	}
	if len(path) != 2 {
		t.Logf("path = %v", path)
		t.Fatalf("expected length of path to be 2, is %d", len(path))
	}
	if path[1].index != 1 {
		t.Logf("path = %v", path)
		t.Errorf("expected slot to be at pos=1 of leaf, is %d", path[1].index)
	}
}

func TestFindInNode(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := NewEmptyNode()
	node.add("1", "2", "3", "4", "5", "6", "7", "8", "9")
	found1, at1 := node.findKey("7")
	found2, at2 := node.findSlot("7")
	if found1 != found2 || at1 != at2 {
		t.Logf("found1 = %v, found2 = %v", found1, found2)
		t.Logf("at1    = %v, at2    = %v", at1, at2)
		t.Error("expected findKey and findSlot to return the same, don't")
	}
}

// ---------------------------------------------------------------------------

func createMockTree() *Tree {
	root := NewEmptyNode()
	root.add("2", "5")

	child0 := NewEmptyNode()
	child0.add("0", "1")
	root.addChildNode(child0)

	child1 := NewEmptyNode()
	child1.add("3", "4")
	root.addChildNode(child1)

	child2 := NewEmptyNode()
	child2.add("6", "7", "8", "9")
	root.addChildNode(child2)

	return newTreeWithRoot(root, minItems)
}

func (n *Node) add(keys ...string) *Node {
	for _, key := range keys {
		n.items = append(n.items, newItem(key, key))
	}
	return n
}
