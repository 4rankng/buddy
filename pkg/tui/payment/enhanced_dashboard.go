package payment

import (
	"fmt"
	"oncall/pkg/core"
	"oncall/pkg/tui/payment/panes"
	"oncall/pkg/services/payment"
	"oncall/pkg/ports"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EnhancedDashboardModel represents the payment team dashboard
type EnhancedDashboardModel struct {
	Focus           FocusManager
	Container       *core.Container
	PaymentService  *payment.PaymentService
	IncidentPane    *panes.IncidentQueuePane
	ActionsPane     *panes.ActionsPane
	LogsPane        *panes.LogsPane
	Width           int
	Height          int
	Quitting        bool
	Initialized     bool
	LastError       string
}

// NewEnhancedDashboardModel creates a new enhanced dashboard model
func NewEnhancedDashboardModel() (EnhancedDashboardModel, error) {
	// Initialize the dependency container
	container, err := core.NewContainer()
	if err != nil {
		return EnhancedDashboardModel{}, fmt.Errorf("failed to initialize container: %w", err)
	}

	// Get payment service from container
	paymentService := container.PaymentService()

	return EnhancedDashboardModel{
		Focus:          NewFocusManager(),
		Container:      container,
		PaymentService: paymentService,
		IncidentPane:   panes.NewIncidentQueuePane(),
		ActionsPane:    panes.NewActionsPane(),
		LogsPane:       panes.NewLogsPane(5), // Show 5 log lines
		Width:          80,
		Height:         24,
		Quitting:       false,
		Initialized:    false,
		LastError:      "",
	}, nil
}

func (m EnhancedDashboardModel) Init() tea.Cmd {
	// Load initial data when dashboard starts
	return tea.Batch(
		m.loadInitialData(),
	)
}

func (m EnhancedDashboardModel) loadInitialData() tea.Cmd {
	return func() tea.Msg {
		// Load incidents
		if m.PaymentService != nil {
			tickets, err := m.PaymentService.GetAssignedTickets()
			if err != nil {
				m.LogsPane.AddLog(panes.LogLevelError, fmt.Sprintf("Failed to load tickets: %v", err))
			} else {
				m.updateIncidentPane(tickets)
				m.LogsPane.AddLog(panes.LogLevelInfo, fmt.Sprintf("Loaded %d tickets", len(tickets)))
			}

			// Load stuck transactions
			stuckTxns, err := m.PaymentService.GetStuckTransactions(24) // Last 24 hours
			if err != nil {
				m.LogsPane.AddLog(panes.LogLevelError, fmt.Sprintf("Failed to load stuck transactions: %v", err))
			} else {
				m.LogsPane.AddLog(panes.LogLevelInfo, fmt.Sprintf("Found %d stuck transactions", len(stuckTxns)))
			}
		}

		// Add initial log entries
		m.LogsPane.AddLog(panes.LogLevelInfo, "Payment dashboard initialized successfully")
		m.LogsPane.AddLog(panes.LogLevelSuccess, "Connected to Jira and Doorman services")

		return nil
	}
}

// updateIncidentPane updates the incident pane with Jira tickets
func (m *EnhancedDashboardModel) updateIncidentPane(tickets []ports.JiraTicket) {
	incidents := make([]panes.Incident, 0, len(tickets))

	for _, ticket := range tickets {
		priority := panes.PriorityMedium
		switch ticket.Priority {
		case "Critical", "Highest":
			priority = panes.PriorityCritical
		case "High":
			priority = panes.PriorityHigh
		case "Low", "Lowest":
			priority = panes.PriorityLow
		}

		status := panes.StatusOpen
		switch ticket.Status {
		case "In Progress", "Reopened":
			status = panes.StatusInProgress
		case "To Do":
			status = panes.StatusToDo
		case "In Review":
			status = panes.StatusInReview
		case "Done", "Closed", "Resolved":
			status = panes.StatusDone
		case "Blocked":
			status = panes.StatusBlocked
		}

		incident := panes.Incident{
			ID:          ticket.Key,
			Title:       ticket.Summary,
			Status:      status,
			Priority:    priority,
			Assignee:    ticket.Assignee,
			Created:     ticket.Created,
			Updated:     ticket.Updated,
			Description: ticket.Description,
		}
		incidents = append(incidents, incident)
	}

	m.IncidentPane.Incidents = incidents
}

// executeTraceAction executes transaction tracing
func (m *EnhancedDashboardModel) executeTraceAction(transactionID string) {
	if m.PaymentService == nil {
		m.LogsPane.AddLog(panes.LogLevelError, "Payment service not available")
		return
	}

	m.LogsPane.AddLog(panes.LogLevelInfo, fmt.Sprintf("Tracing transaction: %s", transactionID))

	flow, err := m.PaymentService.TraceTransaction(transactionID)
	if err != nil {
		m.LogsPane.AddLog(panes.LogLevelError, fmt.Sprintf("Failed to trace transaction: %v", err))
		m.ActionsPane.TraceResult = nil
		return
	}

	// Convert to display format
	steps := make([]panes.TransactionStep, 0, len(flow.Steps))
	for _, step := range flow.Steps {
		status := panes.StatusPending
		switch step.Status {
		case payment.StatusCompleted:
			status = panes.StatusCompleted
		case payment.StatusProcessing:
			status = panes.StatusProcessing
		case payment.StatusFailed:
			status = panes.StatusFailed
		case payment.StatusTimeout:
			status = panes.StatusTimeout
		}

		transactionStep := panes.TransactionStep{
			Name:      step.Name,
			System:    step.System,
			Status:    status,
			Duration:  step.Duration,
			Timestamp: step.Timestamp,
		}
		steps = append(steps, transactionStep)
	}

	m.ActionsPane.TraceResult = &panes.TransactionFlow{
		TransactionID: flow.TransactionID,
		CustomerID:    flow.CustomerID,
		TotalAmount:   flow.TotalAmount,
		Currency:      flow.Currency,
		Steps:         steps,
	}

	m.LogsPane.AddLog(panes.LogLevelSuccess, fmt.Sprintf("Transaction traced: %s steps completed", len(steps)))
}

// executeDeregisterAction executes PayNow deregistration
func (m *EnhancedDashboardModel) executeDeregisterAction(input string) {
	if m.PaymentService == nil {
		m.LogsPane.AddLog(panes.LogLevelError, "Payment service not available")
		return
	}

	// Split input to handle multiple accounts
	accounts := []string{input}
	if len(input) > 0 && input[0] == '+' {
		// Assume phone number format
		accounts = []string{input}
	}

	m.LogsPane.AddLog(panes.LogLevelInfo, fmt.Sprintf("Creating PayNow deregistration request for %d account(s)", len(accounts)))

	result, err := m.PaymentService.CreateDeregistrationSHIPRM(accounts)
	if err != nil {
		m.LogsPane.AddLog(panes.LogLevelError, fmt.Sprintf("Failed to create deregistration request: %v", err))
		return
	}

	m.LogsPane.AddLog(panes.LogLevelSuccess, fmt.Sprintf("Created SHIPRM ticket %s for deregistration", result.TicketID))
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
		// Execute action based on current mode
		if m.ActionsPane.Mode == panes.ModeTrace && len(m.ActionsPane.TraceInput) > 0 {
			m.executeTraceAction(m.ActionsPane.TraceInput)
		} else if m.ActionsPane.Mode == panes.ModeDeregister && len(m.ActionsPane.DeregisterInput) > 0 {
			m.executeDeregisterAction(m.ActionsPane.DeregisterInput)
			m.ActionsPane.DeregisterSubmitted = true
		}

		m.ActionsPane.UnfocusInput()
		m.Focus.InputMode = false
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
