package btree

import (
	"fmt"
	"strconv"
	"strings"
)

// --- Slot ------------------------------------------------------------------

// slot holds a step of a path.
type slot struct {
	node  *xnode
	index int
}

func (s slot) String() string {
	return strconv.Itoa(s.index) + "@" + s.node.String()
}

func (s slot) replaceItem(item xitem) xitem {
	assertThat(s.index < len(s.node.items), "internal inconsistency: item index overflow")
	old := s.node.items[s.index]
	s.node.items[s.index] = item
	return old
}

func (s slot) leftSibling(child slot) slot {
	if s.node == nil || len(s.node.children) == 0 || s.index == 0 {
		return slot{}
	}
	assertThat(s.index <= len(s.node.children), "internal inconsistency: item index overflow")
	lsib := s.node.children[s.index-1]
	return slot{node: lsib, index: len(lsib.items)}
}

func (s slot) rightSibling(child slot) slot {
	if s.node == nil || len(s.node.children) == 0 || s.index == len(s.node.children) {
		return slot{}
	}
	assertThat(s.index <= len(s.node.children), "internal inconsistency: item index overflow")
	rsib := s.node.children[s.index+1]
	return slot{node: rsib, index: len(rsib.items)}
}

// siblings2 returns child and a non-void sibling as a correctly ordered pair.
// If child is an only child, a pair with an empty right sibling will be returned.
func (s slot) siblings2(child slot) (siblings [2]slot) {
	sbl := s.leftSibling(child)
	siblings[0], siblings[1] = sbl, child
	if sbl.node == nil {
		sbl = s.rightSibling(child)
		siblings[0], siblings[1] = child, sbl
	}
	assertThat(siblings[0].node != nil, "sibling-pair needs to have non-empty left sibling")
	return
}

func (s slot) item() xitem {
	return s.node.items[s.index]
}

// items returns a slice of items contained in s.node. If s is an empty slot (no node
// contained), a valid zero-length slice is returned (i.e., making it safe to call
// `s.items()`` for empty slots).
func (s slot) items() []xitem {
	if s.node == nil {
		return []xitem{}
	}
	return s.node.items
}

func (s slot) len() int {
	if s.node == nil {
		return 0
	}
	return len(s.node.items)
}

func (s slot) underfull(lowWaterMark uint) bool {
	if s.node == nil {
		return true
	}
	return s.node.underfull(lowWaterMark)
}

// --- Path ------------------------------------------------------------------

type slotPath []slot

func (path slotPath) String() string {
	var sb = strings.Builder{}
	sb.WriteRune('[')
	for _, s := range path {
		sb.WriteString(fmt.Sprintf("⟨%s⟩", s))
	}
	sb.WriteRune(']')
	return sb.String()
}

func (path slotPath) last() slot {
	if len(path) == 0 {
		return slot{}
	}
	return path[len(path)-1]
}

func (path slotPath) foldR(f func(slot, slot) slot, zero slot) slot {
	if len(path) == 0 {
		return zero
	}
	r := zero
	for i := len(path) - 1; i >= 0; i-- {
		r = f(path[i], r)
	}
	return r
}

func (path slotPath) dropLast() slotPath {
	if len(path) == 0 {
		return path
	}
	return path[:len(path)-1]
}
