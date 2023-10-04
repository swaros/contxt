// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

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
	cursorAndTableFilter()
}

// baseFunc is just a simple example to show the basic usage of the ctxout module
// and the ctxout.PrintLn function
func baseFunc() {
	printNextExampleHeader("color codes")
	// simple print with line break. nothing special
	ctxout.PrintLn("hello", " ", "world")
	time.Sleep(time.Millisecond * 1400)

	// print with color. but we will see the color codes
	// because we have nothing that will handle the color codes
	ctxout.PrintLn(ctxout.ForeRed, "(okay now we should see markups....) hello", " ", ctxout.ForeGreen, "world", ctxout.ResetCode)
	time.Sleep(time.Millisecond * 1400)

	// now the colors are shown, because we injected a PrinterInterface
	// that will handle the color codes
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeRed, "(but now colored output) hello", " ", ctxout.ForeGreen, "world", ctxout.ResetCode)
	time.Sleep(time.Millisecond * 1400)

	// the same as above, but with the PrinterInterface injected in between
	// the arguments. this is just to point out that the PrinterInterface
	// should be the first argument. at least before the first color code
	// or other special code is used
	ctxout.PrintLn(ctxout.ForeRed, "(and here mixed loutput because of maybe wrong placed printerInterface) hello", " ", ctxout.NewMOWrap(), ctxout.ForeGreen, "world", ctxout.ResetCode)
	time.Sleep(time.Millisecond * 1400)

	// the filter decides depending on the environment if the color codes
	// should be shown or not.
	// because this is done automatically, we have to force disabling the
	// the color codes here.
	// this is done by setting the NoColored flag to true

	// first we keep the default behavior
	originBehavior := ctxout.GetBehavior()
	// and set the NoColored flag to true
	ctxout.SetBehavior(ctxout.CtxOutBehavior{NoColored: true})
	// and here we will get a clean string, even if the color codes are used
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.ForeRed, "hello", " ", ctxout.ForeGreen, "world", ctxout.ResetCode)
	time.Sleep(time.Millisecond * 1400)

	// for the next example we will use the the original behavior again
	ctxout.SetBehavior(originBehavior)
}

func printNextExampleHeader(name string) {
	ctxout.PrintLn("   ")
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.BoldTag, ctxout.ForeWhite, name, ctxout.ResetCode)
	ctxout.PrintLn(ctxout.NewMOWrap(), ctxout.BoldTag, ctxout.ForeWhite, "---------------------------", ctxout.ResetCode)
}

