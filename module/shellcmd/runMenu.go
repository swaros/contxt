package shellcmd

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
	"github.com/swaros/contxt/module/trigger"
)

var (
	logOutStyle = lipgloss.NewStyle().Margin(0, 0)
	menuStyle   = lipgloss.NewStyle().Border(lipgloss.NormalBorder(), true, true, true, true).
			BorderForeground(lipgloss.Color("#333333"))
)

type CmdMenuItem struct {
	Name     string
	Running  bool
	Selected bool
}

type RundCmd struct {
	targets []string
	menu    list.Model
	log     *LogOutput
	spinner spinner.Model
}

type updateMsg struct {
	duration time.Duration
	content  string
	origin   any
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

	str := fmt.Sprintf("  %s", i.Name)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s string) string {
			return selectedItemStyle.Render(">>" + s)
		}
	}

	fmt.Fprint(w, fn(str))
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
	spin := spinner.New()
	return RundCmd{
		targets: targets,
		menu:    menuList,
		log:     &log,
		spinner: spin,
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
					if !i.Running {
						i.Running = true
						go taskrun.RunTargets(i.Name, true)
					}
				}

			}
		}

	case updateMsg:
		m.log.Update(msg.content)
		m.spinner.Tick()
		m.spinner, _ = m.spinner.Update(msg)

	case tea.WindowSizeMsg:
		h, v := menuStyle.GetFrameSize()
		m.menu.SetSize(msg.Width-h, msg.Height-v)
		//lh, lw := logOutStyle.GetFrameSize()
		//m.log.SetSize(msg.Width-lh, msg.Height-lw)

	case spinner.TickMsg:
		m.spinner, _ = m.spinner.Update(msg)
	}

	var cmd tea.Cmd
	m.menu, cmd = m.menu.Update(msg)
	return m, cmd
}

func (m RundCmd) View() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		menuStyle.Render(m.menu.View()),
		logOutStyle.Render(m.log.View()),
	)

}

func (m RundCmd) registerEvent(p *tea.Program) {
	exechandler := trigger.NewListener("runListener", func(any ...interface{}) {
		if len(any) > 0 {
			for _, line := range any {
				switch line := line.(type) {
				case taskrun.EventScriptLine:
					msg := fmt.Sprintf("[%s]:%s", line.Target, line.Line)
					m.log.Add(msg)
					p.Send(updateMsg{content: msg, duration: time.Duration(50 * time.Millisecond)})
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
					msg := fmt.Sprintf(" ++++ %s +++ ", line.Target)
					m.log.Add(msg)
					p.Send(updateMsg{content: msg, duration: time.Duration(50 * time.Millisecond)})
				}
			}
		}
	})

	statusTrigger.RegisterToEvent(taskrun.EventTaskStatusUpdate)

}

func (m RundCmd) Run() (tea.Model, error) {
	p := tea.NewProgram(m, tea.WithAltScreen())
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
