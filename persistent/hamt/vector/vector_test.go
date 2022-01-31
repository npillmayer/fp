package vector

import (
	"fmt"
	"testing"

	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	tp "github.com/xlab/treeprint"
)

func TestVectorEmpty(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	defer teardown()
	//
	v := Vector[int]{}
	if v.Len() != 0 {
		t.Errorf("expected empty vector to have length 0, has %d", v.Len())
	}
	if x := v.Last().WithDefault(99); x != 99 {
		t.Error("expected empty vector to have last element of 'nothing', didn't")
	}
}

func TestVectorConstructor(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](DegreeExponent(2))
	if v.mask != 0x03 {
		t.Errorf("expected mask to be 0011, is %x", v.mask)
	}
}

func TestVectorPushTail(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](DegreeExponent(1))
	v = v.Push(77)
	if len(v.tail) != 1 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 1, is '%v'", v.tail)
	}
	v = v.Push(78)
	if len(v.tail) != 2 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 2, is '%v'", v.tail)
	}
	v = v.Push(80)
	if len(v.tail) != 1 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 1, is '%v'", v.tail)
	}
	v = v.Push(81)
	if len(v.tail) != 2 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 2, is '%v'", v.tail)
	}
	v = v.Push(90)
	if len(v.tail) != 1 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 1, is '%v'", v.tail)
	}
}

func TestVectorGet(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](DegreeExponent(1))
	v = v.Push(77)
	x := v.Get(0)
	if x != 77 {
		t.Logf(printVec(v))
		t.Errorf("expected 1st element of vector to be 77, isn't: %d", x)
	}
	v = v.Push(78).Push(79).Push(80).Push(81)
	//t.Logf(printVec(v))
	x = v.Get(2)
	if len(v.tail) != 1 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 1, is '%v'", v.tail)
	}
	if x != 79 {
		t.Logf(printVec(v))
		t.Errorf("expected 3rd element of vector to be 79, isn't: %d", x)
	}
}

func TestVectorPop(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "persistent.vector")
	//tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v := Immutable[int](DegreeExponent(1))
	v = v.Push(77).Push(78).Push(79).Push(80)
	//t.Logf(printVec(v))
	v = v.Pop()
	//t.Logf(printVec(v))
	v = v.Pop()
	//t.Logf(printVec(v))
	v = v.Push(79).Push(80).Push(81)
	//t.Logf(printVec(v))
	v = v.Pop()
	//t.Logf(printVec(v))
	if len(v.tail) != 2 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail to be of length 2, is '%v'", v.tail)
	}
	if v.tail[len(v.tail)-1] != 80 {
		t.Logf(printVec(v))
		t.Errorf("expected v.tail.last() to be 80, is '%v'", v.tail)
	}
}

// --- Print vector tree -----------------------------------------------------

func printVec[T any](v Vector[T]) string {
	header := fmt.Sprintf("\nVector(length=%d, shift=%x, degree=%d)\n", v.length, v.shift, v.degree)
	tail := fmt.Sprintf("       tail=%v\n", v.tail)
	printer := tp.New()
	h := v.degree
	if v.shift != 0 {
		h = v.degree / v.shift
	}
	printNode(printer, v.root, h, 0, v.degree)
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
