package vector

import (
	"fmt"
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	tp "github.com/xlab/treeprint"
)

func TestVectorConstructor(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](BitsPerLevel(2))
	if v.mask != 0x03 {
		t.Errorf("expected mask to be 0011, is %x", v.mask)
	}
}

func TestVectorPush1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](BitsPerLevel(1))
	v = v.Push(77)
	if len(v.tail) != 7 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 1, is '%v'", v.tail)
	}
	v = v.Push(78)
	if len(v.tail) != 7 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 2, is '%v'", v.tail)
	}
	v = v.Push(80)
	if len(v.tail) != 2 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 2, is '%v'", v.tail)
	}
}

// --- Print vector tree -----------------------------------------------------

func printVec[T any](v Vector[T]) string {
	header := fmt.Sprintf("\nVector(length=%d, shift=%x, degree=%d)\n", v.length, v.shift, v.degree)
	tail := fmt.Sprintf("       tail=%v\n", v.tail)
	printer := tp.New()
	printNode(printer, v.root, v.shift, 0, v.degree)
	return header + tail + printer.String() + "\n"
}

func printNode[T any](printer tp.Tree, node *vnode[T], h, j, k uint32) {
	if node == nil {
		return
	}
	if node.leafs != nil {
		pp := capacity(k, h)
		printer.AddNode(node.String() + fmt.Sprintf("%d  %d…%d", pp, j, j+pp-1))
		return
	}
	pp := capacity(k, h)
	branch := printer.AddBranch(node.String() + fmt.Sprintf("%d  %d…%d", pp, j, j+pp-1))
	pp = capacity(k, h-1)
	for i, ch := range node.children {
		printNode(branch, ch, h-1, (uint32(i)*pp)+j, k)
	}
}

func capacity(k, height uint32) uint32 {
	if height == 0 {
		return 0
	}
	c := k
	for height > 1 {
		c *= k
		height--
	}
	return c
}
