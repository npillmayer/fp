package fp

// --- Matchable -------------------------------------------------------------

// Matchable is an interface for type which can be pattern-matched.
type Matchable[T, A, B comparable] interface {
  Matches(other T) bool
  Decompose() (A, B)
}

// --- Pair ------------------------------------------------------------------

type Pair[A, B comparable] struct {
  Left A
  Right B
}

func P[A, B comparable](x A, y B) Pair[A, B] {
  return Pair[A, B]{x, y}
}

func (p Pair[A, B]) Matches(other Pair[A, B]) bool {
  return p.Left == other.Left && p.Right == other.Right
}

func (p Pair[A, B]) Decompose() (A, B) {
  return p.Left, p.Right
}

var _ Matchable[Pair[int, int], int, int] = Pair[int, int]{1, 2}
var _ Matchable[Pair[int, int], int, int] = P(1, 2)

// --- Match -----------------------------------------------------------------

type PatternList[T, M, A, B comparable] []struct{
  eval func(M, A, B) (bool, A, B)
  result func(A, B) Option[T]
}

func Match[T, M, A, B comparable](ps PatternList[T, M, A, B]) T {
  var x T
  return x
}
