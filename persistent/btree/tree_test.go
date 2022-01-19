package btree

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/npillmayer/schuko/tracing"
	"github.com/npillmayer/schuko/tracing/gotestingadapter"
	tp "github.com/xlab/treeprint"
)

func TestTreeCreateEmptyTree(t *testing.T) {
	tree := Immutable(Degree(2))
	if tree.lowWaterMark != 2 || tree.highWaterMark != 6 {
		t.Logf("empty tree =\n%s", printTree(tree))
		t.Error("expected empty tree to have water marks 2 | 6, hasn't")
	}
}

func TestTreeCreateTreeForTest(t *testing.T) {
	tree := createTreeForTest()
	if tree.root == nil {
		t.Error("cannot create tree for test")
	}
	t.Logf("tree for tests =\n%s", printTree(tree))
	if tree.lowWaterMark != defaultLowWaterMark || tree.highWaterMark != defaultHighWaterMark {
		t.Error("expected test tree to have default water marks, hasn't")
	}
}

func TestTreeFindPathInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	defer teardown()
	//
	tree := Tree{}
	_, path := tree.findKeyAndPath(7, nil)
	if len(path) > 0 {
		t.Errorf("expected path for 7 to be nil, is %v", path)
	}
}

func TestTreeFindKeyAndPath(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	found, path := tree.findKeyAndPath(9, nil)
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

// --- Find ------------------------------------------------------------------

func TestTreeFindInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	v, found := Tree{}.Find(7)
	if found {
		t.Error("did not expect to find '7' in empty tree")
	}
	if v != nil {
		t.Errorf("expected value for '7' in empty tree to be void, is %v", v)
	}
}

func TestTreeFindInTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	v, found := tree.Find(8)
	if !found {
		t.Error("expected to find '8' in tree, didn't")
	}
	if v != "8" {
		t.Errorf("expected value for '8' in empty tree to be %#v, is %#v", "8", v)
	}
}

// --- Insert ----------------------------------------------------------------

func TestTreeInsertInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := Tree{}.With(7, "7")
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

func TestTreeInsertTwiceInEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := Tree{}.With(7, "7")
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	if tree.depth != 1 {
		t.Logf("tree = %#v", tree)
		t.Errorf("expected tree to have depth = 1, has %d", tree.depth)
	}
	if tree.lowWaterMark != defaultLowWaterMark {
		t.Errorf("expected tree to have low water mark of %d, has %d", defaultLowWaterMark, tree.lowWaterMark)
	}
	tree = tree.With(3, "3")
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	if tree.depth != 1 {
		t.Logf("tree = %#v", tree)
		t.Errorf("expected tree to have depth = 1, has %d", tree.depth)
	}
}

func TestTreeInsertInLeaf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	tree = tree.With(7, "7")
	if tree.root == nil {
		t.Fatalf("expected to have tree.With(…) to have a root, hasn't:\n%#v", tree)
	}
	if tree.depth != 2 {
		t.Logf("tree = %#v", tree)
		t.Logf("tree =\n%s", printTree(tree))
		t.Errorf("expected tree to have depth = 2, has %d", tree.depth)
	}
	if tree.lowWaterMark != defaultLowWaterMark {
		t.Logf("tree = %#v", tree)
		t.Logf("tree =\n%s", printTree(tree))
		t.Errorf("expected tree to have low water mark of %d, has %d", defaultLowWaterMark, tree.lowWaterMark)
	}
	ch2 := tree.root.children[2]
	if ch2 == nil || len(ch2.items) != 4 {
		t.Logf("tree = %s", printTree(tree))
		t.Fatalf("expected node root->2 to be of length=4, isn't")
	} else if ch2.items[1].key != K(7) {
		t.Logf("tree = %s", printTree(tree))
		t.Errorf("expected inserted item[1] to have key=7, is %#v", ch2.items[2])
	}
}

func TestTreeInsertWithSplit(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	tree.highWaterMark = 4
	tree = tree.With(7, "7")
	tree = tree.With(99, "99") // should trigger overfull(highWaterMark) -> split
	if tree.root == nil || tree.depth != 2 {
		t.Logf("tree = %s", printTree(tree))
		t.Fatalf("unexpected tree shape after insert of 7 and 99")
	}
	if len(tree.root.children) != 4 {
		t.Logf("tree = %s", printTree(tree))
		t.Fatalf("expected 4 root->children, have %d", len(tree.root.children))
	}
	ch4 := tree.root.children[3]
	if ch4 == nil || len(ch4.items) != 2 {
		t.Logf("tree = %s", printTree(tree))
		t.Fatalf("expected node root->child.3 to be of length=2, isn't")
	} else if ch4.items[1].key != K(99) {
		t.Logf("tree = %s", printTree(tree))
		t.Errorf("expected inserted child.3.item[1] to have key=7, is %#v", ch4.items[1])
	}
}

