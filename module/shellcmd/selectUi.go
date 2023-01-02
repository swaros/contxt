package shellcmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	items []selectItem
)

func (i selectItem) Title() string       { return i.title }
func (i selectItem) Description() string { return i.desc }
func (i selectItem) FilterValue() string { return i.title }

type selectUiModel struct {
	list list.Model
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
	return docStyle.Render(m.list.View())
}

func ClearSelectItems() {
	items = []selectItem{}
}

func AddItemToSelect(item selectItem) {
	items = append(items, item)
}

func uIselectItem(title string, asMenu bool) selectResult {
	displayItems := []list.Item{}
	selected.isSelected = false

	for _, itm := range items {
		displayItems = append(displayItems, itm)
	}

	listModel := selectUiModel{list: list.New(displayItems, list.NewDefaultDelegate(), 0, 0)}
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
