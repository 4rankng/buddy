package payment

import (
	"fmt"
	"oncall/pkg/tui/payment/panes"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EnhancedDashboardModel represents the payment team dashboard
type EnhancedDashboardModel struct {
	Focus           FocusManager
	IncidentPane    *panes.IncidentQueuePane
	ActionsPane     *panes.ActionsPane
	LogsPane        *panes.LogsPane
	Width           int
	Height          int
	Quitting        bool
}

// NewEnhancedDashboardModel creates a new enhanced dashboard model
func NewEnhancedDashboardModel() EnhancedDashboardModel {
	return EnhancedDashboardModel{
		Focus:           NewFocusManager(),
		IncidentPane:    panes.NewIncidentQueuePane(),
		ActionsPane:     panes.NewActionsPane(),
		LogsPane:        panes.NewLogsPane(5), // Show 5 log lines
		Width:           80,
		Height:          24,
		Quitting:        false,
	}
}

func (m EnhancedDashboardModel) Init() tea.Cmd {
	return nil
}

func (m EnhancedDashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		m.handleWindowSize(msg)
	}

	return m, nil
}

func (m EnhancedDashboardModel) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle input mode for actions
	if m.Focus.InputMode {
		return m.handleInputMode(msg)
	}

	// Normal navigation mode
	switch msg.Type {
	case tea.KeyTab:
		m.Focus.NavigateNextPane()
	case tea.KeyShiftTab:
		m.Focus.NavigatePrevPane()
	case tea.KeyEscape:
		if m.Focus.ActivePane == ActionsPane {
			if m.ActionsPane.Mode != panes.ModeMenu {
				m.ActionsPane.BackToMenu()
			}
		} else if m.Focus.ActivePane == IncidentPane {
			if m.IncidentPane.ShowDetails {
				m.IncidentPane.ToggleDetails()
			}
		}
		m.Focus.InputMode = false
	case tea.KeyUp:
		m.handleUpNavigation()
	case tea.KeyDown:
		m.handleDownNavigation()
	case tea.KeyEnter:
		return m.handleEnterKey()
	case tea.KeyCtrlC:
		m.Quitting = true
		return m, tea.Quit
	}

	// Handle specific character keys
	switch msg.String() {
	case "q":
		if !m.Focus.InputMode {
			m.Quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m EnhancedDashboardModel) handleInputMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.ActionsPane.UnfocusInput()
		m.ActionsPane.BackToMenu()
		m.Focus.InputMode = false
	case tea.KeyEnter:
		// Execute action
		m.ActionsPane.ExecuteAction()
		m.ActionsPane.UnfocusInput()
		m.Focus.InputMode = false

		// Add log entry
		modeStr := "Trace"
		if m.ActionsPane.Mode == panes.ModeDeregister {
			modeStr = "Deregister"
		}
		m.LogsPane.AddLog(panes.LogLevelInfo, fmt.Sprintf("Action executed: %s", modeStr))
	case tea.KeyBackspace:
		m.ActionsPane.HandleInput("backspace")
	case tea.KeyLeft:
		m.ActionsPane.MoveCursor(-1)
	case tea.KeyRight:
		m.ActionsPane.MoveCursor(1)
	default:
		// Handle text input
		if len(msg.String()) == 1 {
			m.ActionsPane.HandleInput(msg.String())
		}
	}

	return m, nil
}

func (m *EnhancedDashboardModel) handleUpNavigation() {
	switch m.Focus.ActivePane {
	case IncidentPane:
		m.IncidentPane.MoveSelection(-1)
		m.Focus.IncidentIdx = m.IncidentPane.Selected
	case ActionsPane:
		m.ActionsPane.MoveMenuSelection(-1)
	case LogPane:
		m.LogsPane.ScrollUp()
		m.Focus.LogScroll = m.LogsPane.Scroll
	}
}

func (m *EnhancedDashboardModel) handleDownNavigation() {
	switch m.Focus.ActivePane {
	case IncidentPane:
		m.IncidentPane.MoveSelection(1)
		m.Focus.IncidentIdx = m.IncidentPane.Selected
	case ActionsPane:
		m.ActionsPane.MoveMenuSelection(1)
	case LogPane:
		m.LogsPane.ScrollDown()
		m.Focus.LogScroll = m.LogsPane.Scroll
	}
}

