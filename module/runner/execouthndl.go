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
	"sync"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func (c *CmdExecutorImpl) drawRow(label, labelColor, content, contentColor, info, infoColor string) {
	c.Println(
		ctxout.Row(
			ctxout.ForeYellow, "<sign runident> ", ctxout.CleanTag,
			ctxout.TD(
				label,
				ctxout.Prop(ctxout.AttrSize, 10),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, labelColor),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				ctxout.Margin(4), // 4 spaces (run + space * 2 )
			),
			ctxout.ForeYellow, "<sign stopident> ", ctxout.CleanTag,
			ctxout.TD(
				content,
				ctxout.Prop(ctxout.AttrSize, 85),
				ctxout.Prop(ctxout.AttrPrefix, contentColor),
				//ctxout.Prop(ctxout.AttrOverflow, "wrap"),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
			ctxout.TD(
				info,
				ctxout.Prop(ctxout.AttrSize, 5),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, infoColor),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
		),
	)
}

func (c *CmdExecutorImpl) drawTwoRow(content, contentColor, info, infoColor string) {
	c.Println(
		ctxout.Row(
			ctxout.TD(
				ctxout.BaseSignScreen+" > ",
				ctxout.Prop(ctxout.AttrSize, "5"),
				ctxout.Prop(ctxout.AttrPrefix, ctxout.ForeBlue),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
			ctxout.TD(
				content,
				ctxout.Prop(ctxout.AttrSize, "90"),
				ctxout.Prop(ctxout.AttrPrefix, contentColor),
				ctxout.Prop(ctxout.AttrOverflow, "wordwrap"),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
			ctxout.TD(
				"| "+info,
				ctxout.Prop(ctxout.AttrSize, "4"),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, infoColor),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
			),
		),
	)
}

// handles all the incomming messages from the tasks
// depending on the message type it will print the message
func (c *CmdExecutorImpl) getOutHandler() func(msg ...interface{}) {
	return func(msg ...interface{}) {
		var m sync.Mutex
		m.Lock()
		for _, m := range msg {

			switch tm := m.(type) {

			case tasks.MsgCommand:
				c.drawRow(
					"executed command",
					ctxout.ForeYellow+ctxout.BoldTag,
					systools.AnyToStrNoTabs(tm),
					ctxout.ForeDarkGrey,
					ctxout.BaseSignScreen+" ",
					ctxout.ForeYellow,
				)

			case tasks.MsgTarget:
				switch tm.Context {
				case "command":
					c.drawRow(
						tm.Target,
						ctxout.ForeYellow+ctxout.BoldTag,
						tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignScreen+" ",
						ctxout.ForeYellow,
					)

				case "needs_required":
					c.drawRow(
						tm.Target,
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignDebug,
						ctxout.ForeBlue,
					)

				case "needs_execute":
					c.drawRow(
						tm.Target,
						ctxout.ForeYellow,
						"request to start ... "+tm.Info,
						ctxout.ForeBlue,
						ctxout.BaseSignScreen+" ",
						ctxout.ForeMagenta,
					)

				case "needs_done":
					c.drawRow(
						tm.Target,
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeDarkGrey,
						ctxout.BaseSignSuccess+ctxout.BaseSignSuccess+ctxout.BaseSignSuccess+" ", // three green success signs for all subneeds done
						ctxout.ForeGreen,
					)

				case "wait_next_done":
					c.drawRow(
						ctxout.BaseSignSuccess+" DONE",
						ctxout.ForeLightGreen,
						tm.Target,
						ctxout.ForeLightYellow,
						ctxout.BaseSignSuccess+" ",
						ctxout.ForeBlue,
					)

				default:
					c.Println(
						ctxout.ForeCyan,
						" [",
						ctxout.ForeLightYellow,
						tm.Target,
						ctxout.ForeCyan,
						"]",
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeCyan,
						tm.Context,
						ctxout.CleanTag,
					)
				}

			case tasks.MsgReason, tasks.MsgType:
				c.Println(
					ctxout.ForeLightMagenta,
					"mixed",
					ctxout.ForeMagenta,
					tm,
					ctxout.CleanTag,
				)

			case *tasks.MsgInfo:
				c.Println(
					ctxout.ForeLightMagenta,
					"[info]",
					ctxout.ForeMagenta,
					tm,
					ctxout.CleanTag,
				)
			case tasks.MsgProcess:
				c.Println(
					ctxout.ForeLightBlue,
					"PROCESS",
					ctxout.ForeBlue,
					tm,
					ctxout.CleanTag,
				)
			case tasks.MsgError:
				c.drawRow(
					tm.Target,
					ctxout.ForeLightYellow+ctxout.BoldTag+ctxout.BackRed,
					tm.Err.Error(),
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
				c.drawTwoRow(
					systools.AnyToStrNoTabs(tm),
					ctxout.CleanTag,
					ctxout.BaseSignDebug,
					ctxout.ForeLightBlue,
				)

			default:

				c.drawRow(
					ctxout.BaseSignWarning+" ",
					ctxout.ForeWhite,
					systools.AnyToStrNoTabs(tm),
					ctxout.ForeLightGrey,
					ctxout.BaseSignDebug+" ",
					ctxout.ForeLightBlue,
				)

			}
		}

		m.Unlock()
	}
}
