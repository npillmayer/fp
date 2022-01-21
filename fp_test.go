package fp_test

import (
	"fmt"
	"testing"

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
