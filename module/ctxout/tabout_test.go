package ctxout_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
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
	expected := " ...or then 20 chars"
	if str != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, str)
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
	if ctxout.UniseqLen(output4) != 160 {
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
	if ctxout.UniseqLen(output4) != 160 {
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
			Out: "this is a test    ...s is another test",
		},
		{
			Row: "<row><tab size='23'>this is a test</tab><tab size='25' origin='1'>and this is another test</tab></row>",
			Out: "this is a test    and this is anoth...",
		},
		{
			Row: "<row><tab size='2'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>",
			Out: "th...s is another test",
		},
	}

	for i, test := range tests {
		output := to.Command(test.Row)
		if output != test.Out {
			t.Errorf("(%d) Expected\n'%s'\n  but got \n'%s'\nlength check %d ? %d ", i, test.Out, output, len(output), len(test.Out))
		}
	}

}

func TestFilterBehavior(t *testing.T) {

	// create a string that is longer than the terminal
	longText := "this is a very long text that should be cut off at the end of the line."
	for i := 0; i < 10; i++ {
		longText += fmt.Sprintf("[%v]", i) + longText
	}

	content := "<row><tab size='50'>" + longText + "</tab><tab size='50' origin='2'>and this is another test " + longText + "</tab></row>"
	// now we test again but now we fake a working terminal
	size := 800
	info := ctxout.PostFilterInfo{
		Width:      size,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}
	to := ctxout.NewTabOut()
	to.Update(info)

	output := to.Command(content)
	realLen := len(output)
	if realLen != size {
		t.Errorf("Expected length is not equal than %d but got %d", size, realLen)
		t.Log(output)
	}

}

func TestFilterBehavior2(t *testing.T) {
	// here we want to test the table output if the filter is disabled

	// create a string that is longer than the terminal
	longText := "this is a very long text that should be cut off at the end of the line."
	for i := 0; i < 2; i++ {
		longText += fmt.Sprintf("[%v]", i) + longText
	}

	lenPerStr := len(longText)
	content := "<row><tab size='50'>" + longText + "</tab><tab size='50' origin='2'>" + longText + "</tab></row>"
	// now we test again but now we fake a working terminal
	size := 800
	info := ctxout.PostFilterInfo{
		Width:      size,
		IsTerminal: false, // we make sure we have the behavior of a terminal
		Disabled:   true,
	}
	to := ctxout.NewTabOut()
	to.Update(info)

	output := to.Command(content)
	realLen := len(output)
	expectedLen := lenPerStr * 2
	if realLen != expectedLen {
		t.Errorf("Expected length is not equal than %d but got %d", expectedLen, realLen)
		t.Log(output)
	}

}

