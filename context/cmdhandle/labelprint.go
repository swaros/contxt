package cmdhandle

import (
	"fmt"

	"github.com/swaros/manout"
)

type CtxOutCtrl struct {
	IgnoreCase bool
}

func CtxOut(msg ...interface{}) {
	var newMsh []interface{}
	for _, chk := range msg {
		switch chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return
			}
		default:
			newMsh = append(newMsh, chk)
		}

	}
	msg = newMsh
	fmt.Print(manout.MessageCln(
		manout.BackDarkGrey, " ",
		manout.BackWhite, manout.ForeLightCyan,
		"   con", manout.ForeDarkGrey, ".", manout.ForeLightGrey, "txt   ",
		manout.BackDarkGrey, " ",
		manout.CleanTag, " "))
	fmt.Println(manout.MessageCln(msg...))
}

func ValF(val interface{}) string {
	return manout.Message(manout.ForeLightBlue, val, " ")
}

func InfoF(val interface{}) string {
	return manout.Message(manout.ForeLightCyan, val, " ")
}

func InfoRed(val interface{}) string {
	return manout.Message(manout.ForeLightRed, val, " ")
}

func InfoMinor(val interface{}) string {
	return manout.Message(manout.ForeDarkGrey, val, " ")
}

func defaultLabel(val interface{}) string {
	return fmt.Sprintf(" %10s", val)
}

func LabelFY(val interface{}) string {
	return manout.Message(manout.ForeLightYellow, "[", manout.ForeYellow, defaultLabel(val), manout.ForeLightYellow, "] ")
}

func LabelOkF(val interface{}) string {
	return manout.Message(manout.ForeLightGreen, "[", manout.ForeGreen, defaultLabel(val), manout.ForeLightGreen, "] ")
}

func LabelErrF(val interface{}) string {
	return manout.Message(manout.ForeLightRed, "[", manout.ForeRed, defaultLabel(val), manout.ForeLightRed, "] ")
}
