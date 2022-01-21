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