package css

import (
	"github.com/npillmayer/tyse/core/dimen"
	. "github.com/npillmayer/tyse/core/percent"
)

const (
	dimenNone uint32 = 0

	dimenAbsolute uint32 = 0x0001
	dimenAuto     uint32 = 0x0002
	dimenInherit  uint32 = 0x0003
	dimenInitial  uint32 = 0x0004
	kindMask      uint32 = 0x000f

	// Flags for content dependent dimensions
	DimenContentMax uint32 = 0x0010
	DimenContentMin uint32 = 0x0020
	DimenContentFit uint32 = 0x0030
	contentMask     uint32 = 0x00f0

	dimenEM      uint32 = 0x0100
	dimenEX      uint32 = 0x0200
	dimenCH      uint32 = 0x0300
	dimenREM     uint32 = 0x0400
	dimenVW      uint32 = 0x0500
	dimenVH      uint32 = 0x0600
	dimenVMIN    uint32 = 0x0700
	dimenVMAX    uint32 = 0x0800
	dimenPercent uint32 = 0x0900
	relativeMask uint32 = 0xff00
)

// DimenT is an option type for CSS dimensions.
type DimenT struct {
	d       dimen.DU
	percent Percent
	flags   uint32
}

/*
type DimenT
	= Auto
	| Inherit
	| Initial
	| JustDimen dimen
	| Percentage Percent
	| ViewRel unit
	| FontRel unit
	| ContentRel Min N
	| ContentRel Max N
*/

func Auto() DimenT {
	return DimenT{flags: dimenAuto}
}

func Inherit() DimenT {
	return DimenT{flags: dimenInherit}
}

func Initial() DimenT {
	return DimenT{flags: dimenInitial}
}

// JustDimen creates a CSS dimension with a fixed value of x.
func JustDimen(x dimen.DU) DimenT {
	return DimenT{d: x, flags: dimenAbsolute}
}

// Percentage creates a CSS dimension with a %-relative value.
func Percentage(n Percent) DimenT {
	return DimenT{percent: n, flags: dimenPercent}
}

// ---------------------------------------------------------------------------

func (d DimenT) Match() *Matcher {
	return &Matcher{dimen: d}
}

type Matcher struct {
	dimen DimenT
}

func (m *Matcher) IsKind(d DimenT) *Matcher {
	switch {
	case (m.dimen.flags & kindMask) == (d.flags & kindMask):
		return m
	case (m.dimen.flags&relativeMask > 0) && (d.flags&relativeMask > 0):
		if (m.dimen.flags&dimenPercent > 0) != (d.flags&dimenPercent > 0) {
			return nil
		}
		return m
	case (m.dimen.flags&contentMask > 0) && (d.flags&contentMask > 0):
		return m
	}
	return nil
}

func (m *Matcher) Just(du *dimen.DU) *Matcher {
	if m.dimen.flags&dimenAbsolute > 0 {
		if du != nil {
			*du = m.dimen.d
		}
		return m
	}
	return nil
}

func (m *Matcher) Percentage(p *Percent) *Matcher {
	if m.dimen.flags&dimenPercent > 0 {
		if p != nil {
			*p = m.dimen.percent
		}
		return m
	}
	return nil
}

// --- Expression matching ---------------------------------------------------

//type DimenPatterns[T any] map[*MatchExpr[T]]T
type DimenPatterns[T any] struct {
  Auto T
  Inherit T
  Initial T
  Just T
  Default T
}

func DimenPattern[T any](d DimenT) *MatchExpr[T] {
  return &MatchExpr[T]{ dimen: d }
}

type MatchExpr[T any] struct {
  dimen DimenT
}

func (m *MatchExpr[T]) OneOf(patterns DimenPatterns[T]) T {
  switch {
  case m.dimen.flags&dimenAuto > 0:
	return patterns.Auto
  case m.dimen.flags&dimenAbsolute > 0:
	return patterns.Just
  case m.dimen.flags&dimenInitial > 0:
	return patterns.Initial
  case m.dimen.flags&dimenInherit > 0:
	return patterns.Inherit
  }
  return patterns.Default
}

func (m *MatchExpr[T]) With(du *dimen.DU) *MatchExpr[T] {
  *du = m.dimen.d
  return m
}

func (m *MatchExpr[T]) Const(x T) T {
  return x
}

/*
func (m *MatchExpr[T]) Just(du *dimen.DU) *MatchExpr[T] {
	if m.dimen.flags&dimenAbsolute > 0 {
		if du != nil {
			*du = m.dimen.d
		}
		return m
	}
	return nil
}

func (m *MatchExpr[T]) IsKind(d DimenT) *MatchExpr[T] {
  return nil // TODO see above
}

func (m *MatchExpr[T]) Default() *MatchExpr[T] {
  return m
}

*/
