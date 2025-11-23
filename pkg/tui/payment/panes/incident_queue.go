package panes

import (
	"fmt"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// TicketStatus represents the status of a Jira ticket
type TicketStatus string

const (
	StatusOpen         TicketStatus = "Open"
	StatusInProgress   TicketStatus = "In Progress"
	StatusToDo         TicketStatus = "ToDo"
	StatusInReview     TicketStatus = "In Review"
	StatusDone         TicketStatus = "Done"
	StatusBlocked      TicketStatus = "Blocked"
)

// TicketPriority represents the priority level of a ticket
type TicketPriority string

const (
	PriorityCritical TicketPriority = "CRITICAL"
	PriorityHigh     TicketPriority = "HIGH"
	PriorityMedium   TicketPriority = "MED"
	PriorityLow      TicketPriority = "LOW"
)

// Incident represents a Jira incident ticket
type Incident struct {
	ID          string
	Title       string
	Status      TicketStatus
	Priority    TicketPriority
	Assignee    string
	Created     time.Time
	Updated     time.Time
	Description string
}

// IncidentQueuePane manages the incident queue pane
type IncidentQueuePane struct {
	Incidents   []Incident
	Selected    int
	ShowDetails bool
}

// NewIncidentQueuePane creates a new incident queue pane
func NewIncidentQueuePane() *IncidentQueuePane {
	return &IncidentQueuePane{
		Incidents:   generateMockIncidents(),
		Selected:    0,
		ShowDetails: false,
	}
}

// ToggleDetails toggles the details view
func (p *IncidentQueuePane) ToggleDetails() {
	p.ShowDetails = !p.ShowDetails
}

// generateMockIncidents creates mock incident data
func generateMockIncidents() []Incident {
	now := time.Now()
	return []Incident{
		{
			ID:          "PAY-1024",
			Title:       "Fix double charge race condition",
			Status:      StatusInProgress,
			Priority:    PriorityCritical,
			Assignee:    "oncall-payment",
			Created:     now.Add(-2 * time.Hour),
			Updated:     now.Add(-30 * time.Minute),
			Description: "Multiple customers reporting duplicate charges for single transactions",
		},
		{
			ID:          "PAY-1025",
			Title:       "Update mTLS certs for Provider X",
			Status:      StatusOpen,
			Priority:    PriorityHigh,
			Assignee:    "oncall-payment",
			Created:     now.Add(-4 * time.Hour),
			Updated:     now.Add(-1 * time.Hour),
			Description: "mTLS certificates expiring in 7 days for external payment provider",
		},
		{
			ID:          "PAY-1030",
			Title:       "Settlement delay investigation",
			Status:      StatusToDo,
			Priority:    PriorityMedium,
			Assignee:    "oncall-payment",
			Created:     now.Add(-6 * time.Hour),
			Updated:     now.Add(-3 * time.Hour),
			Description: "Settlement batches taking 2x longer than normal processing time",
		},
	}
}

// MoveSelection moves the selection up or down
func (p *IncidentQueuePane) MoveSelection(direction int) {
	switch direction {
	case -1: // Up
		if p.Selected > 0 {
			p.Selected--
		}
	case 1: // Down
		if p.Selected < len(p.Incidents)-1 {
			p.Selected++
		}
	}
}

// GetSelected returns the currently selected incident
func (p *IncidentQueuePane) GetSelected() *Incident {
	if p.Selected >= 0 && p.Selected < len(p.Incidents) {
		return &p.Incidents[p.Selected]
	}
	return nil
}

// Render renders the pane content
func (p *IncidentQueuePane) Render() string {
	if p.ShowDetails {
		return p.renderDetails()
	}
	return p.renderList()
}

func (p *IncidentQueuePane) renderList() string {
	if len(p.Incidents) == 0 {
		return "No active incidents."
	}

	var lines []string
	for i, incident := range p.Incidents {
		// Priority color
		priorityStyle := p.getPriorityStyle(incident.Priority)

		// Selection indicator
		cursor := " "
		style := p.getItemStyle()
		if i == p.Selected {
			cursor = ">"
			style = p.getSelectedItemStyle()
		}

		// Format: [PRIORITY] ID: Title (Time)
		line := fmt.Sprintf("%s %s %s: %s (%s)",
			cursor,
			priorityStyle.Render(fmt.Sprintf("[%s]", incident.Priority)),
			style.Render(incident.ID),
			p.truncateTitle(incident.Title, 30),
			incident.Created.Format("15:04"),
		)
		lines = append(lines, line)
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func (p *IncidentQueuePane) renderDetails() string {
	incident := p.GetSelected()
	if incident == nil {
		return "No incident selected."
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}).Background(lipgloss.AdaptiveColor{Light: "#DDDDDD", Dark: "#444444"}).Padding(0, 1)
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#666666", Dark: "#AAAAAA"})
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"})

	priorityStyle := p.getPriorityStyle(incident.Priority)

	content := []string{
		titleStyle.Render(fmt.Sprintf("%s: %s", incident.ID, incident.Title)),
		"",
		fmt.Sprintf("%s %s", labelStyle.Render("Status:"), valueStyle.Render(string(incident.Status))),
		fmt.Sprintf("%s %s", labelStyle.Render("Priority:"), priorityStyle.Render(string(incident.Priority))),
		fmt.Sprintf("%s %s", labelStyle.Render("Assignee:"), valueStyle.Render(incident.Assignee)),
		fmt.Sprintf("%s %s", labelStyle.Render("Created:"), valueStyle.Render(incident.Created.Format("2006-01-02 15:04:05"))),
		"",
		labelStyle.Render("Description:"),
		valueStyle.Render("This is a mock description for the incident. It would contain details from Jira."),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888888", Dark: "#808080"}).Render("[Enter] Back to List"),
	}

	return lipgloss.JoinVertical(lipgloss.Left, content...)
}

// getStatusIcon returns an icon for the ticket status
func (p *IncidentQueuePane) getStatusIcon(status TicketStatus) string {
	switch status {
	case StatusOpen:
		return "○"
	case StatusInProgress:
		return "●"
	case StatusToDo:
		return "◇"
	case StatusInReview:
		return "◐"
	case StatusDone:
		return "✓"
	case StatusBlocked:
		return "✗"
	default:
		return "○"
	}
}

// getPriorityStyle returns the appropriate style for the priority
func (p *IncidentQueuePane) getPriorityStyle(priority TicketPriority) lipgloss.Style {
	switch priority {
	case PriorityCritical:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC0000", Dark: "#FF0000"}).Bold(true)
	case PriorityHigh:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#CC8800", Dark: "#FFA500"}).Bold(true)
	case PriorityMedium:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#888800", Dark: "#FFFF00"}).Bold(true)
	case PriorityLow:
		return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF00"}).Bold(true)
	default:
		return lipgloss.NewStyle()
	}
}

// truncateTitle truncates a title to the specified length
func (p *IncidentQueuePane) truncateTitle(title string, maxLen int) string {
	if len(title) <= maxLen {
		return title
	}
	return title[:maxLen-3] + "..."
}

func (p *IncidentQueuePane) getSelectedItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#008800", Dark: "#00FF00"})
}

func (p *IncidentQueuePane) getItemStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#F0F0F0"})
}
