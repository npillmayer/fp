package btree

import (
	"strconv"
	"testing"

	"github.com/npillmayer/schuko/tracing"
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

// --- Nodes -----------------------------------------------------------------

func TestInternalNodeSlice(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	defer teardown()
	//
	node := xnode{}
	var keys = []K{1, 2, 3, 4, 5}
	cap := ceiling(len(keys))
	node.items = make([]xitem, len(keys), cap)
	node.children = make([]*xnode, len(keys), cap)
	grandson := &xnode{}
	for i := 0; i < len(keys); i++ {
		node.items[i] = xitem{key: keys[i], value: strconv.Itoa(int(keys[i]))}
		node.children[i] = grandson
	}
	var slices = []struct{ f, t, l int }{ // from, to, length
		{f: 0, t: 0, l: 0},
		{f: 0, t: 2, l: 2},
		{f: 0, t: 5, l: 5},
		{f: 2, t: 4, l: 2},
		{f: 2, t: 2, l: 0},
		{f: 5, t: 5, l: 0},
		{f: 2, t: -1, l: 3},
	}
	for i, x := range slices {
		s := node.slice(x.f, x.t)
		if len(s.items) != x.l {
			t.Logf("node = %s, slice(%d,%d) = %s", node, x.f, x.t, s)
			t.Errorf("%d: expected items slice of length = %d, have %d", i, x.l, len(s.items))
		}
		if len(s.children) != x.l {
			t.Errorf("%d: expected children slice of length = %d, have %d", i, x.l, len(s.items))
		}
	}
}

func TestInternalNodeInsert(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: 1, value: "1"}, 0)
	if len(node.items) != 1 {
		t.Errorf("expected item count of node to be 1, is %d", len(node.items))
	}
	if cap(node.items) != ceiling(1) {
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(1), cap(node.items))
	}
	node = node.withInsertedItem(xitem{key: 3, value: "3"}, 1)
	if len(node.items) != 2 {
		t.Errorf("expected item count of node to be 2, is %d", len(node.items))
	}
	if cap(node.items) != ceiling(2) {
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(2), cap(node.items))
	}
	if node.items[0].key != 1 {
		t.Logf("node = %s", node)
		t.Errorf("expected item 0 to be 1, is %v", node.items[0])
	}
	node = node.withInsertedItem(xitem{key: 7, value: "7"}, 1)
	if len(node.items) != 3 {
		t.Fatalf("expected item count of node to be 3, is %d", len(node.items))
	}
	if cap(node.items) != ceiling(3) {
		t.Logf("node = %s", node)
		t.Errorf("expected node-capacity to be %d, is %d", ceiling(3), cap(node.items))
	}
	if node.items[1].key != 7 {
		t.Logf("node = %s", node)
		t.Errorf("expected item 1 to be 7, is %v", node.items[1])
	}
	if node.items[2].key != 3 {
		t.Logf("node = %s", node)
		t.Errorf("expected item 2 to be 3, is %v", node.items[2])
	}
}

func TestInternalNodeReplaceValue(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: 1, value: 1}, 0)
	node = node.withReplacedValue(xitem{key: 1, value: 7}, 0)
	if node.items[0].value != 7 {
		t.Errorf("expected item.0.value to be 7, is %v", node.items[0])
	}
}

func TestInternalNodeDeleteItem(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: 1, value: 1}, 0)
	node = node.withInsertedItem(xitem{key: 3, value: 3}, 1)
	node = node.withInsertedItem(xitem{key: 5, value: 5}, 2)
	node = node.withDeletedItem(1)
	if node.items[1].value != 5 {
		t.Errorf("expected item[1].value to be 5, is %v", node.items[0])
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
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	defer teardown()
	//
	node := xnode{}.withInsertedItem(xitem{key: 1, value: 1}, 0)
	node = node.withInsertedItem(xitem{key: 3, value: 3}, 1)
	node = node.withInsertedItem(xitem{key: 5, value: 5}, 2)
	rest, item, _ := node.withCutLeft()
	if len(rest.items) != 2 {
		t.Errorf("expected len(rest) of cut-off to be 2, is %d", len(rest.items))
	}
	if item.key != 1 {
		t.Errorf("expected cut-off item to be '1', is %q", item.key)
	}
	node = rest
	rest, item, _ = node.withCutRight()
	if len(rest.items) != 1 {
		t.Logf("rest = %v", rest)
		t.Errorf("expected len(rest) of cut-off to be 1, is %d", len(rest.items))
	}
	if item.key != 5 {
		t.Errorf("expected cut-off item to be '5', is %q", item.key)
	}
	node = rest
	rest, item, _ = node.withCutRight()
	if len(rest.items) != 0 {
		t.Errorf("expected len(rest) of cut-off to be 1, is %d", len(rest.items))
	}
}

func TestInternalFindSlot(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	node := (&xnode{}).add(1, 2, 3, 4, 5, 6, 7, 8, 9)
	found, at := node.findSlot(7)
	if !found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("1: expected findSlot to find 7 at position 6, didn't")
	}
	node = (&xnode{}).add(1, 2, 3, 4, 5, 6, 8, 9)
	found, at = node.findSlot(7)
	if found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("2: expected findSlot to find empty slot for 7 at position 6, didn't")
	}
	node = &xnode{}
	found, at = node.findSlot(7)
	if found || at != 0 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("3: expected empty.findSlot to find empty slot for 7 at position 0, didn't")
	}
	node = (&xnode{}).add(1, 2, 3, 4, 5, 6)
	found, at = node.findSlot(7)
	if found || at != 6 {
		t.Logf("found = %v, at = %d", found, at)
		t.Error("4: expected findSlot to find empty slot for 7 at final position 6, didn't")
	}
}

func TestInternalLocate(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	var ext ext
	loc := tree.locate(6, ext, nil)
	if !loc.present {
		t.Logf("found = %v, path = %s", loc.present, loc.path)
		t.Error("expected to find key 6 in tree, didn't")
	}
}

type ext struct{}

func (e ext) Agg(key, agg K) K {
	return lagg(key, agg)
}

func (e ext) Cmp(key, itemKey, agg K) int {
	return find(key, itemKey, agg)
}

// --- Paths -----------------------------------------------------------------

func TestInternalPathFold(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.btree")
	defer teardown()
	//
	var path slotPath = make([]slot, 3)
	node1 := xnode{}.withInsertedItem(xitem{1, 1}, 0)
	path[0] = slot{node: &node1, index: 0}
	node2 := xnode{}.withInsertedItem(xitem{2, 2}, 0)
	path[1] = slot{node: &node2, index: 0}
	node3 := xnode{}.withInsertedItem(xitem{3, 3}, 0)
	path[2] = slot{node: &node3, index: 0}
	//t.Logf("path = %v", path)
	node4 := xnode{}.withInsertedItem(xitem{4, 4}, 0)
	zero := slot{node: &node4, index: 0}
	result := path.foldR(func(p, ch slot) slot {
		sum := p.item().value.(int) + ch.item().value.(int)
		//t.Logf("%2d <- p = %v, ch = %v", sum, p, ch)
		node := xnode{}.withInsertedItem(xitem{-1, sum}, 0)
		return slot{node: &node, index: 0}
	}, zero)
	if result.item().value.(int) != 10 {
		t.Logf("result of fold %v, %v = %v", path, zero.item(), result.item())
		t.Error("expected result of path.fold(+, 4) to be 10, isn't")
	}
}
