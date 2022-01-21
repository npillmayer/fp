package maybe_test

import (
	"testing"

	. "github.com/npillmayer/fp/maybe"
)

func TestMaybeSimple(t *testing.T) {
	x := Just(7) // infers type
	y := Nothing[int]()
	//t.Logf("x = %d", x.Just()) // might panic

	var v int
	switch m := x.Match(); m {
	case m.Just(&v):
		t.Logf("Just(%d)", v)
	case m.Nothing():
		t.Logf("Nothing")
	}
	if v != 7 {
		t.Errorf("expected v to be 7, is %#v", v)
	}

	var w int
	switch m := y.Match(); m {
	case m.Just(&w):
		t.Logf("Just(%d)", w)
	case m.Nothing():
		t.Logf("Nothing")
	}
	if w != 0 {
		t.Errorf("expected w to be 0, is %#v", w)
	}
}

func TestMaybeWithDefault(t *testing.T) {
	x := Just(7)
	xx := x.WithDefault(100)
	if xx != 7 {
		t.Logf("y = %d", xx)
		t.Error("expected Just(7) to have value 7, isn't")
	}

	y := Nothing[int]()
	yy := y.WithDefault(100)
	if yy != 100 {
		t.Logf("y = %d", yy)
		t.Error("expected Nothing to default to 100, isn't")
	}
}

func TestMaybeMap(t *testing.T) {
	x := Just(7)
	xx := x.Map(func(n int) int {
		return n * 2
	})
	var v int
	switch m := xx.Match(); m {
	case m.Just(&v):
	case m.Nothing():
	}
	if v != 14 {
		t.Logf("x * 2 = %d", v)
		t.Error("expected Just(7).Map(…) to return 14, didn't")
	}

	x = Just(10)
	xx = Map(func(n int) int {
		return n * 2
	}, x)
	switch m := xx.Match(); m {
	case m.Just(&v):
	case m.Nothing():
	}
	if v != 20 {
		t.Logf("x * 2 = %d", v)
		t.Error("expected Map(…, Just 10) to return 20, didn't")
	}

	y := Nothing[int]()
	yy := y.Map(func(n int) int {
		return n * 2
	})
	var w int
	switch m := yy.Match(); m {
	case m.Just(&w):
	case m.Nothing():
		w = 99
	}
	if w != 99 {
		t.Logf("nothing * 2 = %d", w)
		t.Error("expected Nothing.Map(…) to return 99, didn't")
	}
}

func TestMaybeAndThen(t *testing.T) {
	gt0 := func(n int) Maybe[bool] {
		if n > 0 {
			return Just(true)
		}
		return Nothing[bool]()
	}

	gt := AndThen(gt0, Just(7))
	var isGreater bool
	switch m := gt.Match(); m {
	case m.Just(&isGreater):
		t.Logf("ok: 7 > 0")
	case m.Nothing():
		t.Error("expected Just(7) |> andThen(gt0) to be true, isn't")
	}
}