func TestPadding(t *testing.T) {
	abc := "abcdefghijklmnopqrstuvwxyz"
	cell := ctxout.NewTabCell(nil)
	cell.SetCutNotifier("").SetOverflowMode("wrap").SetText(abc).SetFillChar(".").SetOrigin(2)
	padStr := cell.CutString(10)
	expected := "abcdefghij"
	if padStr != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, padStr)
	}

	expected = "klmnopqrst\nuvwxyz"
	if cell.GetOverflowContent() != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, cell.GetOverflowContent())
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
				SetOverflowMode("wrap").
				SetText("abcdefghijklmnopqrstuvwxyz").
				SetFillChar(".").
				SetSize(10)))
	tableAsString := table.Render()
	expected := "abcdefghij\nklmnopqrst\nuvwxyz...."
	if tableAsString != expected {
		t.Errorf("Expected \n%s\n  [but got]\n%s", expected, tableAsString)
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
		t.Errorf("Expected \n%s\n<<<\n>>>> but got \n%s\n<<", expected, tableAsString)
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

func deb(output string) string {
	replaces := []string{
		"\n", "[\\n]",
		"\t", "[\\t-->]",
		"\r", "[\\r<--]",
	}
	for i := 0; i < len(replaces); i += 2 {
		output = strings.ReplaceAll(output, replaces[i], replaces[i+1])
	}
	return output

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
		// No 0
		{
			TestInput: "<row><tab overflow='wrap' fill='.' size='10'>0123456789</tab><tab fill='-' size='10' overflow='wrap' origin='0'>abcdefghijklmnopqrstuvwxyz</tab></row>",
			Out:       "01234567abcdefgh\n89......ijklmnop\n........qrstuvwx\n........yz------",
			Raw:       "0123456789abcdefghij\n..........klmnopqrst\n..........uvwxyz----",
			Info:      info,
		},
		// No 1
		{
			TestInput: "<row><tab overflow='wrap' size='5'>0123456789</tab><tab size='10' overflow='wrap' origin='0'>abcdefghijklmnopqrstuvwxyz</tab></row>",
			Out:       "0123abcdefgh\n4567ijklmnop\n89  qrstuvwx\n    yz      ",
			Raw:       "01234abcdefghij\n56789klmnopqrst\n     uvwxyz    ",
			Info:      info,
		},
		// No. 2
		{
			TestInput: "<row><tab size='23'>this is a test</tab><tab size='25' origin='2'>and this is another test</tab></row>",
			Out:       "this is a test    ...s is another test",
			Raw:       "this is a test          and this is another test",
			Info:      info,
		},
		// wordwrap test No. 3
		{
			TestInput: "<row><tab overflow='wordwrap' size='23'>this is a test about wordwarpping</tab><tab overflow='wordwrap' size='25' origin='2'>itisakwardtosplitlingtextifwedonothaveanywhhitespace</tab></row>",
			Out:       "this is a test    itisakwardtosplitlin\nabout wordwarppinggtextifwedonothavean\n                          ywhhitespace",
			Raw:       "this is a test about   itisakwardtosplitlingtext\nwordwarpping           ifwedonothaveanywhhitespa\n                                              ce",
			Info:      info,
		},

		// wordwrap test No. 4 TODO: this is not working yet
		/* not done yet. keep this for later and DAMN ME if this will ever get merged
		{
			TestInput: ctxout.Row(
				ctxout.TD("START-L first line of content END-L", ctxout.Prop("size", 50), ctxout.Prop("overflow", "wordwrap")),
				ctxout.TD("START-R this content is way morbigger then the other one END-R", ctxout.Prop("size", 50), ctxout.Prop("overflow", "wordwrap")),
			),
			Out:  "this is a test    \n about            \n wordwarpping     ",
			Raw:  "this is a test about\n wordwarpping       ",
			Info: info,
		},
		*/
	}

	forBreakPoint := 0

	for round, test := range tests {

		// this is just for debugging
		// and to have a place, where a breakpoint could be set
		if round == forBreakPoint {
			t.Log("breakpoint here")
		}
		to.Update(*resetInfo)
		output := to.Command(test.TestInput)
		if output != test.Raw {
			t.Errorf("round [%d] RAW Expected\nexpect:\"%s\"\n   got:\"%s\"\n", round, deb(test.Raw), deb(output))
		}
		to.Update(*test.Info)
		output = to.Command(test.TestInput)
		if output != test.Out {
			t.Errorf("round [%d] OUT Expected\nexpect:\"%s\"\n   got:\"%s\"\n\n", round, deb(test.Out), deb(output))
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

func TestUtf8Chars(t *testing.T) {
	type utfTesting struct {
		left  string
		right string
	}
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      100,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	to.Update(*info)

	utf8Chars := []utfTesting{
		{"testing left", "testing right"},
		{"\u2588 check", "\u2588 also check"},
		{"\u2588", "\u2588\u2588"},
		{"verify \u2588\u2588 is okay?", "block \u2588 check"},
		{"plain", "🖵"},
	}
	for _, utf8Char := range utf8Chars {
		testStr := ctxout.OTR +
			ctxout.TD(utf8Char.left, ctxout.Prop("size", 50), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 1)) +
			ctxout.TD(utf8Char.right, ctxout.Prop("size", 50), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 1)) +
			ctxout.CRT

		result := to.Command(testStr)
		// resulting length should be 100.
		// but be aware that we need to count the visible characters and they required space on screen.
		// so we need to count the utf8 characters and add the space they require (not the bytes)
		// and ignore escape sequences and other invisible characters
		// thats why we use ctxout.LenPrintable instead of len
		if ctxout.UniseqLen(result) != 100 {
			t.Errorf("Expected length %d but got %d", 100, len(result))
		}
	}
}

// helper to verify the content caculation by terminal size
func assertTestRowSize(t *testing.T, left, right string, leftSize int) {
	t.Helper()
	type utfTesting struct {
		left  string
		right string
	}
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      100,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	to.Update(*info)

	leftRightContent := []utfTesting{
		{left, right},
	}
	rightSize := 100 - leftSize
	for _, testContent := range leftRightContent {
		testStr := ctxout.Row(
			ctxout.TD(testContent.left, ctxout.Prop("size", leftSize), ctxout.Prop("fill", "_"), ctxout.Prop("origin", 1)),
			ctxout.TD(testContent.right, ctxout.Prop("size", rightSize), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 1)),
		)

		result := to.Command(testStr)
		if systools.StrLen(result) != 100 {
			t.Errorf("Expected length %d but got %d (%d)", 100, len(result), systools.StrLen(result))
			t.Log("composed string:")
			t.Log(result)
		}
	}
}

