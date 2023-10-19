package main

import (
	"fmt"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/runner"
)

// go run ./module/runner/example/randcolor/main.go
func main() {
	randColor := runner.NewRandColorStore()

	for i := 0; i < randColor.GetMaxVariants(); i++ {
		targetName := "variant" + fmt.Sprintf("_no_%v", i)
		colorPicked := randColor.GetOrSetIndexColor(targetName)
		ctxout.PrintLn(
			ctxout.NewMOWrap(),
			colorPicked.ColorMarkup(),
			"  is this an nice combination? can you read this? ", targetName, "  ", ctxout.CleanTag, "\t", i, "\t", colorPicked)
	}
}
