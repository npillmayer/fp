package tree

import (
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/npillmayer/schuko/schukonf/testconfig"
	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gologadapter"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	"github.com/npillmayer/schuko/tracing/trace2go"
)

func TestAddChild(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	parent := NewNode(-1)
	parent.AddChild(NewNode(0)).AddChild(NewNode(1))
	ch4 := NewNode(4)
	parent.SetChildAt(4, ch4)
	ch, _ := parent.Child(4)
	if ch == nil {
		t.Errorf("Inserted child at position 4 should have payload of 4, is nil")
	} else if ch != ch4 {
		t.Errorf("Inserted child at position 4 should have payload of 4, has %d", ch.Payload)
	}
	ch3 := NewNode(3)
	parent.InsertChildAt(1, ch3)
	ch, _ = parent.Child(1)
	if ch == nil {
		t.Errorf("Inserted child at position 1 should have payload of 3, is nil")
	} else if ch != ch3 {
		t.Errorf("Inserted child at position 1 should have payload of 3, has %d", ch.Payload)
	}
	ch, _ = parent.Child(5)
	if ch == nil {
		t.Errorf("Inserted child at position 5 should have payload of 4, is nil")
	} else if ch != ch4 {
		t.Errorf("Inserted child at position 5 should have payload of 4, has %d", ch.Payload)
	}
}

func TestEmptyWalker(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	w := NewWalker[int](nil)
	future := w.Parent().Promise()
	nodes, err := future()
	if err != nil {
		t.Log(err)
	} else {
		t.Error("Walker for empty tree should return an error")
	}
	if len(nodes) != 0 {
		t.Errorf("result set of empty pipeline should be empty")
	}
	checkRuntime(t, n)
}

func TestParentPredicate(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	node1, node2 := NewNode(1), NewNode(2)
	node1.AddChild(node2) // simple tree: (1)-->(2)
	w := NewWalker(node2)
	future := w.Parent().Promise()
	nodes, err := future()
	if err != nil {
		t.Error(err)
	}
	if len(nodes) != 1 || !checkNodes[int](nodes, 1) {
		t.Errorf("did not find parent, nodes = %v", nodes)
	}
	checkRuntime(t, n)
}

func TestParentOfRoot(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	node1 := NewNode(1)
	w := NewWalker(node1)
	future := w.Parent().Promise()
	nodes, err := future()
	if err != nil {
		t.Error(err)
	}
	if len(nodes) != 0 {
		t.Errorf("root should have no parent, nodes = %v", nodes)
	}
	checkRuntime(t, n)
}

func TestAncestor(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	node1, node2 := NewNode(1), NewNode(2)
	node1.AddChild(node2) // simple tree: (1)-->(2)
	w := NewWalker(node2)
	future := w.AncestorWith(Whatever[int]()).Promise()
	nodes, err := future()
	if err != nil {
		t.Error(err)
	}
	for _, node := range nodes {
		t.Logf("ancestor = %v\n", node)
	}
	if len(nodes) != 1 || !checkNodes[int](nodes, 1) {
		t.Errorf("did not find single ancestor (parent), nodes = %v", nodes)
	}
	checkRuntime(t, n)
}

func TestDescendents(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	node1, node2, node3, node4 := NewNode(1), NewNode(2), NewNode(3), NewNode(4)
	// build tree:
	// (1)
	//  +---(2)
	//  |    +---(3)
	//  |
	//  +---(4)
	node1.AddChild(node2)
	node2.AddChild(node3)
	node1.AddChild(node4)
	gr3 := func(node *Node[int], n *Node[int]) (*Node[int], error) {
		val := node.Payload
		if val >= 3 { // match nodes (3) and (4)
			return node, nil
		}
		return nil, nil
	}
	w := NewWalker(node1)
	future := w.DescendentsWith(gr3).Promise()
	nodes, err := future()
	if err != nil {
		t.Error(err)
	}
	for _, node := range nodes {
		t.Logf("descendent = %v\n", node)
	}
	if len(nodes) != 2 || !checkNodes[int](nodes, 3, 4) {
		t.Errorf("did not find descendents (3) and (4), nodes = %v", nodes)
	}
	checkRuntime(t, n)
}

func ExampleWalker_Promise() {
	// Build a tree:
	//
	//                 (root:1)
	//          (n2:2)----+----(n4:10)
	//  (n3:10)----+
	//
	// Then query for nodes with value > 5
	//
	root, n2, n3, n4 := NewNode(1), NewNode(2), NewNode(10), NewNode(10)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	// Define our ad-hoc predicate
	greater5 := func(node *Node[int], n *Node[int]) (*Node[int], error) {
		val := node.Payload
		if val > 5 { // match nodes with value > 5
			return node, nil
		}
		return nil, nil
	}
	// Now navigate the tree (concurrently)
	future := NewWalker(root).DescendentsWith(greater5).Promise()
	// Any time later call the promise ...
	nodes, err := future() // will block until walking is finished
	if err != nil {
		fmt.Print(err)
	}
	for _, node := range nodes {
		fmt.Printf("matching descendent found: (Node %d)\n", node.Payload)
	}
	// Output:
	// matching descendent found: (Node 10)
	// matching descendent found: (Node 10)
}

func TestTopDown1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	// Build a tree:
	//                 (root:1)
	//          (n2:2)----+----(n4:10)
	//  (n3:10)----+
	//
	root, n2, n3, n4 := NewNode(1), NewNode(2), NewNode(10), NewNode(10)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	i := 0
	myaction := func(n *Node[int], parent *Node[int], position int) (*Node[int], error) {
		tracer().Debugf("input node is %v", n)
		i++
		return n, nil
	}
	future := NewWalker(root).TopDown(myaction).Promise()
	_, err := future() // will block until walking is finished
	if err != nil {
		t.Error(err)
	}
	if i != 4 {
		t.Errorf("Expected action to be called 4 times, was %d", i)
	}
	checkRuntime(t, n)
}

