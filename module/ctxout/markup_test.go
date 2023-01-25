package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

// print out the diff between the parsed and the expected
func diffParsed(t *testing.T, parsed []ctxout.Parsed, expected []string) {
	for i, p := range parsed {
		if p.Text != expected[i] {
			t.Error(i, " fail [", expected[i], "] got [", p.Text, "]")
		} else {
			t.Log(i, " ok ", expected[i], " got ", p.Text)
		}
	}

	if len(parsed) > len(expected) {
		for i := len(expected); i < len(parsed); i++ {
			t.Error(i, " not expected got ", parsed[i].Text)
		}
	}

	if len(parsed) < len(expected) {
		for i := len(parsed); i < len(expected); i++ {
			t.Error(i, " expected [", expected[i], "] got nothing")
		}
	}
}

func TestMarkup(t *testing.T) {
	mu := ctxout.NewMarkup()
	res := mu.Parse(`Hello <style color="red">World</style>`)

	if len(res) != 4 {
		t.Error("expected 4 results. got ", len(res))
	}

	expected := []string{"Hello ", "<style color=\"red\">", "World", "</style>"}
	for i, r := range res {
		if r.Text != expected[i] {
			t.Error("expected ", expected[i], " got ", r.Text)
		}
	}
}

func TestMarkup2(t *testing.T) {
	mu := ctxout.NewMarkup()
	res := mu.Parse(`Marlon Brando <stay clean><some else>gotcha right<> nana <style color="red">World</style> chacka`)

	expected := []string{"Marlon Brando ", "<stay clean>", "<some else>", "gotcha right<> nana ", "<style color=\"red\">", "World", "</style>", " chacka"}
	if len(res) != len(expected) {
		t.Error("expected ", len(expected), " got ", len(res))
		diffParsed(t, res, expected)
	} else {
		for i, r := range res {
			if r.Text != expected[i] {
				t.Error("expected ", expected[i], " got ", r.Text)
			}
		}
	}
}
