package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	spinnerStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Margin(1, 0)
	dotStyle      = helpStyle.Copy().UnsetMargins()
	durationStyle = dotStyle.Copy()
	appStyle      = lipgloss.NewStyle().Margin(1, 2, 0, 2)
)

type ResultMsg struct {
	Msg   string
	Ready bool
}

func (r ResultMsg) String() string {
	if r.Msg == "" {
		return dotStyle.Render(fmt.Sprintf("\t%s", strings.Repeat(".", 30)))
	}

	if r.Ready {
		return ""
	}

	return fmt.Sprintf("\t%s", r.Msg)
}

type Model struct {
	msg      string
	spinner  spinner.Model
	results  []ResultMsg
	quitting bool
}

func NewModel(msg string) Model {
	const numLastResults = 1
	s := spinner.New()
	s.Spinner = spinner.Jump
	s.Style = spinnerStyle

	return Model{
		spinner: s,
		results: make([]ResultMsg, numLastResults),
		msg:     msg,
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true

			return m, tea.Quit
		default:
			return m, nil
		}

	case ResultMsg:
		if msg.Ready {
			m.quitting = true

			return m, tea.Quit
		}

		m.results = append(m.results[1:], msg)

		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)

		return m, cmd

	default:
		return m, nil
	}
}

func (m Model) View() string {
	var s string

	if !m.quitting {
		s += fmt.Sprintf("%s\t%s", m.spinner.View(), m.msg)
	}

	s += "\n\n"

	for _, res := range m.results {
		s += res.String() + "\n"
	}

	if !m.quitting {
		s += helpStyle.Render("\tPress q to quit")
	}

	if m.quitting {
		// s += "\n"
		s = ""
	}

	return appStyle.Render(s)
}
