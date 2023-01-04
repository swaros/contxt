package shellcmd

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).Bold(true).
		Padding(0, 1)
)

type TextInputModel struct {
	Label    string
	txtInput textinput.Model
	Value    string
	err      error
}

type (
	txtErrMsg error
)

func TextInput(label string, placeHolder string, limit int, width int) (string, error) {
	model := InitTextInput(placeHolder, 128, 25)
	model.Label = label
	p := tea.NewProgram(&model, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		return "", err
	}

	return model.GetValue(), nil
}

func InitTextInput(placeHolder string, limit int, width int) TextInputModel {
	txt := textinput.New()
	txt.Placeholder = placeHolder
	txt.Focus()
	txt.CharLimit = limit
	txt.Width = width

	return TextInputModel{
		txtInput: txt,
		err:      nil,
	}
}

func (m *TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter, tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case txtErrMsg:
		m.err = msg
		return m, nil
	}

	m.txtInput, cmd = m.txtInput.Update(msg)
	return m, cmd
}

func (m *TextInputModel) GetValue() string {
	return m.txtInput.Value()
}

func (m *TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *TextInputModel) View() string {

	return fmt.Sprintf(
		"%s\n\n%s\n\n%s",
		labelStyle.Render(m.Label),
		m.txtInput.View(),
		"(esc to quit)",
	) + "\n"
}
