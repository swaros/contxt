package cmdhandle

import (
	"fmt"

	"github.com/swaros/manout"
)

func LabelPrint(msg ...interface{}) {
	fmt.Print(manout.MessageCln(manout.BackWhite, manout.ForeCyan, "con", manout.ForeDarkGrey, ".", manout.ForeLightGrey, "txt", manout.CleanTag, " "))
	fmt.Println(manout.MessageCln(msg...))
}
