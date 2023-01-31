package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestPadString(t *testing.T) {
	str := ctxout.PadString("this is a test", 20, " ")
	if str != "this is a test      " {
		t.Errorf("Expected 'this is a test      ' but got '%s'", str)
	}

	str = ctxout.PadString("we will now check that the text is cutted before we reach mor then 20 chars", 20, "-")
	if str != "we will now check th" {
		t.Errorf("Expected 'we will now check th' but got '%s'", str)
	}
}

func TestPadStringToRight(t *testing.T) {
	str := ctxout.PadStringToRight("this is a test", 20, " ")
	if str != "      this is a test" {
		t.Errorf("Expected '      this is a test' but got '%s'", str)
	}

	str = ctxout.PadStringToRight("we will now check that the text is cutted before we reach mor then 20 chars", 20, "-")
	if str != "we will now check th" {
		t.Errorf("Expected 'we will now check th' but got '%s'", str)
	}

	str = ctxout.PadStringToRight("", 20, ".")
	if str != "...................." {
		t.Errorf("Expected '....................' but got '%s'", str)
	}

}

func TestBasicTabout(t *testing.T) {
	to := ctxout.NewTabOut()
	output1 := to.Command("<table>")
	output2 := to.Command("<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>")
	output3 := to.Command("<row><tab size='23'>second line</tab><tab size='25' origin='2'>we do it different</tab></row>")
	output4 := to.Command("</table>")

	expect := ""
	if output1 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output1)
	}

	expect = ""
	if output2 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output2)
	}

	expect = ""
	if output3 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output3)
	}

	expect = `this is a test          and this is another test
second line                   we do it different`
	if output4 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output4)
	}
}

func TestSizedTabout(t *testing.T) {
	to := ctxout.NewTabOut()
	info := ctxout.PostFilterInfo{
		Width:      80,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}
	to.Update(info)
	output1 := to.Command("<table>")
	output2 := to.Command("<row><tab fill='.' size='50'>this is a test</tab><tab fill='+' size='50' origin='2'>and this is another test</tab></row>")
	output3 := to.Command("<row><tab fill='_' size='50'>second line</tab><tab fill='-' size='50' origin='2'>we do it different</tab></row>")
	output4 := to.Command("</table>")

	expect := ""
	if output1 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output1)
	}

	expect = ""
	if output2 != expect {
		t.Errorf("Expected '%s' but got\n'%s'\n", expect, output2)
	}

	expect = ""
	if output3 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output3)
	}

	// len of text should exactly match the size 80 chars multiplied by 2 rows
	if ctxout.LenPrintable(output4) != 160 {
		t.Errorf("Expected 160 chars but got %d", len(output4))
	}

	expect = `this is a test..........................++++++++++++++++and this is another test
second line_____________________________----------------------we do it different`
	if output4 != expect {
		t.Errorf("Expected\n%s\nbut got\n%s", expect, output4)
	}
}

func TestSizedTaboutWithDrawModes(t *testing.T) {
	to := ctxout.NewTabOut()
	info := ctxout.PostFilterInfo{
		Width:      80,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}
	to.Update(info)
	output1 := to.Command("<table>")
	output2 := to.Command("<row><tab fill='.' draw='content' size='50'>this is a test</tab><tab fill='-' size='50' draw='extend' origin='2'>and this is another test</tab></row>")
	output3 := to.Command("<row><tab fill='.' draw='content' size='50'>second line</tab><tab fill='-' size='50' draw='extend' origin='2'>we do it different</tab></row>")
	output4 := to.Command("</table>")

	expect := ""
	if output1 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output1)
	}

	expect = ""
	if output2 != expect {
		t.Errorf("Expected '%s' but got\n'%s'\n", expect, output2)
	}

	expect = ""
	if output3 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output3)
	}

	// len of text should exactly match the size 80 chars multiplied by 2 rows
	if ctxout.LenPrintable(output4) != 160 {
		t.Errorf("Expected 160 chars but got %d", len(output4))
	}

	expect = `this is a test------------------------------------------and this is another test
second line...------------------------------------------------we do it different`
	if output4 != expect {
		t.Errorf("Expected\n%s\nbut got\n%s", expect, output4)
	}
}

func TestSizedTaboutNotClosedExtra(t *testing.T) {
	to := ctxout.NewTabOut()

	output1 := to.Command("<table>")
	output2 := to.Command("<row><tab fill='.' size='50'>this is a test</tab><tab fill='+' size='50' origin='2'>and this is another test</tab></row>")
	output3 := to.Command("<row><tab fill='_' size='50'>second line</tab><tab fill='-' size='50' origin='2'>we do it different</tab></row></table>")

	expect := ""
	if output1 != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output1)
	}

	expect = ""
	if output2 != expect {
		t.Errorf("Expected\n%s\nbut got\n'%s'\n", expect, output2)
	}

	expect = `this is a test....................................++++++++++++++++++++++++++and this is another test
second line_______________________________________--------------------------------we do it different`
	if output3 != expect {
		t.Errorf("Expected\n%s\n but got\n%s\n__", expect, output3)
	}

}

// testing single row in a table. all in one in one line
func TestTableTabout(t *testing.T) {
	to := ctxout.NewTabOut()
	output := to.Command("<table><row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row></table>")

	expect := "this is a test          and this is another test"
	if output != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output)
	}
}

func TestGettingEscape(t *testing.T) {
	escapInStr := "Hello \033[33m" + "World" + "\033[0m"

	lastEscape := ctxout.GetLastEscapeSequence(escapInStr)

	if lastEscape != "\033[0m" {
		t.Errorf("Expected '%s' but got '%s'", "\033[0m", lastEscape)
	}
}