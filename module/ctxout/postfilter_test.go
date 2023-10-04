package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestPluginAdd(t *testing.T) {
	ctxout.AddPostFilter(ctxout.NewTabOut())

	table := ctxout.Table(
		ctxout.Row(
			ctxout.TD(
				"hello",
				ctxout.Size(50),
			),
			ctxout.TD(
				"world",
				ctxout.Size(50),
			),
		),
	)
	out := ctxout.ToString(table)
	expected := "hello                                             world                                             "
	if out != expected {
		t.Errorf("expected [%s], got [%s]", expected, out)
	}

	filter := ctxout.GetPostFilterbyRef(ctxout.NewTabOut())
	if filter == nil {
		t.Errorf("expected a filter")
	}
	ctxout.UpdateFilterByRef(filter, ctxout.PostFilterInfo{Disabled: true})
	out = ctxout.ToString(table)
	expected = "helloworld"
	if out != expected {
		t.Errorf("expected [%s], got [%s]", expected, out)
	}
}
