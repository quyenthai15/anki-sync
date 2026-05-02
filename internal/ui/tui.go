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

	planHeaderStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("75"))
	planAddStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("42"))
	planUpdateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	planItemStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
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

func NewTUI() *TUI {
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
	}
}

type MessageBus struct {
	incoming chan status.Event
	confirm  chan bool
}

func (b *MessageBus) WaitForProgress() tea.Cmd {
	return func() tea.Msg {
		// This blocks until an event is received, then returns it as a tea.Msg
		return <-b.incoming
	}
}

func (b *MessageBus) SendProgress(msg status.Event) {
	// Use a non-blocking send to ensure the business logic never hangs
	// if the TUI is busy or lagging.
	select {
	case b.incoming <- msg:
	default:
		// If the buffer is full, we drop the message. For UI progress,
		// this is usually acceptable to keep the system responsive.
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
	m.bus.SendProgress(status.ConfirmMsg{Summary: summary})
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
	case "q", "ctrl+c", "esc":
		m.quitting = true
		if m.awaitingConfirm() {
			m.bus.SendConfirmation(false)
		}
		return m, tea.Quit

	case "y", "Y":
		if m.awaitingConfirm() {
			m.bus.SendConfirmation(true)
			return m, m.bus.WaitForProgress()
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

func (m *TUI) handleEvent(msg status.Event) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case status.PhaseMsg:
		m.step = msg.Step
		m.stepMsg = msg.Label

	case status.LogMsg:
		if msg.Text != "" {
			m.logs = append(m.logs, msg.Text)
			if len(m.logs) > 8 { // Keep a small, readable log window
				m.logs = m.logs[1:]
			}
		}

	case status.ProgressMsg:
		if msg.Total > 0 {
			cmd = m.progress.SetPercent(float64(msg.Current) / float64(msg.Total))
		}

	case status.ConfirmMsg:
		m.planSummary = msg.Summary
		m.step = status.Confirming

	case status.ErrMsg:
		m.err = msg.Error
		m.step = status.Error
	}

	// Handle termination states
	if m.step == status.Done || m.step == status.Error {
		m.quitting = true
		return m, tea.Sequence(cmd, tea.Quit)
	}

	// If we're waiting for user input, stop asking for more progress events
	// until the user responds (which will either quit or move to Executing)
	if m.awaitingConfirm() {
		return m, cmd
	}

	// Otherwise, keep the event loop alive
	return m, tea.Batch(cmd, m.bus.WaitForProgress())
}

func (m *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	case status.Event:
		return m.handleEvent(msg)
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
		return fmt.Sprintf("\n %s\n\n", statusStyle.Render(fmt.Sprintf("✘ Error: %v", m.err)))
	}

	var s strings.Builder
	s.WriteString(m.renderHeader())
	s.WriteString(m.renderMainContent())
	s.WriteString(m.renderLogs())
	s.WriteString(m.renderFooter())

	return s.String()
}

func (m *TUI) renderHeader() string {
	return fmt.Sprintf("\n%s %s\n\n",
		titleStyle.Render(" ANKI SYNC "),
		m.stepMsg,
	)
}

func (m *TUI) renderMainContent() string {
	switch m.step {
	case status.Validating, status.Fetching:
		return "  " + m.spinner.View() + " Processing...\n\n"

	case status.Planning:
		return "  " + m.spinner.View() + " Processing...\n\n"

	case status.Confirming:
		return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Render("  ? Confirm sync? [y/N]") +
			"\n\n" + renderPlanSummary(m.planSummary) + "\n\n"

	case status.Executing:
		return "  " + m.progress.View() + "\n\n"
	}
	return ""
}

func (m *TUI) renderLogs() string {
	if m.awaitingConfirm() || len(m.logs) == 0 {
		return ""
	}

	var s strings.Builder
	for _, log := range m.logs {
		s.WriteString(logStyle.Render("  • "+log) + "\n")
	}
	s.WriteString("\n")
	return s.String()
}

func (m *TUI) renderFooter() string {
	if m.quitting {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("  ✔ Done!\n")
	}
	return logStyle.Render("  press q to quit\n")
}

func renderPlanSummary(summary string) string {
	var out strings.Builder
	for _, line := range strings.Split(summary, "\n") {
		switch {
		case strings.HasPrefix(line, "Adding"):
			out.WriteString("  " + planAddStyle.Render(line) + "\n")
		case strings.HasPrefix(line, "Updating"):
			out.WriteString("  " + planUpdateStyle.Render(line) + "\n")
		case line == "":
			out.WriteString("\n")
		case strings.Contains(line, "···"):
			out.WriteString(logStyle.Render(line) + "\n")
		default:
			out.WriteString(planItemStyle.Render(line) + "\n")
		}
	}
	return strings.TrimRight(out.String(), "\n")
}

func (m *TUI) awaitingConfirm() bool {
	return m.step == status.Confirming
}

func (m *TUI) Report(msg status.Event) {
	m.bus.SendProgress(msg)
}
