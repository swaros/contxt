package runner

import (
	"sync"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func (c *CmdExecutorImpl) drawRow(label, labelColor, content, contentColor, info, infoColor string) {
	c.Println(
		ctxout.OPEN_ROW,
		ctxout.TD(
			"[",
			ctxout.Prop(ctxout.ATTR_SIZE, "2"),
			ctxout.Prop(ctxout.ATTR_PREFIX, labelColor+ctxout.BoldTag+ctxout.ForeLightYellow),
			ctxout.Prop(ctxout.ATTR_ORIGIN, 2),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.TD(
			label,
			ctxout.Prop(ctxout.ATTR_SIZE, "11"),
			ctxout.Prop(ctxout.ATTR_ORIGIN, 2),
			ctxout.Prop(ctxout.ATTR_PREFIX, labelColor),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.TD(
			"]",
			ctxout.Prop(ctxout.ATTR_SIZE, "2"),
			ctxout.Prop(ctxout.ATTR_PREFIX, labelColor+ctxout.BoldTag+ctxout.ForeLightYellow),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.TD(
			content,
			ctxout.Prop(ctxout.ATTR_SIZE, "80"),
			ctxout.Prop(ctxout.ATTR_PREFIX, contentColor),
			ctxout.Prop(ctxout.ATTR_OVERFLOW, "wordwrap"),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.TD(
			info,
			ctxout.Prop(ctxout.ATTR_SIZE, "5"),
			ctxout.Prop(ctxout.ATTR_ORIGIN, 2),
			ctxout.Prop(ctxout.ATTR_PREFIX, infoColor),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.CLOSE_ROW,
	)
}

func (c *CmdExecutorImpl) drawTwoRow(content, contentColor, info, infoColor string) {
	c.Println(
		ctxout.OPEN_ROW,
		ctxout.TD(
			content,
			ctxout.Prop(ctxout.ATTR_SIZE, "95"),
			ctxout.Prop(ctxout.ATTR_PREFIX, contentColor),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.TD(
			info,
			ctxout.Prop(ctxout.ATTR_SIZE, "4"),
			ctxout.Prop(ctxout.ATTR_ORIGIN, 2),
			ctxout.Prop(ctxout.ATTR_PREFIX, infoColor),
			ctxout.Prop(ctxout.ATTR_OVERFLOW, "wrap"),
			ctxout.Prop(ctxout.ATTR_SUFFIX, ctxout.CleanTag),
		),
		ctxout.CLOSE_ROW,
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
						ctxout.ForeLightYellow+ctxout.BackCyan,
						tm.Info,
						ctxout.ForeDarkGrey+ctxout.BackLightGrey,
						"cmd",
						ctxout.ForeBlue+ctxout.BackLightBlue,
					)

				case "needs_required":
					c.drawRow(
						tm.Target,
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeDarkGrey+ctxout.BackLightGrey,
						"req",
						ctxout.ForeBlue,
					)

				case "needs_execute":
					c.drawRow(
						tm.Target,
						ctxout.ForeYellow+ctxout.BackLightGrey,
						"request to start ... "+tm.Info,
						ctxout.ForeBlue+ctxout.BackLightGrey,
						"launch",
						ctxout.ForeBlue,
					)

				case "needs_done":
					c.drawRow(
						tm.Target,
						ctxout.ForeLightCyan,
						tm.Info,
						ctxout.ForeDarkGrey+ctxout.BackLightGrey,
						"done",
						ctxout.ForeBlue,
					)

				case "wait_next_done":
					c.drawRow(
						tm.Target,
						ctxout.ForeBlue+ctxout.BackLightBlue,
						"....task finished ..."+tm.Info,
						ctxout.ForeDarkGrey+ctxout.BackLightGrey,
						"done",
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
					"cmd",
					ctxout.ForeLightCyan+ctxout.BackBlue,
				)

			default:

				c.drawRow(
					"?",
					ctxout.ForeWhite+ctxout.BackBlue,
					systools.AnyToStrNoTabs(tm),
					ctxout.ForeDarkGrey+ctxout.BackLightGrey,
					"<<<",
					ctxout.ForeLightCyan+ctxout.BackBlue,
				)

			}
		}

		m.Unlock()
	}
}
