package panes

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ActionMode represents the active mode in the Actions pane
type ActionMode int

const (
	ModeMenu ActionMode = iota
	ModeTrace
	ModeDeregister
)

// TransactionStatus represents the status of a transaction step
type TransactionStatus int

const (
	StatusPending TransactionStatus = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusTimeout
)

// TransactionStep represents a single step in the transaction flow
type TransactionStep struct {
	Name      string
	System    string
	Status    TransactionStatus
	Duration  time.Duration
	Timestamp *time.Time
}

// TransactionFlow represents the full trace of a transaction
type TransactionFlow struct {
	TransactionID string
	CustomerID    string
	TotalAmount   string
	Currency      string
	Steps         []TransactionStep
}

// ActionsPane combines Diagnostics and Admin actions
type ActionsPane struct {
	Mode            ActionMode
	MenuSelected    int

	// Trace Transaction State
	TraceInput      string
	TraceCursor     int
	TraceResult     *TransactionFlow

	// Deregister PayNow State
	DeregisterInput string
	DeregisterCursor int
	DeregisterSubmitted bool

	InputFocused    bool
}

// NewActionsPane creates a new actions pane
func NewActionsPane() *ActionsPane {
	return &ActionsPane{
		Mode:            ModeMenu,
		MenuSelected:    0,
		TraceInput:      "",
		TraceCursor:     0,
		TraceResult:     nil,
		DeregisterInput: "",
		DeregisterCursor: 0,
		DeregisterSubmitted: false,
		InputFocused:    false,
	}
}

// SelectAction selects the current menu item
func (p *ActionsPane) SelectAction() {
	if p.Mode == ModeMenu {
		if p.MenuSelected == 0 {
			p.Mode = ModeTrace
		} else {
			p.Mode = ModeDeregister
		}
		p.InputFocused = true // Auto-focus input when entering mode
	}
}

// BackToMenu returns to the menu
func (p *ActionsPane) BackToMenu() {
	p.Mode = ModeMenu
	p.InputFocused = false
	p.ResetState()
}

// MoveMenuSelection moves the menu selection
func (p *ActionsPane) MoveMenuSelection(direction int) {
	if p.Mode != ModeMenu {
		return
	}
	p.MenuSelected += direction
	if p.MenuSelected < 0 {
		p.MenuSelected = 0
	}
	if p.MenuSelected > 1 { // 2 items
		p.MenuSelected = 1
	}
}

// FocusInput focuses the input field for the current mode
func (p *ActionsPane) FocusInput() {
	if p.Mode != ModeMenu {
		p.InputFocused = true
	}
}

// UnfocusInput unfocuses the input field
func (p *ActionsPane) UnfocusInput() {
	p.InputFocused = false
}

// HandleInput handles text input for the active mode
func (p *ActionsPane) HandleInput(char string) {
	if !p.InputFocused || p.Mode == ModeMenu {
		return
	}

	var input *string
	var cursor *int
	maxLen := 30

	if p.Mode == ModeTrace {
		input = &p.TraceInput
		cursor = &p.TraceCursor
	} else {
		input = &p.DeregisterInput
		cursor = &p.DeregisterCursor
	}

	if char == "backspace" {
		if len(*input) > 0 && *cursor > 0 {
			*input = (*input)[:*cursor-1] + (*input)[*cursor:]
			*cursor--
		}
	} else if len(char) == 1 && *cursor < maxLen {
		*input = (*input)[:*cursor] + char + (*input)[*cursor:]
		*cursor++
	}
}

// MoveCursor moves the cursor in the input field
func (p *ActionsPane) MoveCursor(direction int) {
	if !p.InputFocused || p.Mode == ModeMenu {
		return
	}

	var inputLen int
	var cursor *int

	if p.Mode == ModeTrace {
		inputLen = len(p.TraceInput)
		cursor = &p.TraceCursor
	} else {
		inputLen = len(p.DeregisterInput)
		cursor = &p.DeregisterCursor
	}

	switch direction {
	case -1: // Left
		if *cursor > 0 {
			*cursor--
		}
	case 1: // Right
		if *cursor < inputLen {
			*cursor++
		}
	}
}

// ExecuteAction executes the action for the current mode
func (p *ActionsPane) ExecuteAction() {
	if p.Mode == ModeTrace {
		if len(p.TraceInput) > 0 {
			// Reuse the mock flow generation logic from diagnostics.go
			// For simplicity, I'll inline a basic version here or we could keep diagnostics.go as a helper
			// Let's inline a simple mock for now
			p.TraceResult = generateMockFlow(p.TraceInput)
		}
	} else if p.Mode == ModeDeregister {
		if len(p.DeregisterInput) > 0 {
			p.DeregisterSubmitted = true
		}
	}
}