func tableFilter() {
	printNextExampleHeader("output in an table")
	// add the tabout filter
	ctxout.AddPostFilter(ctxout.NewTabOut())
	// some different ways to print a table
	// table means a row with tabs that fits the size of the the current terminal
	// first it could be done just by a simple string line that contains the markups
	ctxout.PrintLn("<table><row><tab size='50'>hello</tab><tab size='50'>world</tab></row></table>")

	// or by seperate arguments
	ctxout.PrintLn("<table>", "<row>", "<tab size='50'>", "hello", "</tab>", "<tab size='50'>", "world", "</tab>", "</row>", "</table>")

	// or by using the constants that are defined in the tabconst.go file
	ctxout.PrintLn(
		ctxout.OpenTable,
		ctxout.OpenRow,
		"<tab size='50'>",
		"hello",
		ctxout.CloseTab,
		"<tab size='50'>",
		"world",
		ctxout.CloseTab,
		ctxout.CloseRow,
		ctxout.CloseTable,
	)
	// also there are some helper functions that can be used to set the size of an tab
	ctxout.PrintLn(ctxout.OpenTableRow, ctxout.Tab(50), "hello", ctxout.CloseTab, ctxout.Tab(50), "world", ctxout.CloseTabRowTable)

	// we can use the ctxout.TabF function to format the tabs
	// here we use the fill and origin attributes to define how the tabs should be filled
	// and where the origin of the tab should be
	ctxout.PrintLn(
		ctxout.OTR,
		ctxout.TabF("size=50", "fill=:", "origin=1"),
		"hello",
		ctxout.CTB,
		ctxout.TabF("size=50", "fill=.", "origin=2"),
		"world",
		ctxout.CTRT,
	)

	// here we fake an ongoing process by changing the size of the tabs
	// NOTE: we delayed the output by 100ms to see the process
	// also we use the shortcut functions to open and close the table cells
	for i := 0; i < 50; i = i + 5 {
		ctxout.PrintLn(
			ctxout.OTR,
			ctxout.TD("hello", ctxout.Prop("size", 50-i), ctxout.Prop("fill", "+"), ctxout.Prop("origin", 1)),
			ctxout.TD("world", ctxout.Prop("size", 50+i), ctxout.Prop("fill", "-"), ctxout.Prop("origin", 2)),
			ctxout.CRT,
		)
	}

	printNextExampleHeader("terminal output simulated")
	// here we delay the output by 100ms to see the process
	// also we use the shortcuts to open and close the table
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords(20)
		rndWord2 := createRandomWords(20)
		ctxout.PrintLn(
			ctxout.Table(
				ctxout.Row(
					ctxout.TD(rndWord1, "size=50", "origin=2"),
					ctxout.TD(rndWord2, "size=50", "origin=1"),
				),
			),
		)
		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}

	printNextExampleHeader("color is also not a problem")
	// now we add the color filter
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords(20)
		rndWord2 := createRandomWords(30)
		rndWord3 := createRandomWords(50)
		ctxout.PrintLn(
			ctxout.NewMOWrap(), // if we use colorcodes we need to inject a printer interface
			ctxout.Table(
				ctxout.Row(
					ctxout.TD(
						rndWord1,
						ctxout.Prop(ctxout.AttrSize, 10),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeLightBlue),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),

					ctxout.TD(
						rndWord2,
						ctxout.Prop(ctxout.AttrSize, 50),
						ctxout.Prop(ctxout.AttrOrigin, 1),
					),

					ctxout.TD(
						rndWord3,
						ctxout.Prop(ctxout.AttrSize, 40),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrPrefix, ctxout.BackCyan+ctxout.ForeLightYellow),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
				),
			),
		)
		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}

	printNextExampleHeader("overflow of content")
	// here we use the overflow attribute to define what should happen if the content is to long
	// the default is to cut the content
	// but we can also define that the content should be wrapped
	// or that the content should be truncated
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords(10)
		rndWord2 := createRandomWords(70)
		rndWord3 := createRandomWords(80)

		oddBackground := ctxout.BackCyan
		// we use the modulo operator to check if the row is odd or even
		if i%2 == 0 {
			oddBackground = ctxout.BackLightCyan
		}

		ctxout.PrintLn(
			ctxout.NewMOWrap(),
			ctxout.Table(
				ctxout.Row(
					ctxout.TD(
						rndWord1,
						ctxout.Prop(ctxout.AttrSize, 10),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrPrefix, oddBackground+ctxout.ForeWhite),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),

					ctxout.TD(
						rndWord2,
						ctxout.Prop(ctxout.AttrSize, 50),
						ctxout.Prop(ctxout.AttrOrigin, 1),
						ctxout.Prop(ctxout.AttrOverflow, ctxout.OverflowWordWrap),
						ctxout.Prop(ctxout.AttrPrefix, oddBackground+ctxout.ForeDarkGrey),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),

					ctxout.TD(
						rndWord3,
						ctxout.Prop(ctxout.AttrSize, 40),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrOverflow, ctxout.OverflowWordWrap),
						ctxout.Prop(ctxout.AttrPrefix, oddBackground+ctxout.ForeLightYellow),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
				),
			),
		)
		// wait a little bit. here a little bit longer because the content is longer
		time.Sleep(time.Millisecond * 400)
	}

	printNextExampleHeader("pure row usage")
	// the same as above, but without the table opened and closed
	// here we just use the row.
	// this is working because the table is not required to display the row
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords(90)
		rndWord2 := createRandomWords(90)
		ctxout.PrintLn(
			ctxout.Row(
				ctxout.TD(rndWord1, "size=30", "origin=2"),
				ctxout.TD(rndWord2, "size=70", "origin=1"),
			),
		)

		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}

	printNextExampleHeader("table right use")
	// here we use the table right
	// and as long as the table is not closed, the content is not printed
	ctxout.PrintLn("wait for it ...")
	ctxout.PrintLn(ctxout.OpenTable)
	for i := 0; i < 20; i++ {
		rndWord1 := createRandomWords(90)
		rndWord2 := createRandomWords(90)
		// NOTE: do not use printLn here, because it would move the cursor to the next line
		ctxout.Print(
			ctxout.Row(
				ctxout.TD("table usage", "size=30", "origin=1"),
				ctxout.TD(rndWord1, "size=30", "origin=2"),
				ctxout.TD(rndWord2, "size=40", "origin=1"),
			),
		)

		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}
	// close the table
	ctxout.PrintLn(ctxout.CloseTable)
	ctxout.PrintLn(" ---- TADAAAAA")
	// wait a second
	time.Sleep(time.Second)

}

func cursorAndTableFilter() {
	printNextExampleHeader("cursor filter + table filter + colored output")
	ctxout.AddPostFilter(ctxout.NewCursorFilter())

	for i := 0; i < 50; i++ {
		lines := 0
		for m := 1; m < 6; m++ {
			rndWord1 := createRandomWords(90)
			rndWord2 := createRandomWords(90)
			rndWord3 := createRandomWords(90)
			ctxout.PrintLn(
				ctxout.NewMOWrap(),
				ctxout.OTR,
				ctxout.TD(m, "size=5", "origin=2", "suffix="+ctxout.ResetCode, "prefix="+ctxout.ForeLightYellow),
				ctxout.TD(i, "size=5", "origin=2", "suffix="+ctxout.ResetCode, "prefix="+ctxout.ForeLightBlue+ctxout.BackBlue),
				ctxout.TD(rndWord1, "size=20", "suffix="+ctxout.ResetCode, "prefix="+ctxout.ForeGreen),
				ctxout.TD(rndWord2, "size=30", "suffix="+ctxout.ResetCode, "prefix="+ctxout.ForeDarkGrey),
				ctxout.TD(rndWord3, "size=40"),
				ctxout.CRT)

			lines++
		}
		lineUpCmd := fmt.Sprintf("cursor:up,%v;", lines+1)
		ctxout.PrintLn(lineUpCmd)

		// wait a little bit
		time.Sleep(time.Millisecond * 100)
	}
	ctxout.PrintLn("cursor:down,10;")
	ctxout.PrintLn("okay... done")
}

func createRandomWords(min int) string {
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

	// create a random string between 5 and 100 words
	offset := 100 - min
	var buffer bytes.Buffer
	for i := 0; i < rand.Intn(offset)+min; i++ {
		buffer.WriteString(words[rand.Intn(len(words))])
		buffer.WriteString(" ")
	}
	return buffer.String()
}