func TestDifferentRowSizes(t *testing.T) {
	assertTestRowSize(t, "testing left", "testing right", 50)

	assertTestRowSize(t,
		"testing left",
		`sjdhf kajsgdfg asfkgfkhafg
		sadfljh
		afjkgjghh    alfgjaffgf   eafgahffgfhfgfggfgg
		lsöa  ashfkahhf  lahflfig  awehfh  kawfhhf    aXX
		`,
		25,
	)
}

// here we are testing the size of the table
// depending on different terminal sizes
// and table render options

type testSetups struct {
	terminalWith int
	tableSource  string
	expectedSize int
}

func newTestSetup(terminalWith int, tableSource string, expectedSize int) *testSetups {
	return &testSetups{
		terminalWith: terminalWith,
		tableSource:  tableSource,
		expectedSize: expectedSize,
	}
}

func new100Setup(tableSource string, expectedSize int) *testSetups {
	return &testSetups{
		terminalWith: 100,
		tableSource:  tableSource,
		expectedSize: expectedSize,
	}
}

func (test *testSetups) assertSize(t *testing.T) *testSetups {
	t.Helper()
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      test.terminalWith,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	to.Update(*info)

	result := to.Command(test.tableSource)
	if ctxout.UniseqLen(result) != test.expectedSize {
		t.Errorf("Expected length %d but got %d", test.expectedSize, ctxout.UniseqLen(result))
	}
	return test
}

func (test *testSetups) assertContent(t *testing.T, expected string) *testSetups {
	t.Helper()
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      test.terminalWith,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	to.Update(*info)

	result := to.Command(test.tableSource)
	if result != expected {
		t.Errorf("content is not matching the expectation")
		fmt.Println("   !:", deb(expected))
		fmt.Println("   ?:", deb(result))
	}
	return test
}

// compare the created markup with the expected markup
func (test *testSetups) assertTableSource(t *testing.T, expected string) *testSetups {
	t.Helper()
	if test.tableSource != expected {
		t.Errorf("table source is not matching the expectation")

		fmt.Println("   !:", deb(expected))
		fmt.Println("   ?:", deb(test.tableSource))
	}
	return test
}

// compare if the output, split by an sperator, have allways the same length for all rows
func (test *testSetups) assertRowLength(t *testing.T, separator string) *testSetups {
	t.Helper()
	t.Helper()
	to := ctxout.NewTabOut()

	info := &ctxout.PostFilterInfo{
		Width:      test.terminalWith,
		IsTerminal: true, // we make sure we have the behavior of a terminal
	}

	to.Update(*info)

	result := to.Command(test.tableSource)
	rows := strings.Split(result, separator)
	firstRowSize := -1
	for _, row := range rows {
		if firstRowSize == -1 {
			firstRowSize = ctxout.UniseqLen(row)
		} else {
			if firstRowSize != ctxout.UniseqLen(row) && ctxout.UniseqLen(row) != 0 {
				t.Errorf("row length is not matching the expectation. fist row size is %d but one of the next row size is %d", firstRowSize, ctxout.UniseqLen(row))

			}
		}
	}
	return test
}

