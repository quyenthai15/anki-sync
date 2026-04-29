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
			Foreground(lipgloss.Color("#FF5F87"))
	logStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#767676"))
)

type TUI struct {
	status      status.SyncStatus
	stage       string
	progress    progress.Model
	spinner     spinner.Model
	logs        []string
	err         error
	msgs        chan status.ProgressMsg
	confirmChan chan bool
	quitting    bool
	dryRun      bool
	planSummary string
}

func NewTUI(dryRun bool) *TUI {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return &TUI{
		status:      status.StatusIdle,
		progress:    progress.New(progress.WithDefaultGradient()),
		spinner:     s,
		msgs:        make(chan status.ProgressMsg, 100),
		confirmChan: make(chan bool),
		dryRun:      dryRun,
	}
}

func (m *TUI) Confirm(summary string) bool {
	return <-m.confirmChan
}

func (m *TUI) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.waitForMsg(),
	)
}

func (m *TUI) waitForMsg() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgs
	}
}

func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			if m.status == status.StatusPlanning && !m.dryRun {
				// If we quit during planning/confirmation, send false
				select {
				case m.confirmChan <- false:
				default:
				}
			}
			return m, tea.Quit
		case "y", "Y":
			if m.status == status.StatusPlanning && !m.dryRun {
				select {
				case m.confirmChan <- true:
				default:
				}
			}
		case "n", "N":
			if m.status == status.StatusPlanning && !m.dryRun {
				select {
				case m.confirmChan <- false:
				default:
				}
				m.quitting = true
				return m, tea.Quit
			}
		}

	case status.ProgressMsg:
		m.status = msg.Status
		if msg.Stage != "" {
			m.stage = msg.Stage
		}
		if msg.Log != "" {
			if strings.HasPrefix(msg.Log, "Built plan:") {
				m.planSummary = msg.Log
			}
			m.logs = append(m.logs, msg.Log)
			if len(m.logs) > 10 {
				m.logs = m.logs[1:]
			}
		}
		if msg.Error != nil {
			m.err = msg.Error
			m.status = status.StatusError
		}

		var cmd tea.Cmd
		if msg.Total > 0 {
			cmd = m.progress.SetPercent(float64(msg.Current) / float64(msg.Total))
		}

		if m.status == status.StatusDone || m.status == status.StatusError {
			m.quitting = true
			return m, tea.Sequence(cmd, tea.Quit)
		}

		// Don't wait for next message if we are waiting for confirmation
		if m.awaitingConfirm() {
			return m, cmd
		}

		return m, tea.Batch(cmd, m.waitForMsg())

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
	}

	m.renderLogs(&s)

	if m.quitting {
		s.WriteString("\n Done!\n")
	} else if m.awaitingConfirm() {
		// Instructions are in the confirmation prompt
	} else {
		s.WriteString("\n Press q to quit\n")
	}

	return s.String()
}

func (m *TUI) awaitingConfirm() bool {
	return m.status == status.StatusPlanning && !m.dryRun && m.planSummary != ""
}

func (m *TUI) renderStatusMsg(s *strings.Builder) {
	s.WriteString("\n")
	s.WriteString(titleStyle.Render(" Anki Sync "))
	s.WriteString("  " + m.stage + "\n\n")

	switch m.status {
	case status.StatusValidating, status.StatusFetching:
		s.WriteString(m.spinner.View() + " Working...\n\n")
	case status.StatusPlanning, status.StatusExecuting:
		s.WriteString(m.progress.View() + "\n\n")
	}
}

func (m *TUI) renderConfirmMsg(s *strings.Builder) {
	s.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("Confirm sync? [y/N]\n"))
	s.WriteString(lipgloss.NewStyle().Italic(true).Render(m.planSummary) + "\n\n")
}

func (m *TUI) renderLogs(s *strings.Builder) {
	if len(m.logs) > 0 {
		s.WriteString("Recent logs:\n")
		for _, log := range m.logs {
			s.WriteString(logStyle.Render(" • "+log) + "\n")
		}
	}
}

func (m *TUI) Report(msg status.ProgressMsg) {
	m.msgs <- msg
}
