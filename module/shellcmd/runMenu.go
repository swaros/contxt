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

 package shellcmd

import (
	"fmt"
	"io"
	"math"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/contxt/module/trigger"
)

var (
	// the style for the right side (log) of the screen
	leftOutputStyle = lipgloss.NewStyle().Margin(0, 0).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color("#333333"))

	// the style for the left side (menu) of the screen
	menuStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color("#333333"))

	selectedMenuItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("214"))
	wasRunningStyle       = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("64"))
	isRunningStyle        = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("219"))
	regularItemStyle      = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("240"))
	errorStyle            = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("196"))
)

type CmdMenuItem struct {
	Name        string
	Running     bool
	Selected    bool
	RunCount    int
	UpdateCount int
	Blocked     bool
	HaveError   bool
}

type RundCmd struct {
	targets []string
	menu    list.Model
	log     *LogOutput
}

type updateMsg struct {
	content string
	origin  any
}

// model functions for the runMenu
func (i CmdMenuItem) Title() string       { return i.Name }
func (i CmdMenuItem) Description() string { return "" }
func (i CmdMenuItem) FilterValue() string { return i.Name }

// create oure own delegate to render the menu items
type RunMenuDelegate struct{}

func (d RunMenuDelegate) Height() int                               { return 1 }
func (d RunMenuDelegate) Spacing() int                              { return 0 }
func (d RunMenuDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d RunMenuDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {

	i, ok := listItem.(CmdMenuItem)
	if !ok {
		return
	}

	//qstr := fmt.Sprintf("  %s [%v] (%v)", i.Name, i.UpdateCount, i.Running)
	str := fmt.Sprintf("  %s", i.Name)

	// actual selected item
	selected := m.Index() == index
	prefix := "  "
	mStyle := regularItemStyle

	if selected { // must be in front of other checks to get at least the selected style once for any item that did nothing
		mStyle = selectedMenuItemStyle.Copy()
	}

	if i.Running {
		mStyle = isRunningStyle.Copy()
		prefix = "[]"
		if i.Running && i.UpdateCount > 0 {

			progressLine := []string{"⠷", "⠾", "⠦", "⠿", "⠹", "⠸", "⠼", "⠴"}
			modulo := math.Mod(float64(i.UpdateCount), float64(len(progressLine)))
			prefix = progressLine[int(modulo)] + " "

		}
	}

	if i.RunCount > 0 && !i.Running {
		mStyle = wasRunningStyle.Copy()
	}

	if selected {
		if !i.Running && prefix != "[]" {
			prefix = "->"
		}
		mStyle = mStyle.Copy().Bold(true)
	}

	if i.Blocked {
		prefix = "[]"
	}

	if i.HaveError {
		mStyle = errorStyle.Copy()
	}

	fmt.Fprint(w, mStyle.Render(prefix+str))

}

func NewRunMenu(targets []string, log LogOutput) RundCmd {
	displayItems := []list.Item{}
	for _, t := range targets {
		displayItems = append(displayItems, CmdMenuItem{Name: t, Running: false, Selected: false})
	}
	if w, h, err := systools.GetStdOutTermSize(); err == nil && h > 10 {
		log.SetSize(w/2, h-6)
	}

	menuList := list.NewModel(displayItems, RunMenuDelegate{}, 0, 0)
	menuList.Title = "Select a target to run"

	return RundCmd{
		targets: targets,
		menu:    menuList,
		log:     &log,
	}
}

func (m RundCmd) Init() tea.Cmd {
	return spinner.Tick
}

func (m RundCmd) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "enter" {
			if itm := m.menu.SelectedItem(); itm != nil {
				if i, ok := itm.(CmdMenuItem); ok {
					i.Selected = !i.Selected
					if !i.Running && !i.Blocked {
						i.Blocked = true // to prevent 'double' runs by keypresses. will be removed by the taskrun.EventTaskStatus
						go taskrun.RunTargets(i.Name, true)
						m.updateMenuItem(i)
					}
				}

			}
		}

	case updateMsg:
		switch ctxmsg := msg.origin.(type) {
		case taskrun.EventTaskStatus:
			if itm, found := m.findItemByName(ctxmsg.Target); found {
				itm.RunCount = ctxmsg.RunCount
				itm.Running = ctxmsg.Running
				itm.Blocked = false
				m.updateMenuItem(itm)
			}

		case taskrun.EventScriptLine:
			m.log.Update(msg.content)
			// update menuitem if running
			if itm, found := m.findItemByName(ctxmsg.Target); found {
				m.updateMenuItem(itm)
			}
			// update error status
			if ctxmsg.Error != nil {
				if itm, found := m.findItemByName(ctxmsg.Target); found {
					itm.HaveError = true
					m.updateMenuItem(itm)
				}
			}
		}
	case tea.WindowSizeMsg:
		h, v := menuStyle.GetFrameSize()          // get the frame additional needed space for the menu
		m.menu.SetSize(msg.Width-h, msg.Height-v) // update the menu size
		lh, lv := leftOutputStyle.GetFrameSize()  // get the frame additional needed space for the log-container
		m.log.Add(fmt.Sprintf("Window size changed to %v x %v  menu is %v", msg.Width, msg.Height, lipgloss.Width(m.menu.View())))

		rightSpaceWidth := msg.Width - (h + lh + lipgloss.Width(m.menu.View()))
		leftOutputStyle.Width(rightSpaceWidth)
		leftOutputStyle.Height(msg.Height - lv)

		m.log.SetSize(rightSpaceWidth-5, msg.Height-lv-5)

	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m RundCmd) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuStyle.Render(m.menu.View()),
		leftOutputStyle.Render(m.log.View()),
	)

}

