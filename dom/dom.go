package dom

import (
	"bytes"
	"fmt"

	"github.com/npillmayer/fp/dom/style"
	"github.com/npillmayer/fp/dom/style/css"
	"github.com/npillmayer/fp/dom/style/cssom"
	"github.com/npillmayer/fp/dom/style/cssom/douceuradapter"
	"github.com/npillmayer/fp/dom/styledtree"
	"github.com/npillmayer/fp/dom/w3cdom"
	"github.com/npillmayer/fp/tree"
	"golang.org/x/net/html"
)

// --- Node -----------------------------------------------------------------------

// A W3CNode is common type from which various types of DOM API objects inherit.
// This allows these types to be treated similarly.
type W3CNode struct {
	//stylednode *styledtree.StyNode
	*styledtree.StyNode
}

var _ w3cdom.Node = &W3CNode{}

// NodeFromStyledNode creates a new DOM node from a styled node.
func NodeFromStyledNode(sn *styledtree.StyNode) *W3CNode {
	return &W3CNode{sn}
}

// NodeFromTreeNode creates a new DOM node from a tree node, which should
// be the inner node of a styledtree.Node.
func NodeFromTreeNode(tn *tree.Node[*styledtree.StyNode]) (*W3CNode, error) {
	if tn == nil {
		return nil, ErrNotAStyledNode
	}
	w := domify(tn)
	if w == nil {
		return nil, ErrNotAStyledNode
	}
	return w, nil
}

// ErrNotAStyledNode is returned if a tree node does not belong to a styled tree node.
var ErrNotAStyledNode = fmt.Errorf("Tree node is not a styled node")

func domify(tn *tree.Node[*styledtree.StyNode]) *W3CNode {
	sn := styledtree.Node(tn)
	if sn != nil {
		return &W3CNode{sn}
	}
	return nil
}

// NodeAsTreeNode returns the underlying tree.Node from a DOM node.
func NodeAsTreeNode(domnode w3cdom.Node) (*tree.Node[*styledtree.StyNode], bool) {
	if domnode == nil {
		return nil, false
	}
	w, ok := domnode.(*W3CNode)
	if !ok {
		tracer().Errorf("DOM node has not been created from w3cdom.go")
		return nil, false
	}
	//return &w.stylednode.Node, true
	return &w.Node, true
}

// IsRoot returns wether this node is the root node of the styled tree.
func (w *W3CNode) IsRoot() bool {
	return w.ParentNode() == nil
}

// IsDocument returns wether this node is the document node of the styled tree.
func (w *W3CNode) IsDocument() bool {
	if w.ParentNode() == nil {
		return false
	}
	parent := w.ParentNode().(*W3CNode)
	return parent.HTMLNode() == w.HTMLNode()
}

// NodeType returns the type of the underlying HTML node, something like
// html.ElementNode, html.TextNode, etc.
func (w *W3CNode) NodeType() html.NodeType {
	if w == nil {
		return html.ErrorNode
	}
	return w.HTMLNode().Type
}

// NodeName read-only property returns the name of the current Node as a string.
//
//      Node         NodeName value
//      ------------+----------------------------
//      Attr         The value of Attr.name
//      Document     "#document"
//      Element      The value of Element.TagName
//      Text         "#text"
//
func (w *W3CNode) NodeName() string {
	if w == nil {
		return ""
	}
	h := w.HTMLNode()
	switch h.Type {
	case html.DocumentNode:
		return "#document"
	case html.ElementNode:
		return h.Data
	case html.TextNode:
		return "#text"
	}
	return "<node>"
}

// NodeValue returns textual content for text/CData-Nodes, and an empty string for any other
// Node type.
func (w *W3CNode) NodeValue() string {
	if w == nil {
		return ""
	}
	h := w.HTMLNode()
	if h.Type == html.TextNode {
		return h.Data
	}
	return ""
}

// HasAttributes returns a boolean indicating whether the current element has any
// attributes or not.
func (w *W3CNode) HasAttributes() bool {
	if w == nil {
		return false
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		return len(styledtree.Node(tn).HTMLNode().Attr) > 0
	}
	return false
}

