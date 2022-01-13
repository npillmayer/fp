package btree

import (
	"fmt"
	"strconv"
	"strings"
)

// slot holds a step of a path.
// A slot is treated to be immutable after creation.
type slot struct {
	node  *xnode
	index int
}

func (slot slot) String() string {
	return strconv.Itoa(slot.index) + "@" + slot.node.String()
}

func (slot slot) item() Item {
	return slot.node.items[slot.index]
}

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

func (path slotPath) First() slot {
	if len(path) == 0 {
		return slot{}
	}
	return path[0]
}

// Map is destructive !
func (path slotPath) Map(apply func(slot) slot) slotPath {
	for i, slot := range path {
		path[i] = apply(slot)
	}
	return path
}

// Reverse is destructive !
func (path slotPath) Reverse() slotPath {
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}