func (m RundCmd) findItemByName(name string) (CmdMenuItem, bool) {
	for _, itm := range m.menu.Items() {
		if i, ok := itm.(CmdMenuItem); ok {
			if i.Name == name {
				return i, true
			}
		}
	}
	return CmdMenuItem{}, false
}

func (m RundCmd) updateMenuItem(item CmdMenuItem) {
	for itmIndex, itm := range m.menu.Items() {
		if i, ok := itm.(CmdMenuItem); ok {
			if i.Name == item.Name {
				item.UpdateCount++
				if item.UpdateCount > 1000 {
					item.UpdateCount = 1
				}
				updMsg := m.menu.SetItem(itmIndex, item)
				m.menu.Update(updMsg)
			}
		}
	}
}

func (m RundCmd) registerEvent(p *tea.Program) {
	exechandler := trigger.NewListener("runListener", func(any ...interface{}) {
		if len(any) > 0 {
			for _, line := range any {
				switch line := line.(type) {
				case taskrun.EventScriptLine:
					msg := fmt.Sprintf("[%s]:%s", line.Target, line.Line)
					m.log.Add(msg)
					p.Send(updateMsg{content: msg, origin: line})
				}
			}
		}
	})

	exechandler.RegisterToEvent(taskrun.EventAllLines)

	statusTrigger := trigger.NewListener("statusListener", func(any ...interface{}) {
		if len(any) > 0 {
			for _, line := range any {
				switch line := line.(type) {
				case taskrun.EventTaskStatus:
					msg := fmt.Sprintf(" >>> update from %s ", line.Target)
					m.log.Add(msg)
					p.Send(updateMsg{content: msg, origin: line})
				}
			}
		}
	})

	statusTrigger.RegisterToEvent(taskrun.EventTaskStatusUpdate)

}

func (m RundCmd) Run() (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseAllMotion())
	m.registerEvent(p)
	taskrun.PreHook = func(msg ...interface{}) bool {
		return true
	}
	model, err := p.StartReturningModel()

	taskrun.PreHook = nil
	if err != nil {
		return nil, err
	} else {
		return model, nil
	}

}

func (m RundCmd) GetSelectedIndex() int {
	return m.menu.Index()
}