// --- Delete ----------------------------------------------------------------

func TestTreeDeleteFromEmptyTree(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := Tree{}.WithDeleted(7)
	if tree.root != nil {
		t.Logf("tree = %#v", tree)
		t.Logf("tree =\n%s", printTree(tree))
		t.Errorf("expected to have without a root")
	}
	if tree.depth != 0 {
		t.Errorf("expected tree.depth to be 0, is %d", tree.depth)
	}
}

func TestTreeDeleteInsertedKeyFromLeaf(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	//t.Logf("tree = %s", printTree(tree))
	modified := tree.With(7, "7")
	//t.Logf("tree = %s", printTree(modified))
	modified = modified.WithDeleted(7)
	orig := printTree(tree)
	mod := printTree(modified)
	if orig != mod {
		t.Log(orig)
		t.Log(mod)
		t.Errorf("different trees after insert+delete; expected to be equal")
	}
}
func TestTreeDeleteAndMerge(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	//t.Logf("tree =\n%s", printTree(tree))
	tree = tree.WithDeleted(9)
	if tree.depth != 2 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Errorf("expected tree to have depth=2, has %d", tree.depth)
	}
	ch := tree.root.children
	if len(ch) != 2 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Fatalf("expected root to have 2 children, has %d", len(ch))
	}
	if len(ch[1].items) != 5 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Fatalf("expected right child to have 5 items, has %d", len(ch[1].items))
	}
	if ch[1].items[2].key != 5 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Fatalf("expected right child to have middle item 5, has %v", ch[1].items[2].key)
	}
}

func TestTreeDeleteInnerItem(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	tree = tree.WithDeleted(5)
	if tree.depth != 2 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Errorf("expected tree to have depth=2, has %d", tree.depth)
	}
	if len(tree.root.children) != 2 {
		t.Logf("tree =\n%s", printTree(tree))
		t.Fatalf("expected child 1 and 2 of root to be merged, haven't")
	}
	//t.Logf("tree =\n%s", printTree(tree))
}

func TestTreeExtFind(t *testing.T) {
	teardown := gotestingadapter.QuickConfig(t, "fp.btree")
	tracer().SetTraceLevel(tracing.LevelError)
	defer teardown()
	//
	tree := createTreeForTest()
	tex := tree.Ext(_DefaultExtension{})
	loc := tex.Locate(6)
	if !loc.present {
		t.Logf("tree =\n%s", printTree(tree))
		t.Error("expected key 6 to be found in tree, haven't")
	}
}

// ---------------------------------------------------------------------------

func createTreeForTest() Tree { // tree with values 0…9, without 7
	root := &xnode{}
	root.add(2, 5)

	child0 := &xnode{}
	child0.add(0, 1)
	root.children = append(root.children, child0)

	child1 := &xnode{}
	child1.add(3, 4)
	root.children = append(root.children, child1)

	child2 := &xnode{}
	child2.add(6, 8, 9) // 7 is missing
	root.children = append(root.children, child2)

	//return newTreeWithRoot(root, minItems)
	return Tree{
		root:          root,
		depth:         2,
		lowWaterMark:  defaultLowWaterMark,
		highWaterMark: defaultHighWaterMark,
	}
}

func (node *xnode) add(keys ...K) *xnode {
	for _, key := range keys {
		node.items = append(node.items, xitem{key, T(strconv.Itoa(int(key)))})
	}
	return node
}

// ---------------------------------------------------------------------------

func printTree(tree Tree) string {
	header := fmt.Sprintf("\nTree(depth=%d ⊥%d ⊤%d)\n", tree.depth, tree.lowWaterMark, tree.highWaterMark)
	p := tp.New()
	ppt(p, tree.root)
	return header + p.String() + "\n"
}

func ppt(p tp.Tree, node *xnode) {
	if node == nil {
		return
	}
	if node.isLeaf() {
		p.AddNode(node.String())
		return
	}
	branch := p.AddBranch(node.String())
	for _, ch := range node.children {
		ppt(branch, ch)
	}
}
