package fp_test

import (
  "testing"
  "fmt"
  "errors"
  "github.com/npillmayer/fp"
)

func TestComposition(t *testing.T) {
	g := func(n int) float32 {
		return float32(n) + 0.5
	}
	f := func(x float32) string {
		return fmt.Sprintf("%.3f", x)
	}
	// h := Compose[int, float32, string](f, g) // works, but type-inference helps
	h := fp.Compose(g, f)
	h7 := h(7)
	if h7 != "7.500" {
		t.Logf("composition h(7) = %q", h(7))
		t.Error("expected h(7) to return string 7.500")
	}
}

func TestConst(t *testing.T) {
	seven := fp.Const(7)
	if seven() != 7 {
		t.Logf("const = %v", seven())
		t.Error("expected const to be integer 7")
	}
}

func TestUnit(t *testing.T) {
	nothing := fp.Unit(7)
	if nothing != 0 {
		t.Logf("Unit(7) = %v", nothing)
		t.Error("expected Unit(7) to be nothing = 0")
	}
}

func TestOptionSimple(t *testing.T) {
	x := fp.Maybe(7)
	if x.Just() != 7 {
		t.Errorf("expected Maybe(7) to be 7, is %v", x.Just())
	}
	y := fp.Option[int]{}
	if !y.IsNone() {
		t.Error("expected Option{} to be none, isn't")
	}
}

func TestErrOption(t *testing.T) {
	succ := fp.Success(7)
	if succ.Just() != 7 {
		t.Errorf("expected Success(7) to be 7, is %v", succ.Just())
	}
	fail := fp.Failure[int](errors.New("failure"))
	if fail.Error() == nil {
		t.Error("expected Failure(err) to reflect error, doesn't")
	}
}

func TestMatchable1(t *testing.T) {
  // match and decompose a tuple
  p1 := fp.P(1, 2)
  ok := p1.Matches(fp.P(1, 2))
  if !ok {
	  t.Logf("pair p1 = %v", p1)
	  t.Errorf("expected p1 to match (1,2), doesn't")
  }
  //var x, y int
  x, y := p1.Decompose()
  if x != 1 || y != 2 {
	  t.Logf(">> decompose into x = %d, y = %d", x, y)
	  t.Errorf("expected p1 to decompose into (1,2), doesn't")
  }
}
