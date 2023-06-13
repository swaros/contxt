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
				ctxout.Prop(ctxout.AttrSize, 15),
				ctxout.Prop(ctxout.AttrOrigin, 2),
				ctxout.Prop(ctxout.AttrPrefix, labelColor),
				ctxout.Prop(ctxout.AttrSuffix, ctxout.CleanTag),
				ctxout.Margin(4), // runundten 4 spaces (run + space * 2 )
			),
			ctxout.ForeYellow, "<sign stopident> ", ctxout.CleanTag,
			ctxout.TD(
				content,
				ctxout.Prop(ctxout.AttrSize, 80),
				ctxout.Prop(ctxout.AttrPrefix, contentColor),
				ctxout.Prop(ctxout.AttrOverflow, "wordwrap"),
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
				c.Print(
					ctxout.ForeLightGreen,
					"COMMAND",
					ctxout.ForeGreen,
					tm,
					ctxout.CleanTag,
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
						"launch",
						ctxout.ForeBlue,
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
						ctxout.BaseSignSuccess+tm.Target,
						ctxout.ForeLightGreen,
						" ..back from need ..."+tm.Info,
						ctxout.ForeDarkGrey,
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
				c.Println(
					ctxout.ForeLightRed,
					"ERROR",
					ctxout.ForeRed,
					tm.Error(),
					ctxout.CleanTag,
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
