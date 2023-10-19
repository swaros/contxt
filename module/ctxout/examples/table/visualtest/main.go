package main

import (
	"bytes"
	"fmt"
	"math/rand"

	"github.com/swaros/contxt/module/ctxout"
)

func main() {
	// add table filter
	ctxout.AddPostFilter(ctxout.NewTabOut())
	// print the table

	for i := 0; i < 5; i++ {
		rndWord1 := createRandomWordsWithRandomNl(90)
		rndWord2 := createRandomWordsWithRandomNl(70)
		rndWord3 := createRandomWords(80)

		rndWord1 = fmt.Sprintf("(%v) %s", len(rndWord1), rndWord1)
		rndWord2 = fmt.Sprintf("(%v) %s", len(rndWord2), rndWord2)
		rndWord3 = fmt.Sprintf("(%v) %s", len(rndWord3), rndWord3)

		blueOddColor := ctxout.BackCyan + ctxout.ForeWhite
		bwOddColor := ctxout.BackWhite + ctxout.ForeBlack
		// we use the modulo operator to check if the row is odd or even
		if i%2 == 0 {
			blueOddColor = ctxout.BackLightCyan + ctxout.ForeCyan
			bwOddColor = ctxout.BackLightGrey + ctxout.ForeBlue
		}

		ctxout.PrintLn(
			ctxout.NewMOWrap(),
			ctxout.Table(
				ctxout.Row(
					ctxout.TD(
						rndWord1,
						ctxout.Prop(ctxout.AttrSize, 10),
						ctxout.Prop(ctxout.AttrOrigin, ctxout.OriginLeft),
						ctxout.Prop(ctxout.AttrPrefix, bwOddColor),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),

					ctxout.TD(
						rndWord2,
						ctxout.Prop(ctxout.AttrSize, 30),
						ctxout.Prop(ctxout.AttrOrigin, ctxout.OriginLeft),
						ctxout.Prop(ctxout.AttrOverflow, ctxout.OfWordWrap),
						ctxout.Prop(ctxout.AttrPrefix, blueOddColor),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
					ctxout.TD(
						rndWord1,
						ctxout.Prop(ctxout.AttrSize, 30),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrOverflow, ctxout.OfWrap),
						ctxout.Prop(ctxout.AttrPrefix, bwOddColor),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
					ctxout.TD(
						rndWord3,
						ctxout.Prop(ctxout.AttrSize, 30),
						ctxout.Prop(ctxout.AttrOrigin, 2),
						ctxout.Prop(ctxout.AttrOverflow, ctxout.OfIgnore),
						ctxout.Prop(ctxout.AttrPrefix, blueOddColor),
						ctxout.Prop(ctxout.AttrSuffix, ctxout.ResetCode),
					),
				),
			),
		)
		// wait a little bit. here a little bit longer because the content is longer
	}
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
	return "[SOT]" + buffer.String() + "[EOT]"
}

func createRandomWordsWithRandomNl(min int) string {
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
		// add random new lines
		if rand.Intn(10) == 0 {
			buffer.WriteString("\n")
		} else {
			buffer.WriteString(" ")
		}
	}
	return "[SOT]" + buffer.String() + " the last [EOT]"
}
