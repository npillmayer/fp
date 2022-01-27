package vector

import (
	"fmt"
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	tp "github.com/xlab/treeprint"
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
	t.Logf(printVec(v))
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
	w, path := v.location(39, nil)
	t.Logf("path=%v", path)
	t.Logf(printVec(w))
}

// --- Print tree ------------------------------------------------------------

func printVec[T any](v Vector[T]) string {
	header := fmt.Sprintf("\nVector(len=%d, depth=%d, k=%d)\n", v.len, v.depth, v.bucketSize)
	printer := tp.New()
	ppt(printer, v.head, v.depth, 0, v.bucketSize)
	return header + printer.String() + "\n"
}

func ppt[T any](printer tp.Tree, node *vnode[T], h, j, k int) {
	if node == nil {
		return
	}
	if node.leaf {
		pp := p(k, h)
		//j = j / pp
		printer.AddNode(node.String() + fmt.Sprintf("%d  %d…%d", pp, j, j+pp-1))
		return
	}
	pp := p(k, h)
	branch := printer.AddBranch(node.String() + fmt.Sprintf("%d  %d…%d", pp, j, j+pp-1))
	//j = j / pp
	pp = p(k, h-1)
	for i, ch := range node.children {
		ppt(branch, ch, h-1, (i*pp)+j, k)
	}
}
