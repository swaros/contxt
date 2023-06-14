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
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return "" }

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 1 }
func (d itemDelegate) Spacing() int                              { return 0 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("  %s", i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render(">>" + s)
		}
	}

	fmt.Fprint(w, fn(str))
}

type simpleSelectModel struct {
	list     list.Model
	choice   string
	quitting bool
}

func (m simpleSelectModel) Init() tea.Cmd {
	return nil
}

func (m simpleSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "esc", "q":
			taskrun.GetLogger().WithField("status", selected).Debug("hit a cancel button")
			m.quitting = true
			selected.aborted = true
			selected.isSelected = false
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				selected.isSelected = true
				selected.aborted = false
				selected.item.desc = ""
				selected.item.title = string(i)
			}
			return m, tea.Quit
			// if some wierd happen again
			/*default:
			taskrun.GetLogger().WithField("msg", msg.String()).Debug("hit a key")*/
		}

	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m simpleSelectModel) View() string {
	taskrun.GetLogger().WithField("status", selected).Debug("update status")
	if selected.isSelected {
		return quitTextStyle.Render(fmt.Sprintf("%s .", selected.item.title))
	}
	if selected.aborted {
		return quitTextStyle.Render("nothing selected... get out")
	}
	return "\n" + m.list.View()
}

func simpleSelect(title string, selectable []string) selectResult {
	selected.isSelected = false
	selected.aborted = false
	items := []list.Item{}

	for _, sel := range selectable {
		items = append(items, item(sel))
	}
	w, h, _ := systools.GetStdOutTermSize()

	l := list.New(items, itemDelegate{}, w, h-2)
	l.Title = title
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := simpleSelectModel{list: l}

	if _, err := tea.NewProgram(m, tea.WithAltScreen()).StartReturningModel(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}

	return selected
}
