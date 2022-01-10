package either_test

import (
  "testing"
  "strconv"
  "github.com/npillmayer/fp/either"
)

func TestEitherMatchType(t *testing.T) {
	one := either.MakeLeft(1)
	t.Logf("one = %v", one)
	var count int
	// Matching on type of one
	switch x := one.(type) {
	case either.LeftType:
		count = int(x)
	case either.RightType:
		count = Atoi(string(x))
	}
	t.Logf("count = %d", count)
}

func TestEitherMatchConstructor1(t *testing.T) {
	one := either.ELeft(1)
	t.Logf("one = %#v", one)
	// Matching on constructor
	count := either.Match(one, either.Patterns{
		{ either.ELeft,  one },
		{ either.ERight, Atoi },
	})
	t.Logf("count = %d", count)
}

func TestEitherMatchConstructor2(t *testing.T) {
	two := either.ERight("2")
	t.Logf("two = %#v", two)
	// Matching on constructor
	count := either.Match(two, either.Patterns{
		{ either.ELeft,  two },
		{ either.ERight, Atoi },
	})
	t.Logf("count = %d", count)
}

func TestGenericEither(t *testing.T) {
	one := either.Left(1)
	t.Logf("one = %#v", one)
	var count int
	// count = one.LeftField // this should result in a compiler error
	// Pattern matching on constructors
	var a int
	var s string
	switch x := one.(type) {
	case one.Left(a):
		count = a
	case one.Right(s):
		count = Atoi(string(s))
	}
	t.Logf("count = %d", count)
}

// ---------------------------------------------------------------------------

func Atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
