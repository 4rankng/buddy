package doorman

import (
	"buddy/internal/clients/doorman"
	"fmt"

	"github.com/manifoldco/promptui"
	"strings"
)

// ProcessServiceDML prompts and creates a Doorman ticket for a single service
func ProcessServiceDML(doormanClient doorman.DoormanInterface, serviceName string, deployStmts, rollbackStmts []string) {
	if len(deployStmts) == 0 {
		return
	}

	fmt.Println()
	prompt := promptui.Select{
		Label: fmt.Sprintf("Create Doorman DML ticket for %s?", serviceName),
		Items: []string{"Yes", "No"},
	}

	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	if result == "Yes" {
		promptNote := promptui.Prompt{
			Label: "Ticket Note",
		}
		note, err := promptNote.Run()
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		originalQuery := strings.Join(deployStmts, "\n")
		rollbackQuery := strings.Join(rollbackStmts, "\n")

		fmt.Printf("Creating ticket for %s...\n", serviceName)
		ticketID, err := doormanClient.CreateTicket(serviceName, originalQuery, rollbackQuery, note)
		if err != nil {
			fmt.Printf("Failed to create ticket: %v\n", err)
			return
		}

		ticketURL := fmt.Sprintf("https://doorman.infra.prd.g-bank.app/rds/dml/%s", ticketID)
		fmt.Printf("Ticket created successfully!\nTicket ID: %s\nTicket URL: %s\n", ticketID, ticketURL)
	}
}
