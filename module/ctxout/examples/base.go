package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"time"

	"github.com/swaros/contxt/module/ctxout"
)

func main() {
	baseFunc()
	tableFilter()
}

// baseFunc is just a simple example to show the basic usage of the ctxout module
// and the ctxout.PrintLn function
func baseFunc() {
	printNextExampleHeader("baseFunc")
	// simple print with line break. nothing special
	ctxout.PrintLn("hello", " ", "world")

	// print with color. but we will see the color codes
	// because we have nothing that will handle the color codes
	ctxout.PrintLn(ctxout.ForeRed, "hello", " ", ctxout.ForeGreen, "world", ctxout.ResetCode)

	// now the colors are shown, because we injected a PrinterInterface
	// that will handle the color codes
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeRed, "hello", " ", ctxout.ForeGreen, "world", ctxout.ResetCode)

	// the same as above, but with the PrinterInterface injected in between
	// the arguments. this is just to point out that the PrinterInterface
	// should be the first argument. at least before the first color code
	// or other special code is used
	ctxout.PrintLn(ctxout.ForeRed, "hello", " ", ctxout.NewMOWrap(), ctxout.ForeGreen, "world", ctxout.ResetCode)
}

func printNextExampleHeader(name string) {
	ctxout.PrintLn("   ")
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.BoldTag, ctxout.ForeWhite, name, ctxout.ResetCode)
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.BoldTag, ctxout.ForeWhite, "---------------------------", ctxout.ResetCode)
}

func tableFilter() {
	printNextExampleHeader("tableFilter")
	// add the tabout filter
	ctxout.AddPostFilter(ctxout.NewTabOut())
	// some different ways to print a table
	// table means a row with tabs that fits the size of the the current terminal
	ctxout.PrintLn("<table><row><tab size='50'>hello</tab><tab size='50'>world</tab></row></table>")
	ctxout.PrintLn("<table>", "<row>", "<tab size='50'>", "hello", "</tab>", "<tab size='50'>", "world", "</tab>", "</row>", "</table>")
	ctxout.PrintLn(ctxout.OPEN_TABLE, ctxout.OPEN_ROW, "<tab size='50'>", "hello", ctxout.CLOSE_TAB, "<tab size='50'>", "world", ctxout.CLOSE_TAB, ctxout.CLOSE_ROW, ctxout.CLOSE_TABLE)
	ctxout.PrintLn(ctxout.OPEN_TABLE, ctxout.OPEN_ROW, "<tab size='50'>", "hello", ctxout.CLOSE_TAB, "<tab size='50'>", "world", ctxout.CLOSE_TAB_ROW_TABLE)
	ctxout.PrintLn(ctxout.OPEN_TABLE_ROW, ctxout.Tab(50), "hello", ctxout.CLOSE_TAB, ctxout.Tab(50), "world", ctxout.CLOSE_TAB_ROW_TABLE)

	// now we can use the ctxout.TabF function to format the tabs
	// here we use the fill and origin attributes to define how the tabs should be filled
	// and where the origin of the tab should be
	ctxout.PrintLn(ctxout.OTR, ctxout.TabF("size=50", "fill=:", "origin=1"), "hello", ctxout.CTB, ctxout.TabF("size=50", "fill=.", "origin=2"), "world", ctxout.CTRT)
	ctxout.PrintLn(ctxout.OTR, ctxout.TabF("size=50", "fill=:", "origin=2"), "hello", ctxout.CTB, ctxout.TabF("size=50", "fill=.", "origin=1"), "world", ctxout.CTRT)
	for i := 0; i < 50; i = i + 5 {
		sizeAtrr := fmt.Sprintf("size='%v'", 50-i)
		otherSizeAtrr := fmt.Sprintf("size='%v'", 50+i)
		ctxout.PrintLn(ctxout.OTR, ctxout.TabF(sizeAtrr, "fill=-", "origin=1"), "hello", ctxout.CTB, ctxout.TabF(otherSizeAtrr, "fill=|", "origin=2"), "world", ctxout.CTRT)
	}

	ctxout.AddPostFilter(ctxout.NewCursorFilter())
	// stick filter
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords()
		rndWord2 := createRandomWords()
		ctxout.PrintLn(
			ctxout.OTR,
			ctxout.TabF("size=30", "fill=-", "origin=1"),
			rndWord1,
			ctxout.CTB,
			ctxout.TabF("size=70", "fill=|", "origin=2"),
			rndWord2,
			ctxout.CTRT)

		//ctxout.PrintLn("cursor:bottom;round no:", i)
		//ctxout.PrintLn("cursor:up,6;")
		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}

	for i := 0; i < 50; i++ {
		lines := 0
		for m := 1; m < 6; m++ {
			rndWord1 := createRandomWords()
			rndWord2 := createRandomWords()
			rndWord3 := createRandomWords()
			//cursorCmd := fmt.Sprintf("cursor:down,%v;", m)

			ctxout.PrintLn(
				//cursorCmd,
				ctxout.OTR,
				ctxout.TabF("size=5", "fill=-", "origin=1"),
				m,
				ctxout.CTB,
				ctxout.TabF("size=25", "fill=-", "origin=2"),
				rndWord1,
				ctxout.CTB,
				ctxout.TabF("size=30", "fill=|", "origin=2"),
				rndWord2,
				ctxout.CTB,
				ctxout.TabF("size=40", "fill=|", "origin=2"),
				rndWord3,
				ctxout.CTRT)

			lines++
		}
		lineUpCmd := fmt.Sprintf("cursor:up,%v;", lines+1)
		ctxout.PrintLn(lineUpCmd)

		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}

}

func createRandomWords() string {
	// just some random words
	// they are not really random, but it is enough for this example

	words := []string{
		"Lorem", "ipsum", "dolor", "sit", "amet,", "consectetur", "adipiscing", "elit.",
		"Nulla", "facilisi.", "Sed", "eu", "diam", "nec", "nisl", "consequat", "viverra.",
		"Vivamus", "nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus", "nec",
		"diam", "nec", "nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec",
		"nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat",
		"viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus",
		"nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec",
		"nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat",
		"viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus",
		"nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec",
		"nisl", "consequat", "viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat",
		"viverra.", "Vivamus", "nec", "diam", "nec", "nisl", "consequat", "viverra.", "Vivamus",
	}

	// create a random string with 100 words
	var buffer bytes.Buffer
	for i := 0; i < 100; i++ {
		buffer.WriteString(words[rand.Intn(len(words))])
		buffer.WriteString(" ")
	}
	return buffer.String()
}
