package either

import (
	"fmt"
)

// ---------------------------------------------------------------------------

type EitherType interface {
	isEither()
}

type LeftType int
func (e LeftType) isEither() {}

type RightType string
func (e RightType) isEither() {}

func MakeLeft(n int) EitherType {
	return LeftType(n)
}

func MakeRight(s string) EitherType {
	return RightType(s)
}

// ---------------------------------------------------------------------------

type left int
func (e left) isEither() {}

type right string
func (e right) isEither() {}

func ELeft(n interface{}) EitherType {
	if i, ok := n.(int); ok {
		return left(i)
	}
	return left(0)
}

func ERight(s interface{}) EitherType {
	if x, ok := s.(string); ok {
		return right(x)
	}
	return right("")
}

func Match(e EitherType, ps Patterns) interface{} {
	for _, p := range ps {
		pp := p.Pattern(nil)
		fmt.Printf("pp = %#v\n", pp)
		if l, ok := pp.(left); ok {
			fmt.Printf("l  = %#v\n", l)
			switch ppp := p.Value.(type) {
			case left:
				return ppp
			case func(int) int:
				a := e.(left)
				n := ppp(int(a))
				return left(n)
			}
		} else if r, ok := pp.(right); ok {
			fmt.Printf("r  = %#v\n", r)
			switch ppp := p.Value.(type) {
			case right:
				return ppp
			case func(string) int:
				a := e.(right)
				n := ppp(string(a))
				return left(n)
			}
		}
	}
	panic(fmt.Sprintf("cannot match either-case for %#v", e))
}

type Patterns []struct {
	Pattern func(interface{}) EitherType
	Value interface{}
}

// ---------------------------------------------------------------------------

// Clients should be able to define a sum type:
//
// Haskell:
//
//     type Either a b = Left a | Right b
//
// Pseudo-Go:
//
//     type Either[A, B any] union {
//         Left  A
//         Right B
//     }
//
// Stand-in in Go:
//
type EitherSum[L, R any] struct {
	discr bool
	LeftField L
	RightField R
}

type Type interface {
	Left(int)     int
	Right(string) string
}

func (sum EitherSum[L, R]) Left(l L) L {
	return l
}

func (sum EitherSum[L, R]) Right(r R) R {
	return r
}

func Left(n int) Type {
	return EitherSum[int, string]{ LeftField: n }
}

func Right(s string) Type {
	return EitherSum[int, string]{ discr: true, RightField: s }
}

//type 
