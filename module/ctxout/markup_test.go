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

func TestCheckFlag(t *testing.T) {
	mu := ctxout.NewMarkup()
	parsed := mu.Parse(`<style color='red'></style><style  maincolor='red'></style><style colorback='blue'></style>`)
	if len(parsed) != 6 {
		t.Error("expected 6 results. got ", len(parsed))
	}

	if parsed[0].GetProperty("color", "blue") != "red" {
		t.Error("expected red got ", parsed[0].GetProperty("color", "blue"))
	}

	if parsed[2].GetProperty("color", "blue") != "blue" {
		t.Error("expected blue got ", parsed[2].GetProperty("color", "blue"))
	}

	if parsed[4].GetProperty("color", "orange") != "orange" {
		t.Error("expected orange got ", parsed[4].GetProperty("color", "orange"))
	}

	if parsed[4].GetProperty("color", float64(15)) != float64(15) {
		t.Error("expected 15 as float64 got ", parsed[4].GetProperty("color", float64(15)))
	}

	if parsed[4].GetProperty("color", 15) != 15 {
		t.Error("expected 15  got ", parsed[4].GetProperty("color", 15))
	}

	val, found := parsed[0].GetMarkupIntValue("color")
	if found {
		t.Error("expected not found got ", val)
	}

	valStr, xfound := parsed[0].GetMarkupStringValue("color")
	if !xfound {
		t.Error("expected found got not found")
	} else {
		if valStr != "red" {
			t.Error("expected red got ", valStr)
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

func TestMixedMarkups(t *testing.T) {
	testStr := "<tab size='5' origin='2'>0</tab><tab size='30' origin='2'>/home/testpath/someplace/check</tab><tab size='30'>no tasks</tab>"
	mu := ctxout.NewMarkup()
	res := mu.Parse(testStr)

	expected := []string{"<tab size='5' origin='2'>", "0", "</tab>", "<tab size='30' origin='2'>", "/home/testpath/someplace/check", "</tab>", "<tab size='30'>", "no tasks", "</tab>"}
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

func TestOwnMarkup(t *testing.T) {
	testStr := "{ONE}hello{-ONE}{TWO}world{-TWO}"
	mu := ctxout.NewMarkup()
	res := mu.SetStartToken('{').SetEndToken('}').SetCloseIdent('-').Parse(testStr)
	if len(res) != 6 {
		t.Error("expected 6 results. got ", len(res))
	}

	expected := []string{"{ONE}", "hello", "{-ONE}", "{TWO}", "world", "{-TWO}"}
	for i, r := range res {
		if r.Text != expected[i] {
			t.Error("expected ", expected[i], " got ", r.Text)
		}
	}
}

func TestBuildInnerSliceEach(t *testing.T) {
	testStr := "<map><item>one</item><item>two</item><item>three</item></map><item>last</item>"
	mu := ctxout.NewMarkup()

	parsed := mu.Parse(testStr)
	if len(parsed) != 14 {
		t.Error("expected 14 results. got ", len(parsed))
	}
	hits := 0
	var hitRes []ctxout.Parsed
	parsedInner := mu.BuildInnerSliceEach(parsed, "map", func(markup []ctxout.Parsed) bool {
		// we should only hit once with an non empty slice
		if len(markup) > 0 {
			hitRes = markup
			hits++
		}
		return true
	})

	if hits != 1 {
		t.Error("expected 1 hit got ", hits)
	}

	if len(hitRes) != 9 {
		t.Error("expected 9 results. got ", len(hitRes), hitRes)
	}
	// the return slice should be empty
	if len(parsedInner) != 0 {
		t.Error("expected 0 results. got ", len(parsedInner))
	}
	// build the inner slice

}
