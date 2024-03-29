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
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxtcell"
)

func injectedTcell(c *CmdExecutorImpl, initFn func(caller *ctxtcell.CtCell, c *CmdExecutorImpl)) (*ctxtcell.CtCell, error) {
	tc := ctxtcell.NewTcell()
	tc.SetMouse(true).SetNoClearScreen(false)
	outHndl := ctxtcell.NewCtOutput(tc)
	c.session.OutPutHdnl = outHndl
	c.session.Printer = nil
	if err := tc.Init(); err != nil {
		return nil, err
	}
	initFn(tc, c)
	if err := tc.Run(); err != nil {
		return nil, err
	}
	return tc, nil
}

func MainScreen(c *CmdExecutorImpl) (*ctxtcell.CtCell, error) {
	return injectedTcell(c, func(caller *ctxtcell.CtCell, c *CmdExecutorImpl) {
		MainMenu(caller, c)
	})
}

func MainMenu(tc *ctxtcell.CtCell, c *CmdExecutorImpl) {
	menu := tc.NewMenu()

	// top bar
	contxtTopMenu := tc.ActiveText("contxt")
	contxtTopMenu.SetPos(1, 0).SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGoldenrod).Background(tcell.ColorBlack))
	contxtTopMenu.OnSelect = func(selected bool) {
		menu.SetVisible(!menu.IsVisible())
	}
	tc.AddElement(contxtTopMenu)

	exitTopMenu := tc.ActiveText("exit")
	exitTopMenu.SetPosProcentage(100, 0).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGoldenrod).Background(tcell.ColorBlack))

	exitTopMenu.GetPos().SetMargin(-5, 0)
	exitTopMenu.OnSelect = func(selected bool) {
		tc.Stop()
	}
	tc.AddElement(exitTopMenu)

	menu.SetTopLeft(1, 1).SetBottomRight(20, 10)

	menu.AddItem("PrintPaths", func(itm *ctxtcell.MenuElement) {
		itm.GetText().SetText("PrintPaths RUNS")
		c.GetLogger().Debug("run command: PrintPaths")
		c.Println("run command: ", ctxout.ForeLightBlue, "PrintPaths")
		runInfos := ctxout.GetRunInfos()
		c.PrintPaths(false, false)
		addTxt := strings.Join(runInfos, "|")
		itm.GetText().SetText("PrintPaths:" + addTxt)
		c.Println(addTxt)
	})

	// add cobra commands to menu
	/*
		for _, cmd := range c.session.Cobra.RootCmd.Commands() {
			menu.AddItemWithRef(cmd.Name(), cmd, func(itm *ctxtcell.MenuElement) {
				itm.GetText().SetText("RUNS")
				cmdIntern := itm.GetReference().(*cobra.Command)
				ctxout.CtxOut(c.session.OutPutHdnl, "run command: ", ctxout.ForeLightBlue, cmdIntern.Name())
				cmdIntern.Run(cmdIntern, []string{})
				itm.GetText().SetText(cmdIntern.Name() + " done")
			})
		}
	*/
	menu.SetVisible(false)
	tc.AddElement(menu)
}
