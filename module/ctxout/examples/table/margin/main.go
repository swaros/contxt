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
	"github.com/swaros/contxt/module/ctxout"
)

// the tables are mainly ment to create output that is aligned in columns on screen, without floating around.
// the goal is not to create nice terminal sheets. so this table is just a collection of rows without any visible borders.
// or in other words, this kind of table, is focused to print data on screen and keep control of the output.
// all of them depends on the terminal size and how the content fits into it, depending on the settings.
// so if you like to have some static visuals in between, then you have to add them by yourself, and tell
// the table to reserve space for them.

// thats what this example is all about. we create a table with 2 cells per row and add a border sign between them.
// and second we add a margin to the cells, so that the border sign has some space to be printed.
// in the next example we add also a border sign to the left and right of the table.

// in the last example we create a full table with border signs on all sides, and between the cells.
// including a header and a footer.
// so this is something like a real table. but keep in mind, that this is not the goal of this table.
// if you just like to have a nice table on screen, in the easiest way possible, then you should maybe use an different approach or library.

func main() {
	// here the same as in the simple example (../simple/main.go). so the commented lines are the difference
	text := " -just-to-fill-some-space- "
	for i := 0; i < 5; i++ {
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
			),
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
			),
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
			),
		),
	)
	ctxout.PrintLn(table) // print the table

	ctxout.PrintLn(" ") // just create some space between the tables
	ctxout.PrintLn("\t------\t example of how margin works \t------\t")
	ctxout.PrintLn(" ")

	// now we create a table with 2 cells again, but this time we add the vertical line char between the cells
	// on the left and right side

	table = ctxout.Table(
		ctxout.Row(
			rowSep, // the left border sign
			ctxout.TD(
				"hello"+text,
				ctxout.Size(50),
				ctxout.Margin(3), // left + middle + right = 3 ...so we need to reserve space for 3 border signs
			),
			rowSep, // the middle border sign
			ctxout.TD(
				"world"+text,
				ctxout.Size(50),
			),
			rowSep, // the right border sign
		),
		ctxout.Row(
			rowSep,
			ctxout.TD(
				"hola"+text,
				ctxout.Size(50),
				ctxout.Margin(3),
			),
			rowSep,
			ctxout.TD(
				"mundo"+text,
				ctxout.Size(50),
			),
			rowSep,
		),
		ctxout.Row(
			rowSep,
			ctxout.TD(
				"hallo"+text,
				ctxout.Size(50),
				ctxout.Margin(3),
			),
			rowSep,
			ctxout.TD(
				"welt"+text,
				ctxout.Size(50),
			),
			rowSep,
		),
	)
	ctxout.PrintLn(table) // print the table

	ctxout.PrintLn(" ") // just create some space between the tables
	ctxout.PrintLn("\t-----\texample of an handcafted table output with a lot of borders ")
	ctxout.PrintLn(" ")

	// now we create a table with 2 cells again, but this time we add the vertical line char between the cells
	type tableData struct {
		left  string
		right string
	}
	data := []tableData{
		{"hello", "world"},
		{"hola", "mundo"},
		{"hallo", "welt"},
		{"bonjour", "monde"},
		{"ciao", "mondo"},
		{"hej", "världen"},
		{"hei", "verden"},
		{"salut", "monde"},
		{"ahoj", "světe"},
	}
	// all the chars we need to draw a nice bordered table
	hline := "─"
	tLeft := "├"
	tRight := "┤"
	tMiddle := "┼"
	tBottomLeft := "└"
	tBottomRight := "┘"
	tBottomMiddle := "┴"
	tTopLeft := "┌"
	tTopRight := "┐"
	tTopMiddle := "┬"

	tabContent := ""
	// create the table content and and a row separator after each row
	for _, d := range data {

		tabContent = tabContent +
			ctxout.Row( // this is the row with the border signs
				ctxout.ForeLightBlue+tLeft,
				ctxout.TD(
					hline,
					ctxout.Size(50),
					ctxout.Fill(hline),
					ctxout.Margin(3),
				),
				tMiddle,
				ctxout.TD(
					hline,
					ctxout.Fill(hline),
					ctxout.Size(50),
				),
				tRight,
			) + // here we add the row data
			ctxout.Row(
				rowSep+ctxout.ForeLightYellow,
				ctxout.TD(
					d.left+text,
					ctxout.Size(50),
					ctxout.Margin(3),
				),
				ctxout.ForeLightBlue+rowSep+ctxout.ForeLightYellow,
				ctxout.TD(
					d.right+text,
					ctxout.Size(50),
				),
				ctxout.ForeLightBlue+rowSep+ctxout.ForeLightYellow,
			)

	}

	// now create the table with header, data content as rows and footer
	table = ctxout.Table(
		ctxout.Row(
			ctxout.ForeLightBlue+tTopLeft,
			ctxout.TD(
				hline,
				ctxout.Size(50),
				ctxout.Fill(hline),
				ctxout.Margin(3),
			),
			tTopMiddle,
			ctxout.TD(
				hline,
				ctxout.Fill(hline),
				ctxout.Size(50),
			),
			tTopRight,
		),

		ctxout.Row(
			rowSep,
			ctxout.TD(
				"left word",
				ctxout.Size(50),
				ctxout.Margin(3),
			),
			rowSep,
			ctxout.TD(
				"right word",
				ctxout.Size(50),
			),
			rowSep+ctxout.ForeYellow,
		),

		tabContent,

		ctxout.Row(
			ctxout.ForeLightBlue+tBottomLeft,
			ctxout.TD(
				hline,
				ctxout.Size(50),
				ctxout.Fill(hline),
				ctxout.Margin(3),
			),
			tBottomMiddle,
			ctxout.TD(
				hline,
				ctxout.Fill(hline),
				ctxout.Size(50),
			),
			tBottomRight,
		),
	)
	ctxout.PrintLn(ctxout.NewMOWrap(), table, ctxout.CleanTag) // print the table. this time we need the colored output, so we use the MOWrap

}