// ParentNode read-only property returns the parent of the specified node in the DOM tree.
func (w *W3CNode) ParentNode() w3cdom.Node {
	if w == nil {
		return nil
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		p := tn.Parent()
		if p != nil {
			return domify(p)
		}
	}
	return nil
}

// HasChildNodes method returns a boolean value indicating whether the given Node
// has child nodes or not.
func (w *W3CNode) HasChildNodes() bool {
	if w == nil {
		return false
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		return tn.ChildCount() > 0
	}
	return false
}

// ChildNodes read-only property returns a live NodeList of child nodes of
// the given element.
func (w *W3CNode) ChildNodes() w3cdom.NodeList {
	if w == nil {
		return nil
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		children := tn.Children(true)
		childnodes := make([]*W3CNode, len(children))
		for i, ch := range children {
			childnodes[i] = &W3CNode{styledtree.Node(ch)}
		}
		return &W3CNodeList{childnodes}
	}
	return nil
}

// Children is a read-only property that returns a node list which contains all of
// the child *elements* of the node upon which it was called
func (w *W3CNode) Children() w3cdom.NodeList {
	if w == nil {
		return nil
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		children := tn.Children(true)
		childnodes := make([]*W3CNode, len(children))
		j := 0
		for _, ch := range children {
			sn := styledtree.Node(ch)
			if sn.HTMLNode().Type == html.ElementNode {
				childnodes[j] = &W3CNode{sn}
				j++
			}
		}
		return &W3CNodeList{childnodes}
	}
	return nil
}

// FirstChild read-only property returns the node's first child in the tree,
// or nil if the node has no children.
func (w *W3CNode) FirstChild() w3cdom.Node {
	if w == nil {
		return nil
	}
	tn, ok := NodeAsTreeNode(w)
	if ok && tn.ChildCount() > 0 {
		ch, ok := tn.Child(0)
		if ok {
			return domify(ch)
		}
	}
	return nil
}

// NextSibling read-only property returns the node immediately following the
// specified one in their parent's childNodes,
// or returns nil if the specified node is the last child in the parent element.
func (w *W3CNode) NextSibling() w3cdom.Node {
	if w == nil {
		return nil
	}
	tn, ok := NodeAsTreeNode(w)
	if ok {
		if parent := tn.Parent(); parent != nil {
			if i := parent.IndexOfChild(tn); i >= 0 {
				sibling, ok := parent.Child(i + 1)
				if ok {
					return domify(sibling)
				}
			}
		}
	}
	return nil
}

// Attributes property returns a collection of all attribute nodes registered
// to the specified node. It is a NamedNodeMap, not an array.
func (w *W3CNode) Attributes() w3cdom.NamedNodeMap {
	if w == nil {
		return emptyNodeMap
	}
	h := w.HTMLNode()
	switch h.Type {
	case html.DocumentNode:
	case html.ElementNode:
		return nodeMapFor(w.StyNode)
	}
	return emptyNodeMap
}

// TextContent property of the Node interface represents the text content of
// the node and its descendants.
//
// This implementation will include error strings in the text output, if errors occur.
// They will be flagged as "(ERROR: ... )".
func (w *W3CNode) TextContent() (string, error) {
	future := w.Walk().DescendentsWith(NodeIsText).Promise()
	textnodes, err := future()
	if err != nil {
		tracer().Errorf(err.Error())
		return "(ERROR: " + err.Error() + " )", err
	}
	var b bytes.Buffer
	var domnode *W3CNode
	for _, t := range textnodes {
		domnode, err = NodeFromTreeNode(t)
		if err != nil {
			b.WriteString("(ERROR: " + err.Error() + " )")
		} else {
			b.WriteString(domnode.NodeValue())
		}
	}
	return b.String(), err
}

// ComputedStyles returns a map of style properties for a given (stylable) Node.
func (w *W3CNode) ComputedStyles() w3cdom.ComputedStyles {
	if w == nil {
		return nil
	}
	return &computedStyles{w, w.Styles()}
}

// --- computed styles -------------------------------------------------------

