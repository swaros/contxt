package cmdhandle

import (
	"fmt"

	"github.com/swaros/contxt/context/systools"
	"github.com/swaros/manout"
)

var PreHook func(msg ...interface{}) bool = nil

type CtxOutCtrl struct {
	IgnoreCase bool
}

type CtxOutLabel struct {
	Message interface{}
	FColor  string
}

type CtxTargetOut struct {
	ForeCol     string
	BackCol     string
	SplitLabel  string
	Target      string
	Alternative string
	PanelSize   int
}

func CtxOut(msg ...interface{}) {
	if PreHook != nil { // if the prehook is defined AND it returns true, we just stop doing anything
		if abort := PreHook(msg...); abort {
			return
		}
	}
	var newMsh []interface{}
	for _, chk := range msg {
		switch chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return
			}
		case CtxTargetOut:
			ctrl := chk.(CtxTargetOut)
			labelStr := systools.LabelPrintWithArg(systools.PadStringToR(ctrl.Target, ctrl.PanelSize), ctrl.ForeCol, ctrl.BackCol, 1)
			newMsh = append(newMsh, labelStr)
		case CtxOutLabel:
			outc := chk.(CtxOutLabel)
			colmsg := manout.Message(outc.FColor, outc.Message)
			newMsh = append(newMsh, colmsg)
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

func ValF(val interface{}) CtxOutLabel {
	return CtxOutLabel{Message: val, FColor: manout.ForeLightBlue}
}

func InfoF(val interface{}) CtxOutLabel {
	return CtxOutLabel{Message: val, FColor: manout.ForeCyan}
}

func InfoRed(val interface{}) CtxOutLabel {
	return CtxOutLabel{Message: val, FColor: manout.ForeLightRed}
}

func InfoMinor(val interface{}) CtxOutLabel {
	return CtxOutLabel{Message: val, FColor: manout.ForeDarkGrey}
}

func defaultLabel(val interface{}) string {
	return fmt.Sprintf(" %10s", val)
}

func LabelFY(val interface{}) CtxOutLabel {
	ctxl := CtxOutLabel{Message: val, FColor: manout.ForeYellow}
	ctxl.Message = manout.Message("[", manout.ForeYellow, defaultLabel(val), manout.ForeLightYellow, "] ")
	return ctxl
}

func LabelOkF(val interface{}) CtxOutLabel {
	ctxl := CtxOutLabel{Message: val, FColor: manout.ForeLightGreen}
	ctxl.Message = manout.Message("[", manout.ForeYellow, defaultLabel(val), manout.ForeLightYellow, "] ")
	return ctxl
}

func LabelErrF(val interface{}) CtxOutLabel {
	ctxl := CtxOutLabel{Message: val, FColor: manout.ForeRed}
	ctxl.Message = manout.Message("[", manout.ForeYellow, defaultLabel(val), manout.ForeLightYellow, "] ")
	return ctxl
}
