package runner

import (
	"sync"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/tasks"
)

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
					c.Println(
						ctxout.ForeWhite,
						ctxout.BackLightBlue,
						" 🙭 ",
						ctxout.CleanTag,
						ctxout.ForeLightCyan,
						" ⚒ ",
						ctxout.ForeCyan,
						" [",
						ctxout.ForeYellow,
						tm.Target,
						ctxout.ForeCyan,
						"] ",

						ctxout.ForeLightBlue,
						tm.Info,
						ctxout.CleanTag,
					)
				case "needs_required":
					c.Println(
						ctxout.ForeLightCyan,
						ctxout.BackLightBlue,
						" 🙭 ",
						ctxout.CleanTag,
						ctxout.ForeCyan,
						" [",
						ctxout.ForeYellow,
						tm.Target,
						ctxout.ForeCyan,
						"] ",
						ctxout.ForeLightYellow,
						" requires ",
						ctxout.ForeDarkGrey,
						tm.Info,
						ctxout.CleanTag,
					)

				case "needs_execute":
					c.Println(
						ctxout.ForeDarkGrey,
						ctxout.BackWhite,
						" 🙭 ",
						ctxout.CleanTag,
						ctxout.ForeCyan,
						" [",
						ctxout.ForeYellow,
						tm.Target,
						ctxout.ForeCyan,
						"] ",
						ctxout.ForeLightBlue,
						tm.Info,
						ctxout.CleanTag,
					)
				case "needs_done":
					c.Println(
						ctxout.ForeGreen,
						ctxout.BackDarkGrey,
						" 🙭 ",
						" ✓ ",
						ctxout.CleanTag,
						ctxout.ForeCyan,
						" [",
						ctxout.ForeYellow,
						tm.Target,
						ctxout.ForeCyan,
						"]",
						ctxout.ForeDarkGrey,
						" (all needs are done) ",
						ctxout.CleanTag,
					)

				case "wait_next_done":
					c.Println(
						ctxout.ForeWhite,
						ctxout.BackLightBlue,
						" 🙭 ",
						ctxout.CleanTag,
						ctxout.ForeLightGreen,
						" ✓ ",
						ctxout.ForeCyan,
						" [",
						ctxout.ForeYellow,
						tm.Target,
						ctxout.ForeCyan,
						"]",
						ctxout.ForeDarkGrey,
						" ⏲ ",
						ctxout.CleanTag,
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
				c.Println(
					ctxout.ForeBlue,
					"       ⌨  >> ",
					ctxout.CleanTag,
					tm,
					ctxout.ForeBlue,
					" << 🎧 ",
					ctxout.CleanTag,
				)
			default:

				c.Println(
					ctxout.ForeDarkGrey,
					"UNKNOWN",
					ctxout.ForeCyan,
					tm,
					ctxout.CleanTag,
				)

			}
		}

		m.Unlock()
	}
}