func TestSizes(t *testing.T) {

	table := ctxout.Row(
		ctxout.TD(
			"hello",
			ctxout.Prop("size", 50),
			ctxout.Prop("fill", "+"),
			ctxout.Prop("origin", ctxout.OriginLeft),
		),
	)
	new100Setup(table, 50).
		assertSize(t).
		assertTableSource(t, "<row><tab size='50' fill='+' origin='0'>hello</tab></row>")

	newTestSetup(50, table, 25).
		assertSize(t).
		assertContent(t, "hello++++++++++++++++++++")

	table = ctxout.Row(
		ctxout.TD(
			"this will be the text for the first row what should be a long text",
			ctxout.Size(75),
			ctxout.Fill("+"),
			ctxout.Right(),
		),
		ctxout.TD(
			"this is the second row and have also a long text to check if we still get the right size",
			ctxout.Size(25),
			ctxout.Fill("-"),
			ctxout.Left(),
		),
	)
	new100Setup(table, 100).
		assertSize(t).
		assertTableSource(t, "<row><tab size='75' fill='+' origin='2'>this will be the text for the first row what should be a long text</tab><tab size='25' fill='-' origin='0'>this is the second row and have also a long text to check if we still get the right size</tab></row>").
		assertContent(t, "+++++++++this will be the text for the first row what should be a long textthis is the second row...")

	newTestSetup(50, table, 50).
		assertSize(t).
		assertContent(t, "...irst row what should be a long textthis is t...")

	newTestSetup(25, table, 25).
		assertSize(t).
		assertContent(t, "...d be a long textthi...")

	// test combination of content size and extended size
	table = ctxout.Row(
		ctxout.TD(
			"this will be the text for the first row what should be a long text",
			ctxout.Size(75),
			ctxout.Fill("+"),
			ctxout.Content(),
		),
		ctxout.TD(
			"this is the second row and have also a long text to check if we still get the right size",
			ctxout.Size(25),
			ctxout.Fill("-"),
			ctxout.Extend(),
		),
	)
	new100Setup(table, 100).
		assertSize(t).
		assertTableSource(t, "<row><tab size='75' fill='+' draw='content'>this will be the text for the first row what should be a long text</tab><tab size='25' fill='-' draw='extend'>this is the second row and have also a long text to check if we still get the right size</tab></row>").
		assertContent(t, "this will be the text for the first row what should be a long textthis is the second row and have...")

}

func TestMarginExample(t *testing.T) {

	// here the same as i the simple example. so the commented lines are the difference
	text := " -just-to-fill-some-space- "
	for i := 0; i < 10; i++ {
		text += text
	}
	text = " : " + text

	//row separator is char alt + 186
	rowSep := "│"

	ctxout.AddPostFilter(ctxout.NewTabOut())

	// create a table that will have 2 cells per row
	// and sperate them with the vertical line char, between the cells
	table := ctxout.Table(
		ctxout.Row(
			ctxout.TD(
				"hello"+text,
				ctxout.Size(50),
				ctxout.Margin(1), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign

			),
			rowSep, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				"world"+text,
				ctxout.Size(50),
				ctxout.Margin(1),
			),
			rowSep, // the border sign
		),
		ctxout.Row(
			ctxout.TD(
				"hola"+text,
				ctxout.Size(50),
				ctxout.Margin(1), // again first row spend space for the border sign
			),
			rowSep, // here again we add the row sign
			ctxout.TD(
				"mundo"+text,
				ctxout.Size(50),
				ctxout.Margin(1),
			),
			rowSep,
		), // and so on...
		ctxout.Row(
			ctxout.TD(
				"hallo"+text,
				ctxout.Size(50),
				ctxout.Margin(1),
			),
			rowSep,
			ctxout.TD(
				"welt"+text,
				ctxout.Size(50),
				ctxout.Margin(1),
			),
			rowSep,
		),
	)
	new100Setup(table, 300).
		assertSize(t).
		assertRowLength(t, rowSep)

}

func TestOverLappingText(t *testing.T) {
	content := `2023-10-15 12:48:50,678 INFO n.b.b.l.LoggingLoader [main] Finished initializing logging from [config/logging]
	2023-10-15 12:48:50,679 DEBUG n.b.b.j.GuardedConfigurationLoader [main] Loading configuration class net.bigpoint.batman.core.config.BatmanConfig
	2023-10-15 12:48:50,680 INFO n.b.b.j.ConfigurationLoader [main] Reading config BatmanConfig from config...
	2023-10-15 12:48:50,703 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/00-libcom-default.conf
	2023-10-15 12:48:50,705 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-default.conf
	2023-10-15 12:48:50,712 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/05-batman-default.conf
	2023-10-15 12:48:50,713 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-tracking-default.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/101-local-config.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/20-sandbox-default.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/20-sandbox-default.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/30-runtime.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/42-sandbox-data.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/43-sandbox-cdn.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/99-docker-service.local.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/99-tracking-version.conf
	2023-10-15 12:48:50,780 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of ItemType
	2023-10-15 12:48:50,785 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock types from: data/0/stock/type.json
	2023-10-15 12:48:50,849 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of InventoryFilterLoaderStub
	2023-10-15 12:48:50,851 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock filter from: data/0/stock/filter
	2023-10-15 12:48:50,885 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 14 instances of ItemData
	2023-10-15 12:48:50,886 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 14 gamedata stock items from: data/0/stock/item
	2023-10-15 12:48:50,894 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 4 instances of ItemList
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 4 gamedata default items from: data/0/stock/item
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 0 instances of CooldownData
	2023-10-15 12:48:50,896 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 0 gamedata default cooldowns from: data/0/cooldown
	2023-10-15 12:48:51,663 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,665 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down`
	ctxout.AddPostFilter(ctxout.NewTabOut())
	tableCnt := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"LABEL",         // the text content, must be a string and the first argument
				ctxout.Size(50), // the size of the cell in percent of the table width

			),
			ctxout.TD( // the next cell
				content,
				ctxout.Size(50),
				//ctxout.Overflow(ctxout.OfIgnore),
			),
		),
	)
	drawCnt := ctxout.ToString(ctxout.NewMOWrap(), tableCnt)
	expectedLen := 100
	if ctxout.UniseqLen(drawCnt) != expectedLen {
		t.Errorf("Expected length %d but got %d", expectedLen, ctxout.UniseqLen(drawCnt))
	}
}

