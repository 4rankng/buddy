package main

import (
	"fmt"
	"os"

	"oncall/pkg/tui/payment"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Launch enhanced payment team dashboard
	model, err := payment.NewEnhancedDashboardModel()
	if err != nil {
		fmt.Printf("Failed to initialize dashboard: %v", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
