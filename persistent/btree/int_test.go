package btree

import (
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

// test internals

func TestInternalCeiling(t *testing.T) {
	c := []struct {
		n    int
		ceil int
	}{
		{0, 0},
		{2, 4},
		{3, 8},
		{4, 8},
		{6, 8},
		{7, 16},
	}
	for i, x := range c {
		xx := ceiling(x.n)
		if xx != x.ceil {
			t.Errorf("%d: expected ceiling(%d) to be %d, is %d", i, x.n, x.ceil, xx)
		}
	}
}

func TestInternalNodeInsert(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: "1", value: 1}, 0)
	if len(node.items) != 1 {
		t.Errorf("expected item count of node to be 1, is %d", len(node.items))
	}
	if cap(node.items) != ceiling(1) {
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(1), cap(node.items))
	}
	node = node.withInsertedItem(xitem{key: "3", value: 3}, 1)
	if len(node.items) != 2 {
		t.Errorf("expected item count of node to be 2, is %d", len(node.items))
	}
	if cap(node.items) != ceiling(2) {
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(2), cap(node.items))
	}
	if node.items[0].key != "1" {
		t.Logf("node = %s", node)
		t.Errorf("expected item 0 to be 1, is %v", node.items[0])
	}
}

func TestInternalSlotReplaceItem(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: "1", value: 1}, 0)
	node = node.withReplacedValue(xitem{key: "1", value: 7}, 0)
	if node.items[0].value != 7 {
		t.Errorf("expected item.0.value to be 7, is %v", node.items[0])
	}
}