func TestOverLappingTextWithMargin(t *testing.T) {
	content := `2023-10-15 12:48:50,678 INFO n.b.b.l.LoggingLoader [main] Finished initializing logging from [config/logging]
	2023-10-15 12:48:50,679 DEBUG n.b.b.j.GuardedConfigurationLoader [main] Loading configuration class net.bigpoint.batman.core.config.BatmanConfig
	2023-10-15 12:48:50,680 INFO n.b.b.j.ConfigurationLoader [main] Reading config BatmanConfig from config...
	2023-10-15 12:48:50,703 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/00-libcom-default.conf
	2023-10-15 12:48:50,705 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-default.conf
	2023-10-15 12:48:50,712 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/05-batman-default.conf
	2023-10-15 12:48:50,713 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-tracking-default.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/101-local-config.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/20-sandbox-default.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/20-sandbox-default.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/30-runtime.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/42-sandbox-data.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/43-sandbox-cdn.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/99-docker-service.local.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/99-tracking-version.conf
	2023-10-15 12:48:50,780 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of ItemType
	2023-10-15 12:48:50,785 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock types from: data/0/stock/type.json
	2023-10-15 12:48:50,849 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of InventoryFilterLoaderStub
	2023-10-15 12:48:50,851 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock filter from: data/0/stock/filter
	2023-10-15 12:48:50,885 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 14 instances of ItemData
	2023-10-15 12:48:50,886 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 14 gamedata stock items from: data/0/stock/item
	2023-10-15 12:48:50,894 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 4 instances of ItemList
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 4 gamedata default items from: data/0/stock/item
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 0 instances of CooldownData
	2023-10-15 12:48:50,896 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 0 gamedata default cooldowns from: data/0/cooldown
	2023-10-15 12:48:51,663 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,665 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down`
	ctxout.AddPostFilter(ctxout.NewTabOut())
	tableCnt := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"LABEL",         // the text content, must be a string and the first argument
				ctxout.Size(50), // the size of the cell in percent of the table width

			),
			ctxout.TD( // the next cell
				content,
				ctxout.Size(50),
				ctxout.Overflow(ctxout.OfWordWrap),
				ctxout.Margin(4),
			),
		),
	)
	drawCnt := ctxout.ToString(ctxout.NewMOWrap(), tableCnt)
	expectedLen := 96 // 100 - 4 (margin)
	lines := strings.Split(drawCnt, "\n")
	for nr, line := range lines {
		if ctxout.UniseqLen(line) != expectedLen {
			t.Errorf("line no(%d) Expected length %d but got %d", nr, expectedLen, ctxout.UniseqLen(line))
		}
	}
}

