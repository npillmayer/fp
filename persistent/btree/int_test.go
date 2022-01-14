package btree

import "testing"

// test internals

func TestInternalCeiling(t *testing.T) {
	c := []struct {
		n    int
		ceil int
	}{
		{0, 0},
		{2, 4},
		{3, 8},
		{4, 8},
		{6, 8},
		{7, 16},
	}
	for i, x := range c {
		xx := ceiling(x.n)
		if xx != x.ceil {
			t.Errorf("%d: expected ceiling(%d) to be %d, is %d", i, x.n, x.ceil, xx)
		}
	}
}
