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

func TestInternalNodeReplaceValue(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: "1", value: 1}, 0)
	node = node.withReplacedValue(xitem{key: "1", value: 7}, 0)
	if node.items[0].value != 7 {
		t.Errorf("expected item.0.value to be 7, is %v", node.items[0])
	}
}

func TestInternalNodeDeleteItem(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: "1", value: 1}, 0)
	node = node.withInsertedItem(xitem{key: "3", value: 3}, 1)
	node = node.withInsertedItem(xitem{key: "5", value: 5}, 2)
	node = node.withDeletedItem(1)
	if node.items[1].value != 5 {
		t.Errorf("expected item.1.value to be 5, is %v", node.items[0])
	}
	if len(node.items) != 2 {
		t.Errorf("expected node to have 2 items, has %v", len(node.items))
	}
	if cap(node.items) != ceiling(3) { // cap shrinking has delay
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(3), cap(node.items))
	}
	node = node.withDeletedItem(1)
	if len(node.items) != 1 {
		t.Errorf("expected node to have 1 item, has %v", len(node.items))
	}
	if cap(node.items) != ceiling(2) { // cap shrinking has delay
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(2), cap(node.items))
	}
}

func TestInternalCut(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: "1", value: 1}, 0)
	node = node.withInsertedItem(xitem{key: "3", value: 3}, 1)
	node = node.withInsertedItem(xitem{key: "5", value: 5}, 2)
	rest, item, _ := node.withCutLeft()
	if len(rest.items) != 2 {
		t.Errorf("expected len(rest) of cut-off to be 2, is %d", len(rest.items))
	}
	if item.key != "1" {
		t.Errorf("expected cut-off item to be '1', is %q", item.key)
	}
	node = rest
	rest, item, _ = node.withCutRight()
	if len(rest.items) != 1 {
		t.Logf("rest = %v", rest)
		t.Errorf("expected len(rest) of cut-off to be 1, is %d", len(rest.items))
	}
	if item.key != "5" {
		t.Errorf("expected cut-off item to be '5', is %q", item.key)
	}
	node = rest
	rest, item, _ = node.withCutRight()
	if len(rest.items) != 0 {
		t.Errorf("expected len(rest) of cut-off to be 1, is %d", len(rest.items))
	}
}
