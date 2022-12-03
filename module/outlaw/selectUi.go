package outlaw

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type selectItem struct {
	title, desc string
}

type selectResult struct {
	isSelected bool
	item       selectItem
}

var (
	selected selectResult
	items    []selectItem
)

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.desc }
func (i selectItem) FilterValue() string { return i.title }

type model struct {
	list list.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m model) View() string {
	return docStyle.Render(m.list.View())
}

func ClearSelectItems() {
	items = []selectItem{}
}

func AddItemToSelect(item selectItem) {
	items = append(items, item)
}

func uIselectItem(title string) selectResult {
	displayItems := []list.Item{}
	selected.isSelected = false

	for _, itm := range items {
		displayItems = append(displayItems, itm)
	}

	listModel := model{list: list.New(displayItems, list.NewDefaultDelegate(), 0, 0)}
	listModel.list.Title = title

	p := tea.NewProgram(listModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	ClearSelectItems()
	return selected
}
