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

package runner

import (
	"fmt"
	"strings"
	"sync"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

type TableOutput struct {
	c *CmdExecutorImpl
}

func NewTableOutput() *TableOutput {
	return &TableOutput{}
}

func (t *TableOutput) GetName() string {
	return "table"
}

func (t *TableOutput) GetOutHandler(c *CmdExecutorImpl) func(msg ...interface{}) {
	t.c = c
	return func(msg ...interface{}) {
		var m sync.Mutex
		m.Lock()
		// go through all messages
		for _, m := range msg {
			// hanlde the message type
			switch tm := m.(type) {

			case tasks.MsgPid:
				targetColor, ok := randColors.GetOrSetRandomColor(tm.Target)
				if !ok {
					targetColor = RandColor{foreColor: "white", backColor: "black"}
				}
				c.Println(
					ctxout.Row(

						ctxout.TD(
							"PROCESS PID "+ctxout.BaseSignScreen+" ",
							ctxout.Prop(ctxout.AttrSize, 10),
							ctxout.Right(),
							ctxout.Prop(ctxout.AttrPrefix, processPreDef),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),

						ctxout.TD(
							tm.Pid,
							ctxout.Right(),
							ctxout.Prop(ctxout.AttrSize, 5),
							ctxout.Prop(ctxout.AttrPrefix, pidPreDef),
							ctxout.Prop(ctxout.AttrOverflow, "ignore"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),
						ctxout.TD(
							" .... ",
							ctxout.Prop(ctxout.AttrSize, 70),
							ctxout.Prop(ctxout.AttrPrefix, commentPreDef),
							ctxout.Prop(ctxout.AttrOverflow, "ignore"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),
						ctxout.TD(
							" "+tm.Target,
							ctxout.Prop(ctxout.AttrSize, 14),
							ctxout.Prop(ctxout.AttrPrefix, targetColor.ColorMarkup()),
							ctxout.Prop(ctxout.AttrOverflow, "ignore"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),
					),
				)

			case tasks.MsgCommand:
				t.drawRow(
					"executed command",
					ctxout.ForeYellow+ctxout.BoldTag,
					systools.AnyToStrNoTabs(tm),
					ctxout.ForeDarkGrey,
					ctxout.BaseSignScreen+" ",
					ctxout.ForeYellow,
				)

			// this is a special case where we need to
			// check against the context of the message
			case tasks.MsgTarget:
				// we get the command for the target
				// because we handle a target message,
				// we look for the target color
				targetColor, ok := randColors.GetOrSetRandomColor(tm.Target)
				if !ok {
					targetColor = RandColor{foreColor: "white", backColor: "black"}
				}
				switch tm.Context {

				case "command":
					t.drawRow(
						tm.Target,
						targetColor.ColorMarkup(),
						"cmd: "+tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignScreen+" ",
						ctxout.ForeYellow,
					)

				case "needs_required":
					t.drawRow(
						tm.Target,
						targetColor.ColorMarkup(),
						"require: "+strings.Join(strings.Split(tm.Info, ","), " "),
						ctxout.ForeDarkGrey,
						ctxout.BaseSignDebug,
						ctxout.ForeBlue,
					)

				case "needs_execute":
					t.drawRow(
						tm.Target,
						targetColor.ColorMarkup(),
						"execute: "+tm.Info,
						ctxout.ForeDarkGrey+ctxout.BoldTag+ctxout.Dim,
						ctxout.BaseSignScreen+" ",
						ctxout.ForeMagenta,
					)

				case "needs_done":
					t.drawRow(
						tm.Target,
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignSuccess+ctxout.BaseSignSuccess+ctxout.BaseSignSuccess+" ", // three green success signs for all subneeds done
						ctxout.ForeGreen,
					)

				case "wait_next_done":
					t.drawRow(
						ctxout.BaseSignSuccess+" "+tm.Target,
						ctxout.ForeGreen,
						"NEXT DONE ..."+tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignSuccess+" ",
						ctxout.ForeBlue,
					)
				case "requirement-check-failed":
					t.drawRow(
						ctxout.BaseSignWarning+" "+tm.Target,
						ctxout.ForeYellow,
						"SKIP REASON: ..."+ctxout.ForeLightCyan+tm.Info+ctxout.CleanTag,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "wait_for_targets":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"waiting for targets ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "wait_targets_done":
					t.drawRow(
						ctxout.BaseSignSuccess+" "+tm.Target,
						ctxout.ForeGreen,
						"TARGET DONE ..."+tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignSuccess+" ",
						ctxout.ForeBlue,
					)
				case "running_in_workingdir":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"running in workingdir ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "not_found":
					t.drawRow(
						ctxout.BaseSignError+" "+tm.Target,
						ctxout.ForeLightRed,
						"target not found ..."+tm.Info,
						ctxout.ForeRed,
						ctxout.BaseSignError+" ",
						ctxout.ForeRed,
					)
				case "ankocommand":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"running anko command ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "run-trigger-target-output":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"running trigger target output ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "execute-trigger-target":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"running trigger target ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				case "needs_ignored_runs_already":
					t.drawRow(
						ctxout.BaseSignInfo+" "+tm.Target,
						ctxout.ForeLightBlue,
						"needs ignored, runs already ..."+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ",
						ctxout.ForeBlue,
					)
				default:
					c.Println(
						ctxout.ForeCyan,
						" [",
						ctxout.ForeLightYellow,
						tm.Target,
						" @ ",
						tm.Context,
						ctxout.ForeCyan,
						"]",
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeCyan,
						tm.Context,
						ctxout.CleanTag,
					)
				}
				// end of switch tm.Context
			case tasks.MsgReason, tasks.MsgType:
				msg := fmt.Sprintf("%v", tm)
				if msg == "target-async-group-created" {
					t.drawRow(
						"system",
						ctxout.ForeLightBlue,
						"running async group ...",
						ctxout.ForeBlue,
						ctxout.BaseSignInfo+" ", // three green success signs for all subneeds done
						ctxout.ForeYellow,
					)
				} else {
					t.drawRow(
						"info",
						ctxout.ForeLightCyan,
						fmt.Sprintf("%v", tm),
						ctxout.ForeDarkGrey,
						ctxout.BaseSignInfo+" ", // three green success signs for all subneeds done
						ctxout.ForeBlue,
					)
				}
			case *tasks.MsgInfo:
				c.Println(
					ctxout.ForeLightMagenta,
					"[info]",
					ctxout.ForeMagenta,
					tm,
					ctxout.CleanTag,
				)
			// something depending the process is changed.
			// could be one of these:
			// - started
			// - stopped
			// - aborted
			// the comment contains more details.
			// on started it contains the command that is executed
			// on stopped it contains the exit codes (first the real code from the process, second the code from the command)
			// on aborted it contains the reason why the process is aborted. this is a controlled abort. nothing from system side.
			//            this depends on a defined stopreason, so contxt itself is aborting the process.
			case tasks.MsgProcess:
				targetColor, ok := randColors.GetOrSetRandomColor(tm.Target)
				if !ok {
					targetColor = RandColor{foreColor: "white", backColor: "black"}
				}
				c.Println(
					ctxout.Row(

						ctxout.TD(
							"PROCESS "+ctxout.BaseSignScreen+" ",
							ctxout.Prop(ctxout.AttrSize, 10),
							ctxout.Right(),
							ctxout.Prop(ctxout.AttrPrefix, processPreDef),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),

						ctxout.TD(
							" "+tm.StatusChange+" ",
							ctxout.Prop(ctxout.AttrSize, 5),
							ctxout.Right(),
							ctxout.Prop(ctxout.AttrPrefix, stateColorPreDef),
							ctxout.Prop(ctxout.AttrOverflow, "ignore"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),
						ctxout.TD(
							" "+tm.Comment,
							ctxout.Prop(ctxout.AttrSize, 70),
							ctxout.Prop(ctxout.AttrPrefix, commentPreDef),
							ctxout.Prop(ctxout.AttrOverflow, "wrap"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),

						ctxout.TD(
							" "+tm.Target,
							ctxout.Prop(ctxout.AttrSize, 14),
							ctxout.Prop(ctxout.AttrPrefix, targetColor.ColorMarkup()),
							ctxout.Prop(ctxout.AttrOverflow, "ignore"),
							ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
						),
					),
				)
			case tasks.MsgError:
				t.drawRow(
					tm.Target,
					ctxout.ForeLightYellow+ctxout.BoldTag+ctxout.BackRed,
					tm.Err.Error(),
					ctxout.ForeLightRed,
					" "+ctxout.BaseSignError+" ",
					//"error",
					ctxout.ForeYellow+ctxout.BoldTag+ctxout.BackRed,
				)
			case tasks.MsgErrDebug:
				t.drawRow(
					tm.Target,
					ctxout.ForeLightYellow+ctxout.BoldTag+ctxout.BackRed,
					c.formatDebugError(tm),
					ctxout.ForeLightRed,
					" "+ctxout.BaseSignError+" ",
					//"error",
					ctxout.ForeYellow+ctxout.BoldTag+ctxout.BackRed,
				)
			case tasks.MsgInfo:
				c.Println(
					ctxout.ForeYellow,
					"INFO",
					ctxout.ForeLightYellow,
					tm,
					ctxout.CleanTag,
				)
			case tasks.MsgExecOutput:
				// getting forground, background and the sign color for the arrow char
				fg, bg, sc := randColors.GetColorAsCtxMarkup(tm.Target)
				t.drawRowWithLabels(
					" ",
					sc+"<sign runident> ",
					tm.Target,
					fg+bg, //ctxout.ForeWhite+ctxout.BackBlue,
					systools.AnyToStrNoTabs(tm.Output),
					ctxout.ResetCode,
					ctxout.BaseSignScreen+" ",
					ctxout.ForeLightBlue,
				)

			case tasks.MsgArgs:
			// we just get ignored here

			case tasks.MsgNumber:
				// we just get ignored here

			default:
				// uncomment for dispaying the type of the message that is not handled yet

				c.Println(
					fmt.Sprintf("%T", tm),
					ctxout.ForeLightYellow,
					fmt.Sprintf("%v", msg),
					ctxout.ForeWhite,
					ctxout.BackBlack,
					"not implemented yet",
					ctxout.CleanTag,
				)

			}
		}

		m.Unlock()
	}
}

// short for drawRowWithLabels("", "", label, labelColor, content, contentColor, info, infoColor)
func (t *TableOutput) drawRow(label, labelColor, content, contentColor, info, infoColor string) {
	t.drawRowWithLabels("", "", label, labelColor, content, contentColor, info, infoColor)
}

// draws a row with labels and colors
// leftLabel and rightLabel are optional
// labelColor is the color markup for the label
// contentColor is the color markup for the content
// infoColor is the color markup for the info
// label, content and info are the strings to print
func (t *TableOutput) drawRowWithLabels(leftLabel, rightLabel, label, labelColor, content, contentColor, info, infoColor string) {
	if leftLabel == "" {
		leftLabel = "<sign runident> "
	}
	if rightLabel == "" {
		rightLabel = "<sign stopident> "
	}
	leftLabel = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeYellow+leftLabel+labelColor)
	rightLabel = ctxout.ToString(ctxout.NewMOWrap(), ctxout.ForeYellow+rightLabel+contentColor)
	t.c.Println(
		ctxout.Row(

			ctxout.TD(
				label,
				ctxout.Prop(ctxout.AttrSize, 10),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, leftLabel),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				//ctxout.Margin(4), // 4 spaces (run + space * 2 )
			),

			ctxout.TD(
				content,
				ctxout.Prop(ctxout.AttrSize, 85),
				ctxout.Prop(ctxout.AttrPrefix, rightLabel),
				ctxout.Prop(ctxout.AttrOverflow, "wordwrap"),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
			ctxout.TD(
				info,
				ctxout.Prop(ctxout.AttrSize, 5),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, infoColor),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				ctxout.Margin(1), // 1 space for being sure
			),
		),
	)
}
