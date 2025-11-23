package payment

// PaneType represents different panes in the dashboard
type PaneType int

const (
	IncidentPane PaneType = iota
	ActionsPane
	LogPane
)

// String returns the string representation of a PaneType
func (p PaneType) String() string {
	switch p {
	case IncidentPane:
		return "Incident Queue"
	case ActionsPane:
		return "Actions"
	case LogPane:
		return "Logs"
	default:
		return "Unknown"
	}
}

// FocusManager manages focus state across panes
type FocusManager struct {
	ActivePane  PaneType
	IncidentIdx int
	LogScroll   int
	InputMode   bool
	ModalActive bool
}

// NewFocusManager creates a new focus manager
func NewFocusManager() FocusManager {
	return FocusManager{
		ActivePane:  IncidentPane,
		IncidentIdx: 0,
		LogScroll:   0,
		InputMode:   false,
		ModalActive: false,
	}
}

// NavigateNextPane moves focus to the next pane
func (f *FocusManager) NavigateNextPane() {
	if f.InputMode || f.ModalActive {
		return
	}

	switch f.ActivePane {
	case IncidentPane:
		f.ActivePane = ActionsPane
	case ActionsPane:
		f.ActivePane = IncidentPane
	case LogPane:
		f.ActivePane = IncidentPane
	}
}

// NavigatePrevPane moves focus to the previous pane
func (f *FocusManager) NavigatePrevPane() {
	if f.InputMode || f.ModalActive {
		return
	}

	switch f.ActivePane {
	case IncidentPane:
		f.ActivePane = ActionsPane
	case ActionsPane:
		f.ActivePane = IncidentPane
	case LogPane:
		f.ActivePane = ActionsPane
	}
}

// GetHelpText returns appropriate help text based on current focus
func (f *FocusManager) GetHelpText() string {
	if f.InputMode {
		return "[Enter] Submit | [Esc] Exit Input Mode"
	}

	switch f.ActivePane {
	case IncidentPane:
		return "[↑/↓] Navigate | [Enter] Open Details | [Tab] Switch Pane"
	case ActionsPane:
		return "[Tab] Switch Pane | [Space] Switch Action | [Enter] Focus Input"
	case LogPane:
		return "[↑/↓] Scroll | [Tab] Switch Pane"
	default:
		return "[Tab] Switch Pane | [Esc] Exit | [Q] Quit"
	}
}
