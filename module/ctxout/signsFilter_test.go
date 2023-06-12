package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestBasicFilter(t *testing.T) {
	// create a new filter
	sf := ctxout.NewSignFilter(nil)
	if sf == nil {
		t.Error("NewSignFilter should not return nil")
	}
	// update the filter with some info
	sf.Update(ctxout.PostFilterInfo{
		IsTerminal: true,
		Width:      80,
		Height:     24,
	})

	// no sign should not be handled
	if sf.CanHandleThis("hello") {
		t.Error("should not handle this")
	}

	// a sign should be handled
	if !sf.CanHandleThis("hello <sign info>") {
		t.Error("should handle this")
	}

}

func TestBasicFilterWorking(t *testing.T) {
	ctxout.ClearPostFilters()
	// create a new filter
	sf := ctxout.NewSignFilter(nil)
	if sf == nil {
		t.Error("NewSignFilter should not return nil")
	}

	source := "hello <sign info>"
	expected := "hello 🗩"
	ctxout.AddPostFilter(sf)

	// force the filter to be enabled
	sf.Enable()

	chk := ctxout.ToString(source)

	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}

	sf.Disable()
	expected = "hello [i]"
	chk = ctxout.ToString(source)
	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}

}

func TestBasicFilterWorkingWithMutliple(t *testing.T) {
	ctxout.ClearPostFilters()
	// create a new filter
	sf := ctxout.NewSignFilter(nil)
	if sf == nil {
		t.Error("NewSignFilter should not return nil")
	}

	source := "hello <sign info> this is a test <sign success>"
	expected := "hello 🗩 this is a test ✔"
	ctxout.AddPostFilter(sf)

	// force the filter to be enabled
	sf.Enable()

	chk := ctxout.ToString(source)

	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}

	sf.Disable()
	expected = "hello [i] this is a test [v]"
	chk = ctxout.ToString(source)
	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}

}

func TestGetSign(t *testing.T) {
	sf := ctxout.NewSignFilter(nil)
	if sf == nil {
		t.Error("NewSignFilter should not return nil")
	}

	sign := sf.GetSign("info")
	if sign.Fallback != "[i]" {
		t.Error("sign should be [i]")
	}

	if sign.Glyph != "🗩" {
		t.Error("sign should be 🗩")
	}
}

func TestSignAdded(t *testing.T) {
	ctxout.ClearPostFilters()
	// create a new filter
	sf := ctxout.NewSignFilter(nil)
	if sf == nil {
		t.Error("NewSignFilter should not return nil")
	}

	source := "hello <sign danger>"
	expected := "hello <!danger>"
	ctxout.AddPostFilter(sf)
	sf.Enable()
	chk := ctxout.ToString(source)

	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}

	sf.AddSign(ctxout.Sign{
		Name:     "danger",
		Glyph:    "🌩",
		Fallback: "[!!!]",
	})

	expected = "hello 🌩"
	chk = ctxout.ToString(source)

	if chk != expected {
		t.Errorf("expected '%s' got '%s'", expected, chk)
	}
}
