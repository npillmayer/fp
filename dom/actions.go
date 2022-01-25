package dom

import (
	"github.com/npillmayer/fp/dom/styledtree"
	"github.com/npillmayer/fp/tree"
)

// NodeIsText is a predicate to match text-nodes of a DOM.
// It is intended to be used in a tree.Walker.
var NodeIsText = func(n *tree.Node[*styledtree.StyNode], unused *tree.Node[*styledtree.StyNode]) (
	match *tree.Node[*styledtree.StyNode], err error) {
	//
	domnode, err := NodeFromTreeNode(n)
	if err != nil {
		return nil, err
	}
	if domnode.NodeName() == "#text" {
		return n, nil
	}
	return nil, nil
}