func TestBottomUp1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	// Build a tree:
	//                 (root:3)
	//          (n2:2)----+----(n4:1)
	//  (n3:1)----+
	//
	root, n2, n3, n4 := NewNode(3), NewNode(2), NewNode(1), NewNode(1)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	i := 0
	nodevals := make([]int, 4)
	myaction := func(n *Node[int], parent *Node[int], position int) (*Node[int], error) {
		tracer().Debugf("node has value=%v", n.Payload)
		nodevals[i] = n.Payload
		i++
		return n, nil
	}
	future := NewWalker(n3).BottomUp(myaction).Promise()
	_, err := future() // will block until walking is finished
	if err != nil {
		t.Error(err)
	}
	if i != 2 { // nodes n4 and root should not be processed
		t.Logf("result values = %v", nodevals)
		t.Errorf("Expected action to be called 2 times, was %d", i)
	}
	checkRuntime(t, n)
}

func TestBottomUp2(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	// Build a tree:
	//                 (root:3)
	//          (n2:2)----+----(n4:1)
	//  (n3:1)----+
	//
	root, n2, n3, n4 := NewNode(3), NewNode(2), NewNode(1), NewNode(1)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	i := 0
	nodevals := make([]int, 6)
	myaction := func(n *Node[int], parent *Node[int], position int) (*Node[int], error) {
		tracer().Debugf("node has value=%v", n.Payload)
		nodevals[i] = n.Payload
		i++
		return n, nil
	}
	future := NewWalker(root).DescendentsWith(NodeIsLeaf[int]()).BottomUp(myaction).Promise()
	_, err := future() // will block until walking is finished
	if err != nil {
		t.Error(err)
	}
	if i != 4 { // all nodes should be processed
		t.Logf("(unreliable) result values = %v", nodevals)
		t.Errorf("Expected action to be called 4 times, was %d", i)
	}
	checkRuntime(t, n)
}

func TestRank(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	// Build a tree:
	//                 (root:3)
	//          (n2:2)----+----(n4:1)
	//  (n3:1)----+
	//
	root, n2, n3, n4 := NewNode(3), NewNode(2), NewNode(1), NewNode(1)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	future := NewWalker(root).DescendentsWith(NodeIsLeaf[int]()).BottomUp(CalcRank[int]).Promise()
	_, err := future() // will block until walking is finished
	if err != nil {
		t.Error(err)
	}
	if root.Rank != 4 || n2.Rank != 2 {
		t.Errorf("Rank of root node should be 4, is %d", root.Rank)
		t.Errorf("Rank of node n2 should be 2, is %d", n2.Rank)
	}
	checkRuntime(t, n)
}

func TestSerial1(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "tyse.frame.tree")
	defer teardown()
	// configureGoTracing(t)
	//
	n := checkRuntime(t, -1)
	// Build a tree:
	//                 (root:6)
	//          (n2:2)----+----(n4:5)
	//  (n3:1)----+        (n5:3)-+--(n6:4)
	//
	root, n2, n3, n4 := NewNode(6), NewNode(2), NewNode(1), NewNode(5)
	n5, n6 := NewNode(3), NewNode(4)
	root.AddChild(n2).AddChild(n4)
	n2.AddChild(n3)
	n4.AddChild(n5).AddChild(n6)
	// calculate rank for each node
	NewWalker(root).DescendentsWith(NodeIsLeaf[int]()).BottomUp(CalcRank[int]).Promise()()
	if root.Rank != 6 {
		t.Errorf("Rank of root node should be 6, is %d", root.Rank)
	}
	myaction := func(n *Node[int], parent *Node[int], position int) (*Node[int], error) {
		t.Logf("rank of node(%d) is %d", n.Payload, n.Rank)
		return n, nil
	}
	future := NewWalker(root).TopDown(myaction).Promise()
	nodes, err := future() // will block until walking is finished
	if err != nil {
		t.Error(err)
	}
	z := 0
	for i, n := range nodes {
		t.Logf("node #%d is (%v) with rank %d", i, n.Payload, n.Rank)
		z = z<<4 + n.Payload
	}
	if z != 1193046 {
		t.Errorf("checksum = %d, should be 1193046", z)
	}
	checkRuntime(t, n)
}

// ----------------------------------------------------------------------

// Helper to check if result nodes are the expected ones.
func checkNodes[T comparable](nodes []*Node[int], vals ...int) bool {
	var found bool
	for _, node := range nodes {
		found = false
		v := node.Payload
		for _, val := range vals {
			if v == val {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Helper to check for leaked goroutines.
func checkRuntime(t *testing.T, N int) int {
	if N < 1 {
		n := runtime.NumGoroutine()
		t.Logf("pre-test %d goroutines are alive", n)
		return n
	}
	time.Sleep(10 * time.Millisecond)
	n := runtime.NumGoroutine()
	if n > N {
		t.Logf("still %d goroutines alive", n)
		if N != n {
			t.Fail()
		}
	}
	return n
}

func configureGoTracing(t *testing.T) {
	tracing.RegisterTraceAdapter("go", gologadapter.GetAdapter(), false)
	conf := &testconfig.Conf{}
	conf.Set("tracing", "go")
	conf.Set("trace.tyse.frame.tree", "Debug")
	if err := trace2go.ConfigureRoot(conf, "trace", trace2go.ReplaceTracers(true)); err != nil {
		t.Error(err)
	}
	tracing.SetTraceSelector(trace2go.Selector())
	tracer().Debugf("testing: DEBUG ok")
}
