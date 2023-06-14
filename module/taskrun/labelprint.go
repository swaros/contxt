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

 package taskrun

import (
	"fmt"

	"github.com/swaros/contxt/module/systools"
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
		switch ctrl := chk.(type) {
		case CtxOutCtrl:
			if chk.(CtxOutCtrl).IgnoreCase { // if we have found this flag set to true, it means ignore the message
				return
			}
		case CtxTargetOut:
			labelStr := ""
			if ctrl.Alternative != "" {
				labelStr = systools.LabelPrint(ctrl.Alternative, 1)
			} else {
				labelStr = systools.LabelPrintWithArg(systools.PadStringToR(ctrl.Target, ctrl.PanelSize), ctrl.ForeCol, ctrl.BackCol, 1)
			}
			newMsh = append(newMsh, labelStr)
		case CtxOutLabel:
			colmsg := manout.Message(ctrl.FColor, ctrl.Message) + " "
			newMsh = append(newMsh, colmsg)
		default:
			newMsh = append(newMsh, chk)
		}

	}
	msg = newMsh
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
