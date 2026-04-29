package ui

import (
	"anki-sync-go/internal/status"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			Bold(true)
	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF7F87"))
	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#767676"))
)

type TUI struct {
	bus         *MessageBus
	step        status.Step
	stepMsg     string
	progress    progress.Model
	spinner     spinner.Model
	logs        []string
	err         error
	quitting    bool
	dryRun      bool
	planSummary string
}

func NewTUI(dryRun bool) *TUI {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &TUI{
		step:     status.Idle,
		progress: progress.New(progress.WithDefaultGradient()),
		spinner:  s,
		bus: &MessageBus{
			incoming: make(chan status.Event, 100),
			confirm:  make(chan bool),
		},
		dryRun: dryRun,
	}
}

type MessageBus struct {
	incoming chan status.Event
	confirm  chan bool
}

func (b *MessageBus) WaitForProgress() tea.Cmd {
	return func() tea.Msg {
		return <-b.incoming
	}
}

func (b *MessageBus) SendProgress(msg status.Event) {
	select {
	case b.incoming <- msg:
	default:
	}

}

func (b *MessageBus) WaitForConfirmation() bool {
	return <-b.confirm
}

func (b *MessageBus) SendConfirmation(approved bool) {
	select {
	case b.confirm <- approved:
	default:
	}
}

func (m *TUI) Confirm(summary string) bool {
	return m.bus.WaitForConfirmation()
}

func (m *TUI) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.bus.WaitForProgress(),
	)
}

func (m *TUI) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		m.quitting = true
		if m.awaitingConfirm() {
			// If we quit during planning/confirmation, send false
			m.bus.SendConfirmation(false)
		}
		return m, tea.Quit
	case "y", "Y":
		if m.awaitingConfirm() {
			m.bus.SendConfirmation(true)
		}
	case "n", "N":
		if m.awaitingConfirm() {
			m.bus.SendConfirmation(false)
			m.quitting = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case status.Event:
		var cmd tea.Cmd

		switch msg := msg.(type) {
		case status.PhaseMsg:
			m.step = msg.Step
			m.stepMsg = msg.Label

		case status.LogMsg:
			if msg.Text != "" {
				m.logs = append(m.logs, msg.Text)
				if len(m.logs) > 10 {
					m.logs = m.logs[1:]
				}
			}
		case status.ProgressMsg:
			if msg.Total > 0 {
				cmd = m.progress.SetPercent(float64(msg.Current) / float64(msg.Total))
			}
		case status.PlanMsg:
			if msg.Summary != "" {
				m.planSummary = msg.Summary
				m.step = status.Planning
			}
		case status.ErrMsg:
			if msg.Error != nil {
				m.err = msg.Error
				m.step = status.Error
			}
		}

		if m.step == status.Done || m.step == status.Error {
			m.quitting = true
			return m, tea.Sequence(cmd, tea.Quit)
		}

		// Don't wait for next message if we are waiting for confirmation
		if m.awaitingConfirm() {
			return m, cmd
		}

		return m, tea.Batch(cmd, m.bus.WaitForProgress())

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case progress.FrameMsg:
		newModel, cmd := m.progress.Update(msg)
		if pm, ok := newModel.(progress.Model); ok {
			m.progress = pm
		}
		return m, cmd
	}

	return m, nil
}

func (m *TUI) View() string {
	if m.err != nil {
		return fmt.Sprintf("\n %s\n\n", statusStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	}

	var s strings.Builder

	m.renderStatusMsg(&s)

	if m.awaitingConfirm() {
		m.renderConfirmMsg(&s)
	} else {
		m.renderLogs(&s)
	}

	if m.quitting {
		s.WriteString("\n Done!\n")
	} else {
		s.WriteString("\n Press q to quit\n")
	}

	return s.String()
}

func (m *TUI) awaitingConfirm() bool {
	return m.step == status.Planning && !m.dryRun && m.planSummary != ""
}

func (m *TUI) renderStatusMsg(s *strings.Builder) {
	s.WriteString("\n")
	s.WriteString(titleStyle.Render(" Anki Sync "))
	s.WriteString("  " + m.stepMsg + "\n\n")

	switch m.step {
	case status.Validating, status.Fetching:
		s.WriteString(m.spinner.View() + " Working...\n\n")
	case status.Executing:
		s.WriteString(m.progress.View() + "\n\n")
	}
}

func (m *TUI) renderConfirmMsg(s *strings.Builder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("Confirm sync? [y/N]\n\n"))
	s.WriteString(lipgloss.NewStyle().Render(m.planSummary) + "\n\n")
}

func (m *TUI) renderLogs(s *strings.Builder) {
	if len(m.logs) > 0 {
		s.WriteString("Recent logs:\n")
		for _, log := range m.logs {
			s.WriteString(logStyle.Render(" • "+log) + "\n")
		}
	}
}

func (m *TUI) Report(msg status.Event) {
	m.bus.SendProgress(msg)
}
