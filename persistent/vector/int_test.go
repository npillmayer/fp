package vector

import (
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
)

func TestCapacity(t *testing.T) {
	if p(4, 0) != 0 {
		t.Errorf("expected capacity(4, 0) to be 0, is %d", p(4, 0))
	}
	if p(4, 1) != 4 {
		t.Errorf("expected capacity(4, 0) to be 4, is %d", p(4, 1))
	}
	if p(4, 2) != 16 {
		t.Errorf("expected capacity(4, 2) to be 16, is %d", p(4, 2))
	}
	if p(4, 3) != 4*4*4 {
		t.Errorf("expected capacity(4, 3) to be %d, is %d", 4*4*4, p(4, 3))
	}
}

func TestFindPath(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	k := 4
	root := nodes[int](k)
	node := leafs[int](k)
	root.children[3] = &node
	node.leafs[2] = 14

	v := Vector[int]{depth: 2, head: &root, bucketSize: k}
	found, inx, path := v.findPath(14, nil)
	t.Logf("found=%v, inx=%d, path=%v", found, inx, path)
}

func TestLocation1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	k := 4
	v := Vector[int]{bucketSize: k}
	var path slotPath[int]
	v, path = v.location(14, path)
	t.Logf("path=%v", path)
	v, path = v.location(39, nil)
	t.Logf("path=%v", path)
}

func TestLocation2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	k := 4
	root := nodes[int](k)
	node := leafs[int](k)
	root.children[3] = &node
	node.leafs[2] = 14

	v := Vector[int]{depth: 2, head: &root, bucketSize: k}
	_, path := v.location(39, nil)
	t.Logf("path=%v", path)
}