// computedStyles is a little proxy type for a node's styles.
//
// TODO include pseudo-elements => implement
//
//    var style = window.getComputedStyle(element [, pseudoElt]);
//
// see https://developer.mozilla.org/de/docs/Web/API/Window/getComputedStyle :
//
// pseudoElt (Optional):
//     A string specifying the pseudo-element to match. Omitted (or null) for real elements.
//
// The returned style is a live CSSStyleDeclaration object, which updates automatically
// when the element's styles are changed.
//
type computedStyles struct {
	domnode  *W3CNode
	propsMap *style.PropertyMap
}

// Styles returns the underlying style.PropertyMap.
func (cstyles *computedStyles) Styles() *style.PropertyMap {
	return cstyles.propsMap
}

// HTMLNode returns the underlying html.Node.
func (cstyles *computedStyles) HTMLNode() *html.Node {
	return cstyles.domnode.HTMLNode()
}

/*
func (cstyles *computedStyles) StylesCascade() style.Styler {
	return cstyles.domnode.StylesCascade()
}

var _ style.Styler = &computedStyles{} // implementing style.Styler may be useful

// Helper implementing style.Interf
func styler(n *tree.Node[*styledtree.StyNode]) style.Styler {
	return styledtree.Node(n)
}
*/

// GetPropertyValue returns the property value for a given key.
// If cstyles is nil or the property could not be found, NullStyle is returned.
func (cstyles *computedStyles) GetPropertyValue(key string) style.Property {
	if cstyles == nil {
		return style.NullStyle
	}
	//p, err := css.GetProperty(cstyles.domnode.AsStyler(), key)
	p, err := css.GetProperty(cstyles.domnode.StyNode, key)
	if err != nil {
		tracer().Errorf("W3C node styles: %v", err)
		//return cstyles.propsMap.GetPropertyValue(key, node, styler)
		return cstyles.domnode.StyNode.GetPropertyValue(key, cstyles.propsMap)
	}
	return p
}

// --- Attributes -----------------------------------------------------------------

// A W3CAttr represents a single attribute of an element Node.
type W3CAttr struct {
	attr *html.Attribute
}

var _ w3cdom.Attr = &W3CAttr{}

// Namespace returns the namespace prefix of an attribute.
func (a *W3CAttr) Namespace() string {
	return a.attr.Namespace
}

// Key is the name of an attribute.
func (a *W3CAttr) Key() string {
	return a.attr.Key
}

// Value is the string value of an attribute.
func (a *W3CAttr) Value() string {
	return a.attr.Val
}

var _ w3cdom.Node = &W3CAttr{} // Attributes are W3C DOM nodes as well

// AttrNode is an additional node type, complementing those defined in
// standard-package html.
const AttrNode = html.NodeType(77)

// NodeName for an attribute is the attribute key
func (a *W3CAttr) NodeName() string {
	if a == nil {
		return ""
	}
	return a.attr.Key
}

// NodeValue for an attribute is the attribute value
func (a *W3CAttr) NodeValue() string {
	if a == nil {
		return ""
	}
	return a.attr.Val
}

// NodeType returns type AttrNode
func (a *W3CAttr) NodeType() html.NodeType { return AttrNode }

// HasAttributes returns false
func (a *W3CAttr) HasAttributes() bool { return false }

// HasChildNodes returns false
func (a *W3CAttr) HasChildNodes() bool { return false }

// ParentNode returns nil
func (a *W3CAttr) ParentNode() w3cdom.Node { return nil }

// ChildNodes returns nil
func (a *W3CAttr) ChildNodes() w3cdom.NodeList { return nil }

// Children returns nil
func (a *W3CAttr) Children() w3cdom.NodeList { return nil }

// FirstChild returns nil
func (a *W3CAttr) FirstChild() w3cdom.Node { return nil }

// NextSibling returns nil
func (a *W3CAttr) NextSibling() w3cdom.Node { return nil }

// Attributes returns nil
func (a *W3CAttr) Attributes() w3cdom.NamedNodeMap { return nil }

// TextContent returns an empty string
func (a *W3CAttr) TextContent() (string, error) { return "", nil }

// ComputedStyles gets null-styles
func (a *W3CAttr) ComputedStyles() w3cdom.ComputedStyles {
	return nullStyles{}
}