func TestOverLappingTextWithMarginAndPrefix(t *testing.T) {
	content := `2023-10-15 12:48:50,678 INFO n.b.b.l.LoggingLoader [main] Finished initializing logging from [config/logging]
	2023-10-15 12:48:50,679 DEBUG n.b.b.j.GuardedConfigurationLoader [main] Loading configuration class net.bigpoint.batman.core.config.BatmanConfig
	2023-10-15 12:48:50,680 INFO n.b.b.j.ConfigurationLoader [main] Reading config BatmanConfig from config...
	2023-10-15 12:48:50,703 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/00-libcom-default.conf
	2023-10-15 12:48:50,705 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-default.conf
	2023-10-15 12:48:50,712 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/05-batman-default.conf
	2023-10-15 12:48:50,713 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/05-batman-tracking-default.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/10-protocol.conf
	2023-10-15 12:48:50,714 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/101-local-config.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/20-sandbox-default.conf
	2023-10-15 12:48:50,715 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/20-sandbox-default.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/30-runtime.conf
	2023-10-15 12:48:50,717 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/42-sandbox-data.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/43-sandbox-cdn.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/44-buildname.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file file:///home/developer/develop/wsd/run/gateway1/config/99-docker-service.local.conf
	2023-10-15 12:48:50,718 INFO n.b.b.j.ConfigurationLoader [main] Loading BatmanConfig configuration file jar:file:/home/developer/develop/wsd/run/gateway1/batman.jar!/config/99-tracking-version.conf
	2023-10-15 12:48:50,780 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of ItemType
	2023-10-15 12:48:50,785 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock types from: data/0/stock/type.json
	2023-10-15 12:48:50,849 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 2 instances of InventoryFilterLoaderStub
	2023-10-15 12:48:50,851 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 2 gamedata stock filter from: data/0/stock/filter
	2023-10-15 12:48:50,885 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 14 instances of ItemData
	2023-10-15 12:48:50,886 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 14 gamedata stock items from: data/0/stock/item
	2023-10-15 12:48:50,894 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 4 instances of ItemList
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 4 gamedata default items from: data/0/stock/item
	2023-10-15 12:48:50,895 INFO n.b.b.c.g.GameDataJSONLoader [main-gateway1] Loaded 0 instances of CooldownData
	2023-10-15 12:48:50,896 INFO n.b.b.c.g.AbstractConfigurationPineappleGuiceModule [main-gateway1] Loaded 0 gamedata default cooldowns from: data/0/cooldown
	2023-10-15 12:48:51,663 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,665 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager is shutting down
	2023-10-15 12:48:51,666 DEBUG o.a.h.i.n.c.PoolingNHttpClientConnectionManager [Finalizer] Connection manager shut down`
	ctxout.AddPostFilter(ctxout.NewTabOut())
	tableCnt := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"LABEL",         // the text content, must be a string and the first argument
				ctxout.Size(50), // the size of the cell in percent of the table width

			),
			ctxout.TD( // the next cell
				content,
				ctxout.Size(50),
				ctxout.Overflow(ctxout.OfWordWrap),
				ctxout.Margin(4),
				ctxout.Prop(ctxout.AttrSuffix, "-start-"),
			),
		),
	)
	drawCnt := ctxout.ToString(ctxout.NewMOWrap(), tableCnt)
	expectedLen := 96 // 100 - 4 (margin)
	lines := strings.Split(drawCnt, "\n")
	for nr, line := range lines {
		if ctxout.UniseqLen(line) != expectedLen {
			t.Errorf("line no(%d) Expected length %d but got %d", nr, expectedLen, ctxout.UniseqLen(line))
		}
	}
}

func TestPrefixAndPostFix(t *testing.T) {
	content := `lsdkyl lasjkgflagf  89789ncm.ks klalslfga alskjfjjfl öalsfkfahf löaösjjj. ------- slhgsg666 akshfk`
	ctxout.AddPostFilter(ctxout.NewTabOut())
	tableCnt := ctxout.Table( // new table
		ctxout.Row( // new row
			ctxout.TD( // new cell
				"LABEL",
				ctxout.Size(40),
				ctxout.Prop(ctxout.AttrPrefix, "LS|"),
				ctxout.Prop(ctxout.AttrSuffix, "LE|"),
			),
			ctxout.TD( // the next cell
				content,
				ctxout.Size(30),
				ctxout.Overflow(ctxout.OfWordWrap),
				ctxout.Prop(ctxout.AttrPrefix, "-start-"),
				ctxout.Prop(ctxout.AttrSuffix, "-end-"),
			),
			ctxout.TD( // the next cell
				content,
				ctxout.Size(30),
				ctxout.Overflow(ctxout.OfWrap),
				ctxout.Prop(ctxout.AttrPrefix, "-start-"),
				ctxout.Prop(ctxout.AttrSuffix, "-end-"),
			),
		),
	)
	drawCnt := ctxout.ToString(ctxout.NewMOWrap(), tableCnt)
	expectedLen := 100 // 100 - 4 (margin)
	lines := strings.Split(drawCnt, "\n")
	for nr, line := range lines {
		if ctxout.UniseqLen(line) != expectedLen {
			t.Errorf("line no(%d) Expected length %d but got %d", nr, expectedLen, ctxout.UniseqLen(line))
		}
	}
}