func (m *EnhancedDashboardModel) handleEnterKey() (tea.Model, tea.Cmd) {
	switch m.Focus.ActivePane {
	case IncidentPane:
		// Toggle incident details
		m.IncidentPane.ToggleDetails()
		if m.IncidentPane.ShowDetails {
			incident := m.IncidentPane.GetSelected()
			if incident != nil {
				m.LogsPane.AddLog(panes.LogLevelInfo,
					fmt.Sprintf("Viewing details for: %s", incident.ID))
			}
		}
	case ActionsPane:
		if m.ActionsPane.Mode == panes.ModeMenu {
			m.ActionsPane.SelectAction()
			m.Focus.InputMode = true // Auto-enter input mode
		} else {
			// Focus input if not in menu
			m.ActionsPane.FocusInput()
			m.Focus.InputMode = true
		}
	case LogPane:
		// Scroll to bottom
		m.LogsPane.ScrollToBottom()
	}

	return m, nil
}

func (m *EnhancedDashboardModel) handleWindowSize(msg tea.WindowSizeMsg) {
	m.Width = msg.Width
	m.Height = msg.Height

	// Update logs pane max lines based on new layout
	m.LogsPane.MaxLines = 6 - 2 // Account for padding
}

func (m EnhancedDashboardModel) View() string {
	if m.Quitting {
		return ""
	}

	// Header
	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#FFFFFF"}).
		Background(lipgloss.AdaptiveColor{Light: "#0055AA", Dark: "#1E90FF"}).
		Width(m.Width).
		Height(1).
		Padding(0, 1).
		Align(lipgloss.Center).
		Render("ONCALL CLI  ▸  TEAM: PAYMENT  ▸  ENV: PRODUCTION")

	// Calculate pane sizes
	mainHeight := m.Height - 1 - 1 - 8 // Reserve space for header, footer, logs, and margins
	if mainHeight < 10 {
		mainHeight = 10 // Minimum height
	}
	halfWidth := m.Width/2 - 2

	// Styles for active/inactive panes
	activeBorder := lipgloss.AdaptiveColor{Light: "#00AA00", Dark: "#00FF00"} // Green for active
	inactiveBorder := lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#4A4A4A"} // Grey for inactive

	getPaneStyle := func(isActive bool) lipgloss.Style {
		borderColor := inactiveBorder
		if isActive {
			borderColor = activeBorder
		}
		return lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(borderColor).
			Padding(0, 1).
			Width(halfWidth).
			Height(mainHeight)
	}

	// Incident Pane
	incidentStyle := getPaneStyle(m.Focus.ActivePane == IncidentPane)
	incidentContent := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#AA8800", Dark: "#FFD700"}).Render("1. INCIDENT QUEUE"),
		lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#808080"}).Render("─────────────────"),
		m.IncidentPane.Render(),
	)
	leftPanel := incidentStyle.Render(incidentContent)

	// Actions Pane (Right Panel)
	actionsStyle := getPaneStyle(m.Focus.ActivePane == ActionsPane)
	actionsContent := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#AA8800", Dark: "#FFD700"}).Render("2. ACTIONS"),
		lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#AAAAAA", Dark: "#808080"}).Render("──────────"),
		m.ActionsPane.Render(),
	)
	rightPanel := actionsStyle.Render(actionsContent)

	// Main content area
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Logs Pane
	logStyle := getPaneStyle(m.Focus.ActivePane == LogPane).
		Width(m.Width - 2).
		Height(6).
		Border(lipgloss.RoundedBorder())

	visibleLogs := m.LogsPane.GetVisibleLogs()
	logTitle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#AA8800", Dark: "#FFD700"}).Render("LOGS & FEEDBACK")

	var logLines []string
	for _, log := range visibleLogs {
		logLine := fmt.Sprintf("[%s] %s: %s",
			log.Timestamp.Format("15:04:22"),
			log.Level,
			log.Message)
		logLines = append(logLines, logLine)
	}
	if len(logLines) == 0 {
		logLines = append(logLines, "No logs yet...")
	}

	logInner := lipgloss.JoinVertical(lipgloss.Left,
		logTitle,
		lipgloss.JoinVertical(lipgloss.Left, logLines...),
	)
	logPanel := logStyle.Render(logInner)

	// Footer / Help
	helpText := m.Focus.GetHelpText()
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#000000"}).
		Background(lipgloss.AdaptiveColor{Light: "#D0D0D0", Dark: "#A0A0A0"}).
		Width(m.Width).
		Padding(0, 1).
		Render(helpText)

	// Combine all
	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		mainContent,
		logPanel,
		footer,
	)
}