type nullStyles struct{}

func (nullStyles) GetPropertyValue(string) style.Property {
	return style.NullStyle
}

func (nullStyles) Styles() *style.PropertyMap {
	return nil
}

// --- NamedNodeMap ---------------------------------------------------------------

// A W3CMap represents a key-value map
type W3CMap struct {
	forNode *styledtree.StyNode
}

var _ w3cdom.NamedNodeMap = &W3CMap{}

var emptyNodeMap = &W3CMap{}

func nodeMapFor(sn *styledtree.StyNode) w3cdom.NamedNodeMap {
	if sn != nil {
		return &W3CMap{sn}
	}
	return nil
}

// Length returns the number of entries in a key-value map
func (wm *W3CMap) Length() int {
	if wm == nil {
		return 0
	}
	return len(wm.forNode.HTMLNode().Attr)
}

// Item returns the i.th item in a key-value map
func (wm *W3CMap) Item(i int) w3cdom.Attr {
	if wm == nil {
		return nil
	}
	attrs := wm.forNode.HTMLNode().Attr
	if len(attrs) <= i || i < 0 {
		return nil
	}
	return &W3CAttr{&attrs[i]}
}

// GetNamedItem returns the attribute with key key.
func (wm *W3CMap) GetNamedItem(key string) w3cdom.Attr {
	if wm == nil {
		return nil
	}
	attrs := wm.forNode.HTMLNode().Attr
	for _, a := range attrs {
		if a.Key == key {
			return &W3CAttr{&a}
		}
	}
	return nil
}

// --- NodeList -------------------------------------------------------------------

// A W3CNodeList is a type for a list of nodes
type W3CNodeList struct {
	nodes []*W3CNode
}

var _ w3cdom.NodeList = &W3CNodeList{}

// Length returns the number of Nodes in a list
func (wl *W3CNodeList) Length() int {
	if wl == nil {
		return 0
	}
	return len(wl.nodes)
}

// Item returns the i.th Node
func (wl *W3CNodeList) Item(i int) w3cdom.Node {
	if wl == nil {
		return nil
	}
	if i >= len(wl.nodes) || i < 0 {
		return nil
	}
	return wl.nodes[i]
}

func (wl *W3CNodeList) String() string {
	var s bytes.Buffer
	s.WriteString("[ ")
	if wl != nil {
		for _, n := range wl.nodes {
			s.WriteString(n.NodeName())
			s.WriteString(" ")
		}
	}
	s.WriteString("]")
	return s.String()
}

// --------------------------------------------------------------------------------

// FromHTMLParseTree returns a W3C DOM from parsed HTML and an optional style sheet.
func FromHTMLParseTree(h *html.Node, css cssom.StyleSheet) *W3CNode {
	if h == nil {
		tracer().Infof("Cannot create DOM for null-HTML")
		return nil
	}
	styles := douceuradapter.ExtractStyleElements(h)
	tracer().Debugf("Extracted %d <style> elements", len(styles))
	s := cssom.NewCSSOM(nil) // nil = no additional properties
	for _, sty := range styles {
		s.AddStylesForScope(nil, sty, cssom.Script)
	}
	if css != nil {
		s.AddStylesForScope(nil, css, cssom.Author)
	}
	stytree, err := s.Style(h, styledtree.Creator())
	if err != nil {
		tracer().Errorf("Cannot style test document: %s", err.Error())
		return nil
	}
	d := domify(stytree)
	return d
}

/*
// XPath creates an xpath navigator with start position w.
func (w *W3CNode) XPath() *xpath.XPath {
	if w == nil {
		return nil
	}
	nav := xpathadapter.NewNavigator(w.StyNode)
	xp, err := xpath.NewXPath(nav, xpathadapter.CurrentNode)
	if err != nil {
		tracer().Errorf("dom xpath: %v", err.Error())
		return nil
	}
	return xp
}
*/

// Walk creates a tree walker set up to traverse the DOM.
func (w *W3CNode) Walk() *tree.Walker[*styledtree.StyNode, *styledtree.StyNode] {
	if w == nil {
		return nil
	}
	return tree.NewWalker(&w.Node)
}
