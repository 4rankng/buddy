package main

import (
	"fmt"
	"os"

	"oncall/pkg/tui/payment"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Launch enhanced payment team dashboard
	p := tea.NewProgram(payment.NewEnhancedDashboardModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
