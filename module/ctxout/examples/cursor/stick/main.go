package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"atomicgo.dev/cursor"
	"github.com/swaros/contxt/module/ctxout"
)

func main() {
	outpOne := ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeYellow, createRandomWordsWithRandomNl(0, 5), ctxout.CleanTag, "\n")
	//area1 := cursor.NewArea()
	//area2 := cursor.NewArea()
	fmt.Println(outpOne)
	//MAXlINES := -1
	for i := 0; i < 100; i++ {

		//	area1.Update(ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeBlue, createRandomWordsWithRandomNl(20), ctxout.CleanTag, "\n"))
		//	area1.Top()
		outp := ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeBlue, createRandomWordsWithRandomNl(10, 20), ctxout.CleanTag, "\n")

		lines := strings.Split(outp, "\n")
		lineCount := len(lines)
		time.Sleep(10 * time.Millisecond)
		cursor.Down(lineCount)
		//fmt.Println("--->", outp, "<--")
		fmt.Println(".", outp, ".")
		//cursor.Up(lineCount)
		time.Sleep(10 * time.Millisecond)
		cursor.DownAndClear(4)
		cursor.Up(3)
		time.Sleep(10 * time.Millisecond)
		cursor.StartOfLine()
		fmt.Println("#--->", "-one")
		time.Sleep(10 * time.Millisecond)
		fmt.Println("#--->", "-two")
		time.Sleep(10 * time.Millisecond)
		fmt.Println("#--->", "-three")
		time.Sleep(10 * time.Millisecond)
		cursor.Up(4)

		time.Sleep(10 * time.Millisecond)
	}

}

func createRandomWordsWithRandomNl(min, max int) string {
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
	var buffer bytes.Buffer
	for i := min; i < max; i++ {
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
