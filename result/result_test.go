package result_test

import (
	"errors"
	"testing"

	. "github.com/npillmayer/fp/result"
)

func TestResultSimple(t *testing.T) {
	x := Ok(7) // infers type
	y := Err[int](errors.New("not ok"))

	var v int
	var e error

	switch m := x.Match(); m {
	case m.Ok(&v):
		t.Logf("Ok(%d)", v)
	case m.Err(&e):
		t.Logf("Err")
	}
	if v != 7 {
		t.Errorf("expected v to be 7, is %#v", v)
	}

	switch m := y.Match(); m {
	case m.Ok(&v):
		t.Logf("Ok(%d)", v)
	case m.Err(&e):
		t.Logf("Err: %s", e.Error())
	}
	if e == nil {
		t.Errorf("expected error to be non-nil, but it is nil")
	}
}
