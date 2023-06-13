package main

import (
	"github.com/swaros/contxt/module/ctxout"
)

// the tables are mainly ment to create output that is aligned in columns
// without any magic tabledrawings. so this table is just a collection of rows without any visible borders.
// or in other words, this kind of table, is focused to print data on screen and keep control of the output.
// all of them depends on the terminal size and how the content fits into it, depending on the settings.
// so if you like to have some static visuals in between, then you have to add them by yourself, and tell
// the table to reserve space for them.

// thats what this example is all about. we create a table with 2 cells per row and add a border sign between them.
// and second we add a margin to the cells, so that the border sign has some space to be printed.

func main() {
	// here the same as i the simple example. so the commented lines are the difference
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
			rowSep, // here we add the row sign
			ctxout.TD(
				"hello"+text,
				ctxout.Size(50),
				ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
			),
			rowSep, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				"world"+text,
				ctxout.Size(50),
			),
			rowSep, // here again we add the row sign
		),
		ctxout.Row(
			rowSep, // here we add the row sign
			ctxout.TD(
				"hola"+text,
				ctxout.Size(50),
				ctxout.Margin(3), // again first row spend space for the border sign
			),
			rowSep, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				"mundo"+text,
				ctxout.Size(50),
			),
			rowSep, // here again we add the row sign
		), // and so on...
		ctxout.Row(
			rowSep, // here we add the row sign
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
			rowSep, // here again we add the row sign
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
	for _, d := range data {

		tabContent = tabContent +
			ctxout.Row(
				tLeft, // here we add the row sign
				ctxout.TD(
					hline,
					ctxout.Size(50),
					ctxout.Fill(hline),
					ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
				),
				tMiddle, // the border sign. we reserved space for it with the margin
				ctxout.TD(
					hline,
					ctxout.Fill(hline),
					ctxout.Size(50),
				),
				tRight, // here again we add the row sign
			) +
			ctxout.Row(
				rowSep, // here we add the row sign
				ctxout.TD(
					d.left+text,
					ctxout.Size(50),
					ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
				),
				rowSep, // the border sign. we reserved space for it with the margin
				ctxout.TD(
					d.right+text,
					ctxout.Size(50),
				),
				rowSep, // here again we add the row sign
			)

	}

	// how coud this looks like in an program, if we would cretae the table with a loop? and with a nice border?
	table = ctxout.Table(
		ctxout.Row(
			ctxout.ForeLightBlue+tTopLeft, // here we add the row sign
			ctxout.TD(
				hline,
				ctxout.Size(50),
				ctxout.Fill(hline),
				ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
			),
			tTopMiddle, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				hline,
				ctxout.Fill(hline),
				ctxout.Size(50),
			),
			tTopRight, // here again we add the row sign

		),

		ctxout.Row(
			rowSep, // here we add the row sign
			ctxout.TD(
				"left word",
				ctxout.Size(50),
				ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
			),
			rowSep, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				"right word",
				ctxout.Size(50),
			),
			rowSep+ctxout.ForeYellow, // here again we add the row sign
		),

		tabContent,

		ctxout.Row(
			tBottomLeft, // here we add the row sign
			ctxout.TD(
				hline,
				ctxout.Size(50),
				ctxout.Fill(hline),
				ctxout.Margin(3), // the margin of the cell in percent of the cell width. this is used to reserve space for the border sign
			),
			tBottomMiddle, // the border sign. we reserved space for it with the margin
			ctxout.TD(
				hline,
				ctxout.Fill(hline),
				ctxout.Size(50),
			),
			tBottomRight, // here again we add the row sign
		),
	)
	ctxout.PrintLn(ctxout.NewMOWrap(), table) // print the table. this time we need the colored output, so we use the MOWrap

}
