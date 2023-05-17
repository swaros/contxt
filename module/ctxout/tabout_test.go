package ctxout_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestPadString(t *testing.T) {
	str := ctxout.PadStrLeft("this is a test", 20, " ")
	if str != "this is a test      " {
		t.Errorf("Expected 'this is a test      ' but got '%s'", str)
	}

	str = ctxout.PadStrLeft("we will now check that the text is cutted before we reach mor then 20 chars", 20, "-")
	if str != "we will now chec ..." {
		t.Errorf("Expected 'we will now chec ...' but got '%s'", str)
	}
}

func TestPadStringToRight(t *testing.T) {
	str := ctxout.PadStrRight("this is a test", 20, " ")
	if str != "      this is a test" {
		t.Errorf("Expected '      this is a test' but got '%s'", str)
	}

	str = ctxout.PadStrRight("we will now check that the text is cutted before we reach mor then 20 chars", 20, "-")
	if str != "we will now chec ..." {
		t.Errorf("Expected 'we will now check th' but got '%s'", str)
	}

	str = ctxout.PadStrRight("", 20, ".")
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

func TestRowOnlyOut(t *testing.T) {

	to := ctxout.NewTabOut()

	content := "<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>"
	output := to.Command(content)
	if !to.IsRow(content) {
		t.Errorf("Expected content is a row but got false")
	}
	expect := "this is a test          and this is another test"
	if output != expect {
		t.Errorf("Expected '%s' but got '%s'", expect, output)
	}

	// now we test again but now we fake a working terminal
	info := ctxout.PostFilterInfo{
		Width:      80,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}
	to.Update(info)

	type rowTesting struct {
		Row string
		Out string
	}

	tests := []rowTesting{
		{
			Row: "<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>",
			Out: "this is a test    and this is anoth...",
		},
		{
			Row: "<row><tab size='23'>this is a test</tab><tab size='25' origin='1'>and this is another test</tab></row>",
			Out: "this is a test    ...s is another test",
		},
		{
			Row: "<row><tab size='2'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>",
			Out: "..and this is anoth...",
		},
	}

	for _, test := range tests {
		output := to.Command(test.Row)
		if output != test.Out {
			t.Errorf("Expected\n'%s'\n  but got \n'%s'\nlength check %d ? %d ", test.Out, output, len(output), len(test.Out))
		}
	}

}

func TestPadding(t *testing.T) {
	abc := "abcdefghijklmnopqrstuvwxyz"
	cell := ctxout.NewTabCell(nil)
	cell.SetCutNotifier("").SetOverflowMode("any").SetText(abc).SetFillChar(".").SetOrigin(2)
	padStr := cell.CutString(10)
	if padStr != "abcdefghij" {
		t.Errorf("Expected 'abcdefghij' but got '%s'", padStr)
	}

	if cell.GetOverflowContent() != "klmnopqrstuvwxyz" {
		t.Errorf("Expected 'klmnopqrstuvwxyz' but got '%s'", cell.GetOverflowContent())
	}

	moved := cell.MoveToWrap()
	if !moved {
		t.Errorf("Expected moved to be true but got false")
	}
	padStr = cell.CutString(10)
	if padStr != "klmnopqrst" {
		t.Errorf("Expected 'klmnopqrst' but got '%s'", padStr)
	}
	if cell.GetOverflowContent() != "uvwxyz" {
		t.Errorf("Expected 'uvwxyz' but got '%s'", cell.GetOverflowContent())
	}

	moved = cell.MoveToWrap()
	if !moved {
		t.Errorf("Expected moved to be true but got false")
	}
	padStr = cell.CutString(10)
	if padStr != "....uvwxyz" {
		t.Errorf("Expected '....uvwxyz' but got '%s'", padStr)
	}
	if cell.GetOverflowContent() != "" {
		t.Errorf("Expected '' but got '%s'", cell.GetOverflowContent())
	}

	moved = cell.MoveToWrap()
	if moved {
		t.Errorf("Expected moved to be false but got true")
	}
	padStr = cell.CutString(10)
	if padStr != ".........." {
		t.Errorf("Expected '..........' but got '%s'", padStr)
	}

}

func TestRowDrawing(t *testing.T) {
	row := ctxout.NewTabRow(nil)

	row.AddCell(ctxout.NewTabCell(nil).SetCutNotifier("...").SetOverflowMode("ignore").SetText("abcdefghijklmnopqrstuvwxyz").SetFillChar(".").SetSize(10))
	row.AddCell(ctxout.NewTabCell(nil).SetCutNotifier("...").SetOverflowMode("ignore").SetText("0123456789,.-;:!§$%&/()*+#").SetFillChar(".").SetSize(10))
	rowAsStr, _, _ := row.Render()
	if row.Err == nil {
		t.Errorf("Expected an error but got nil")
	}
	fmt.Println(rowAsStr)

	table := ctxout.NewTableHandle(ctxout.NewTabOut())
	table.AddRow(row)

	tableAsString := table.Render()
	if tableAsString == "" {
		t.Error("Expected a table but got nothing")
	} else {
		if tableAsString != "abcdefg...0123456..." {
			t.Errorf("Expected 'abcdefg...0123456...' but got '%s'", tableAsString)
		}
	}
}

func TestRowWrap1(t *testing.T) {
	table := ctxout.NewTableHandle(ctxout.NewTabOut())
	table.AddRow(
		ctxout.NewTabRow(nil).
			AddCell(ctxout.NewTabCell(nil).SetCutNotifier("...").
				SetOverflowMode("any").
				SetText("abcdefghijklmnopqrstuvwxyz").
				SetFillChar(".").
				SetSize(10)))
	tableAsString := table.Render()
	expected := "abcdefghij\nklmnopqrst\nuvwxyz...."
	if tableAsString != expected {
		t.Errorf("Expected \n%s<<<\n>>>> but got \n%s<<", expected, tableAsString)
	}
}

func TestRowWrap2(t *testing.T) {
	table := ctxout.NewTableHandle(ctxout.NewTabOut())
	table.AddRow(
		ctxout.NewTabRow(nil).
			AddCell(ctxout.NewTabCell(nil).SetCutNotifier("...").
				SetOverflowMode("any").
				SetText("abcdefghijklmnopqrstuvwxyz").
				SetFillChar(".").
				SetOrigin(2).
				SetSize(10)))
	tableAsString := table.Render()
	expected := "abcdefghij\nklmnopqrst\n....uvwxyz"
	if tableAsString != expected {
		t.Errorf("Expected \n%s<<<\n>>>> but got \n%s<<", expected, tableAsString)
	}

}

func TestRowWrap3(t *testing.T) {
	table := ctxout.NewTableHandle(ctxout.NewTabOut())

	table.AddRow(
		ctxout.NewTabRow(table).
			AddCell(
				ctxout.NewTabCell(nil).
					SetCutNotifier("...").
					SetOverflowMode("any").
					SetText("abcdefghijklmnopqrstuvwxyz").
					SetFillChar(".").
					SetOrigin(2).
					SetSize(10),
			).AddCell(
			ctxout.NewTabCell(nil).
				SetCutNotifier("...").
				SetOverflowMode("any").
				SetText("0123456789,.-;:_!§$%&/()=?").
				SetFillChar("-").
				SetOrigin(0).
				SetSize(10),
		),
	)
	tableAsString := table.Render()
	expected := "abcdefghij0123456789\nklmnopqrst,.-;:_!§$%\n....uvwxyz&/()=?----"
	if tableAsString != expected {
		t.Errorf("Expected \n%s<<<\n>>>> but got \n%s<<", expected, tableAsString)
	}

}

func TestMultiple(t *testing.T) {

	to := ctxout.NewTabOut()
	type rowTesting struct {
		Info        *ctxout.PostFilterInfo // post filter info
		TestInput   string                 // input
		Out         string                 // expected output with post filter
		Raw         string                 // expected output without post filter
		ExpectedLen int
	}
	resetInfo := &ctxout.PostFilterInfo{
		Width:      100,
		IsTerminal: false, // we make sure we have the behavior of a terminal
	}
	info := &ctxout.PostFilterInfo{
		Width:      80,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	tests := []rowTesting{

		{
			TestInput: "<row><tab overflow='any' fill='.' size='10'>0123456789</tab><tab fill='-' size='10' overflow='any' origin='0'>abcdefghijklmnopqrstuvwxyz</tab></row>",
			Out:       "01234567abcdefgh\n89......ijklmnop\n........qrstuvwx\n........yz------",
			Raw:       "0123456789abcdefghij\n..........klmnopqrst\n..........uvwxyz----",
			Info:      info,
		},

		{
			TestInput: "<row><tab overflow='any' size='5'>0123456789</tab><tab size='10' overflow='any' origin='0'>abcdefghijklmnopqrstuvwxyz</tab></row>",
			Out:       "0123abcdefgh\n4567ijklmnop\n89  qrstuvwx\n    yz      ",
			Raw:       "01234abcdefghij\n56789klmnopqrst\n     uvwxyz    ",
			Info:      info,
		},
		{
			TestInput: "<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>",
			Out:       "this is a test    and this is anoth...",
			Raw:       "this is a test          and this is another test",
			Info:      info,
		},
		{
			TestInput: "<row><tab overflow='wordwrap' size='23'>this is a test about wordwarpping</tab><tab overflow='wordwrap' size='25' origin='2'>itisakwardtosplitlingtextifwedonothaveanywhhitespace</tab></row>",
			Out:       "this is a test    itisakwardtosplitlin\n about            gtextifwedonothavean\n wordwarpping             ywhhitespace",
			Raw:       "this is a test about   itisakwardtosplitlingtext\n wordwarpping          ifwedonothaveanywhhitespa\n                                              ce",
			Info:      info,
		},
	}

	for round, test := range tests {
		to.Update(*resetInfo)
		output := to.Command(test.TestInput)
		if output != test.Raw {
			t.Errorf("round [%d] RAW Expected\n\"%s\"\n>>>>> but got \n\"%s\"\n______%s", round, test.Raw, output, strings.Join(strings.Split(output, "\n"), "|"))
		}
		to.Update(*test.Info)
		output = to.Command(test.TestInput)
		if output != test.Out {
			t.Errorf("round [%d] OUT Expected\n\"%s\"\n>>>>>  but got \n\"%s\"\n_______%s", round, test.Out, output, strings.Join(strings.Split(output, "\n"), "|"))
		}
	}

}

func TestSizeCalculation(t *testing.T) {
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      80,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	// stop spamming the console with debug messages
	// if we get a couple of errors already
	maxErrors := 10

	to.Update(*info)

	for width := 10; width < 200; width++ {
		info.Width = width
		to.Update(*info)
		runCnt := 0
		for i := 6; i < 96; i++ {
			runCnt++
			right := i
			left := 100 - i
			// just to make sure left + right is always 100
			if left+right != 100 {
				t.Errorf("left + right should be 100 but is %d", left+right)
			}

			if width == -1 && left == 25 && right == 75 {
				fmt.Println("breakpoint here")
			}

			testStr := ctxout.OTR +
				ctxout.TD("hello", ctxout.Prop("size", left), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 1)) +
				ctxout.TD("world", ctxout.Prop("size", right), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 1)) +
				ctxout.CRT

			result := to.Command(testStr)

			if len(result) != width {
				t.Errorf("round[%d] origin[1,1] Expected length %d but got %d [left %d right %d]", runCnt, width, len(result), left, right)
				maxErrors--
			}

			testStr = ctxout.OTR +
				ctxout.TD("hello", ctxout.Prop("size", left), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 2)) +
				ctxout.TD("world", ctxout.Prop("size", right), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 1)) +
				ctxout.CRT

			result = to.Command(testStr)
			if len(result) != width {
				t.Errorf("round[%d] Origin[2,1] Expected length %d but got %d [left %d right %d]", runCnt, width, len(result), left, right)
				maxErrors--
			}

			testStr = ctxout.OTR +
				ctxout.TD("hello", ctxout.Prop("size", left), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 1)) +
				ctxout.TD("world", ctxout.Prop("size", right), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 2)) +
				ctxout.CRT

			result = to.Command(testStr)
			if len(result) != width {
				t.Errorf("round[%d] Origin[1,2] Expected length %d but got %d [left %d right %d]", runCnt, width, len(result), left, right)
				maxErrors--
			}

			testStr = ctxout.OTR +
				ctxout.TD("hello", ctxout.Prop("size", left), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 2)) +
				ctxout.TD("world", ctxout.Prop("size", right), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 2)) +
				ctxout.CRT

			result = to.Command(testStr)
			if len(result) != width {
				t.Errorf("round[%d] Origin[2,2] Expected length %d but got %d [left %d right %d]", runCnt, width, len(result), left, right)
				maxErrors--
			}
			if maxErrors <= 0 {
				t.Errorf("Too many errors, aborting")
				return
			}
		}

	}

}
