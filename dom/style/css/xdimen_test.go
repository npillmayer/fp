package css_test

/*
func TestDimen(t *testing.T) {
	teardown := testconfig.QuickConfig(t)
	defer teardown()
	gtrace.CoreTracer.SetTraceLevel(tracing.LevelDebug)
	//
	p := style.Property("100pt")
	d := DimenOption(p)
	if d.Unwrap() != dimen.DU(100)*dimen.PT {
		t.Errorf("expected 100 PT (%d), have %d", 100*dimen.PT, d)
	}
	//
	p = style.Property("auto")
	d = DimenOption(p)
	x, err := d.Match(option.Of{
		option.None: "NONE",
		Auto:        "AUTO",
	})
	if err != nil || x != "AaUTO" {
		t.Errorf("expected AUTO, have %v with error %v", x, err)
	}
}
*/
