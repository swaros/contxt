package shellcmd

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/swaros/contxt/module/systools"
)

type LogOutput struct {
	buffer     []string
	max        int
	maxWith    int
	autoSize   bool
	maxEntries int
}

var (
	logOutputStyle = lipgloss.NewStyle().
			Margin(0, 0, 0, 0).
			Align(lipgloss.Left).
			Border(lipgloss.NormalBorder(), true, true, true, true).
			Background(lipgloss.Color("#222222")).
			BorderForeground(lipgloss.Color("#333333"))

	rowStyle = lipgloss.NewStyle().Align(lipgloss.Left).Foreground(lipgloss.Color("#aaaaaa"))
)

func NewLogOutput(max int, maxWidth int) *LogOutput {
	return &LogOutput{
		buffer:   make([]string, 0, max),
		max:      max,
		maxWith:  maxWidth,
		autoSize: false,
	}
}

func NewAutoSizeLogOutput() *LogOutput {
	asize := true

	w, h, err := systools.GetStdOutTermSize()
	if err != nil {
		w = 80
		h = 24
		asize = false
	}
	//logOutputStyle = logOutputStyle.MaxHeight(h - 2).MaxWidth(w / 2)
	//logOutputStyle.Width(w / 2)
	maxRows := 100 + 4
	return &LogOutput{
		maxEntries: maxRows,
		buffer:     make([]string, 0, maxRows),
		max:        h - 4,
		maxWith:    (w / 2) - 2,
		autoSize:   asize,
	}
}

func (l *LogOutput) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return l, tea.Quit
		}
	}
	return l, nil
}

func (l *LogOutput) Init() tea.Cmd {
	return nil
}

func (l *LogOutput) View() string {
	outPut := ""
	end := len(l.buffer)
	if end == 0 {
		return "... still empty ..."
	}
	start := end - l.max
	if start < 0 {
		start = 0
	}
	for i := start; i < end; i++ {
		if len(l.buffer[i]) > 0 {
			str := systools.PrintableChars(l.buffer[i])
			outPut += rowStyle.Inline(true).MaxWidth(l.maxWith).Render(str) + "\n"
		}
	}
	return logOutputStyle.Render(outPut)
}

func (l *LogOutput) SetSize(maxWidth, maxRowsShown int) {
	l.max = maxRowsShown
	l.maxWith = maxWidth

	//logOutputStyle = logOutputStyle.MaxHeight(maxRowsShown + 2).MaxWidth(maxWidth + 2)
	logOutputStyle.Width(maxWidth)
	logOutputStyle.Height(maxRowsShown)

}

// Add adds a string to the buffer
// if the string contains a newline, it will be split and added as multiple lines
// if the buffer is full, the oldest line will be removed
// if the string is empty, nothing will be added
// if the string is a newline, nothing will be added
func (l *LogOutput) Add(s string) {
	if s == "" {
		return
	}
	preCheck := strings.Split(s, "\n")
	if len(preCheck) > 1 {
		for _, v := range preCheck {
			l.Add(v)
		}
		return
	}

	l.buffer = append(l.buffer, s)
	if len(l.buffer) > l.max {
		l.buffer = l.buffer[1:]
	}
}

func (l *LogOutput) GetBuffer() []string {
	return l.buffer
}

func (l *LogOutput) Clear() {
	l.buffer = make([]string, 0, l.max)
}

func (l *LogOutput) GetMax() int {
	return l.max
}

func (l *LogOutput) SetMax(max int) {
	l.max = max
}

func (l *LogOutput) GetSize() int {
	return len(l.buffer)
}