// Render renders the actions pane content
func (p *ActionsPane) Render() string {
	if p.Mode == ModeMenu {
		return p.renderMenu()
	}

	// Header for active mode
	var title string
	if p.Mode == ModeTrace {
		title = "TRACE TRANSACTION"
	} else {
		title = "DEREGISTER PAYNOW"
	}

	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#444444"}).Padding(0, 1).Render(title)

	var content string
	if p.Mode == ModeTrace {
		content = p.renderTraceView()
	} else {
		content = p.renderDeregisterView()
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		content,
	)
}

func (p *ActionsPane) renderMenu() string {
	items := []string{"Trace Transaction", "Deregister PayNow"}
	var lines []string

	for i, item := range items {
		cursor := " "
		style := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"})
		if i == p.MenuSelected {
			cursor = ">"
			style = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF00"}).Bold(true)
		}

		lines = append(lines, fmt.Sprintf("%s %s", cursor, style.Render(item)))
	}

	help := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("\n[Enter] Select Action")

	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinVertical(lipgloss.Left, lines...),
		help,
	)
}

func (p *ActionsPane) renderTraceView() string {
	if p.TraceResult != nil {
		return p.renderTraceResult()
	}

	// Input View
	label := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"}).Render("Transaction ID:")

	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).Background(lipgloss.AdaptiveColor{Light: "#EEEEEE", Dark: "#333333"}).Padding(0, 1).Width(30)
	if p.InputFocused && p.Mode == ModeTrace {
		inputStyle = inputStyle.Background(lipgloss.AdaptiveColor{Light: "#CCFFCC", Dark: "#005500"})
	}

	inputText := p.TraceInput
	if inputText == "" && !p.InputFocused {
		inputText = "Enter ID..."
	}

	input := inputStyle.Render(inputText)

	help := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("[Enter] to Search | [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		input,
		"",
		help,
	)
}

func (p *ActionsPane) renderTraceResult() string {
	// Simplified result view
	header := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#0000AA", Dark: "#8888FF"}).Render(fmt.Sprintf("Flow: %s", p.TraceResult.TransactionID))

	var steps []string
	for _, step := range p.TraceResult.Steps {
		statusColor := lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF00"}
		if step.Status == StatusFailed {
			statusColor = lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF0000"}
		}

		line := fmt.Sprintf("• %s", lipgloss.NewStyle().Foreground(statusColor).Render(step.Name))
		steps = append(steps, line)
	}

	back := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("[Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		"",
		lipgloss.JoinVertical(lipgloss.Left, steps...),
		"",
		back,
	)
}

func (p *ActionsPane) renderDeregisterView() string {
	if p.DeregisterSubmitted {
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF00"}).Bold(true).Render("✓ Request Submitted"),
			"",
			fmt.Sprintf("Ticket created for: %s", p.DeregisterInput),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("[Esc] Back"),
		)
	}

	label := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"}).Render("Phone Number / Proxy ID:")

	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).Background(lipgloss.AdaptiveColor{Light: "#EEEEEE", Dark: "#333333"}).Padding(0, 1).Width(30)
	if p.InputFocused && p.Mode == ModeDeregister {
		inputStyle = inputStyle.Background(lipgloss.AdaptiveColor{Light: "#CCFFCC", Dark: "#005500"})
	}

	inputText := p.DeregisterInput
	if inputText == "" && !p.InputFocused {
		inputText = "+65..."
	}

	input := inputStyle.Render(inputText)

	help := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("[Enter] to Submit | [Esc] Back")

	return lipgloss.JoinVertical(lipgloss.Left,
		label,
		input,
		"",
		help,
	)
}

// Helper to clear state
func (p *ActionsPane) ResetState() {
	p.TraceResult = nil
	p.DeregisterSubmitted = false
	p.InputFocused = false
	p.TraceInput = ""
	p.DeregisterInput = ""
}

// Reusing mock data generation
func generateMockFlow(id string) *TransactionFlow {
	now := time.Now()
	return &TransactionFlow{
		TransactionID: id,
		Steps: []TransactionStep{
			{Name: "Payment Initiated", Status: StatusCompleted, Timestamp: &now},
			{Name: "Risk Check", Status: StatusCompleted, Timestamp: &now},
			{Name: "Ledger Update", Status: StatusFailed, Timestamp: &now},
		},
	}
}
