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
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swaros/contxt/module/systools"
)

var (
	docStyle = lipgloss.NewStyle().Margin(1, 2)

	menuTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065")).Bold(true).
			Padding(0, 1)

	selectionTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#ccccFF")).Bold(true).
				Padding(0, 1)
)

type selectItem struct {
	title, desc string
}

var (
	items      []selectItem
	logPointer LogOutput
)

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.desc }
func (i selectItem) FilterValue() string { return i.title }

type selectUiModel struct {
	list list.Model
	log  LogOutput
}

func (m selectUiModel) Init() tea.Cmd {
	return nil
}

func (m selectUiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if msg.String() == "enter" {

			if itm := m.list.SelectedItem(); itm != nil { // just to check if we have something selected
				selected.item = items[m.list.Index()]
				selected.isSelected = true
				return m, tea.Quit
			}

		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectUiModel) View() string {
	return lipgloss.JoinHorizontal(lipgloss.Top, docStyle.Render(m.list.View()), m.log.View())
}

func ClearSelectItems() {
	items = []selectItem{}
}

func AddItemToSelect(item selectItem) {
	items = append(items, item)
}

func ApplyLogOut(log LogOutput) {
	logPointer = log
}

func uIselectItem(title string, asMenu bool) selectResult {
	displayItems := []list.Item{}
	selected.isSelected = false

	for _, itm := range items {
		displayItems = append(displayItems, itm)
	}
	w, h, _ := systools.GetStdOutTermSize()
	listModel := selectUiModel{list: list.New(displayItems, list.NewDefaultDelegate(), w/2, h-3)}
	listModel.log = logPointer
	listModel.list.Title = title
	if asMenu {
		listModel.list.Styles.Title = menuTitleStyle
	} else {
		listModel.list.Styles.Title = selectionTitleStyle
	}

	p := tea.NewProgram(listModel, tea.WithAltScreen())

	if _, err := p.StartReturningModel(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	ClearSelectItems()
	return selected
}
