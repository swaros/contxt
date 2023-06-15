// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package runner

import (
	"fmt"
	"os"
	"runtime"

	"github.com/sirupsen/logrus"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

// Init initializes the application
// and starts the main loop
func Init() error {
	// create the application session
	app := NewCmdSession()
	// set the default log level
	app.Log.Logger.SetLevel(logrus.ErrorLevel)
	// create the the command executor instance
	functions := NewCmd(app)

	// add support for utf-8 signs
	glyps := ctxout.NewSignFilter(nil)
	ctxout.AddPostFilter(glyps)

	// enable the sign filter if possible
	// in current ctxout version, must be done before NewTabOut
	if runtime.GOOS != "windows" && systools.IsStdOutTerminal() {
		// check if unicode is supported. if not, disable the sign filter
		// this is the only way i see to check for unicode support
		code1, code2, errorEx := tasks.Execute("bash", []string{"-c"}, "echo -e $TERM", func(s string, err error) bool {
			if err == nil {
				terminalsTheyNotSupportUt8Chars := []string{"xterm", "screen", "tmux"}
				// if one of these terminals is used, we do not enable the sign filter
				for _, term := range terminalsTheyNotSupportUt8Chars {
					if s == term {
						return false
					}
				}
			}
			glyps.Enable()
			return true
		}, func(p *os.Process) {

		})
		// just to be sure, if anything gos wrong by checking the terminal, we disable the sign filter
		if errorEx != nil && code1 == 0 && code2 == 0 {
			glyps.Disable()
		} else {
			glyps.AddSign(ctxout.Sign{Glyph: "🭬", Name: "runident", Fallback: ">"})
			glyps.AddSign(ctxout.Sign{Glyph: "🭮", Name: "stopident", Fallback: ">"})
		}
	} else {
		fmt.Println("no unicode support")
	}

	// set the default output filter
	ctxout.AddPostFilter(ctxout.NewTabOut())

	// initialize the cobra commands
	if err := app.Cobra.Init(functions); err != nil {
		return err
	}
	// and execute the root command
	if err := app.Cobra.RootCmd.Execute(); err != nil {
		return err
	}
	return nil
}
