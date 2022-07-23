package cmdhandle

import (
	"fmt"

	"github.com/swaros/manout"
)

func LabelPrint(msg ...interface{}) {
	fmt.Print(manout.MessageCln(
		manout.BackDarkGrey, " ",
		manout.BackWhite, manout.ForeLightCyan,
		"   con", manout.ForeDarkGrey, ".", manout.ForeLightGrey, "txt   ",
		manout.BackDarkGrey, " ",
		manout.CleanTag, " "))
	fmt.Println(manout.MessageCln(msg...))
}
