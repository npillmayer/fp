package fp

// Unit returns unit for any input => the zero value for T.
func Unit[T any](_ T) T {
  var a T
  return a
}

// Const returns a function that produces a.
func Const[T any](a T) func() T {
  return func() T {
	return a
  }
}

// Compose returns h = f . g
func Compose[A, B, C any](g func(a A) B, f func(b B) C) func(A) C {
	return func(a A) C {
		b := g(a)
		return f(b)
	}
}

// --- Option type -----------------------------------------------------------

type Option[T comparable] struct {
  just func() T
}

func Maybe[T comparable](a T) Option[T] {
  var nothing T
  if a == nothing {
	return Option[T]{nil}
  }
  return Option[T]{Const(a) }
}

func (o Option[T]) Just() T {
  return o.just()
}

func (o Option[_]) IsNone() bool {
  return o.just == nil
}

// --- Error option type -----------------------------------------------------

type ErrOption[T any] struct {
  just func() T
  err error
}

func Success[T any](a T) ErrOption[T] {
  return ErrOption[T]{just:Const(a) }
}

func Failure[T any](err error) ErrOption[T] {
  return ErrOption[T]{err:err}
}

func (o ErrOption[T]) Just() T {
  return o.just()
}

func (o ErrOption[_]) Error() error {
  return o.err
}

func (o ErrOption[_]) IsErr() bool {
  return o.err != nil
}

// --- Monoid ----------------------------------------------------------------

type Monoid interface {
  Empty() Monoid
  Append(Monoid) Monoid
}

func MConcat(m Monoid, ms ...Monoid) Monoid {
  for _, mm := range ms {
	m = m.Append(mm)
  }
  return m
}

type Sequence[S []any] struct {
  seq S
}

func (_ Sequence[S]) Empty() Sequence[S] {
  var s Sequence[S]
  return s
}

/*
func (s Sequence[S]) Append(other Sequence[S]) Sequence[S] {
  l := len(s.seq)
  var t Sequence[S]
  if other == nil {
  }
  t := Sequence[S]{seq: append(s.seq, other.seq...) }
  if t.seq != s.seq {
	return t
  }
  s.seq = t.seq[:l]
}
*/
