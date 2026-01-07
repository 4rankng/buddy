package doorman

import (
	"buddy/internal/clients/doorman"
	"buddy/internal/txn/domain"
	"fmt"

	"github.com/manifoldco/promptui"
	"strings"
)

// PromptForDoormanTicket prompts user to create Doorman DML tickets for all services
// This function is shared between mybuddy and sgbuddy to avoid circular dependencies
// If autoCreate is true and note is provided, skips prompts and creates tickets automatically
func PromptForDoormanTicket(doormanClient doorman.DoormanInterface, statements domain.SQLStatements, autoCreate bool, note string) {
	if doormanClient == nil {
		return
	}

	ProcessServiceDML(doormanClient, "payment_core", statements.PCDeployStatements, statements.PCRollbackStatements, autoCreate, note)
	ProcessServiceDML(doormanClient, "rpp_adapter", statements.RPPDeployStatements, statements.RPPRollbackStatements, autoCreate, note)
	ProcessServiceDML(doormanClient, "payment_engine", statements.PEDeployStatements, statements.PERollbackStatements, autoCreate, note)
	ProcessServiceDML(doormanClient, "partnerpay_engine", statements.PPEDeployStatements, statements.PPERollbackStatements, autoCreate, note)
}

// ProcessServiceDML prompts and creates a Doorman ticket for a single service
// If autoCreate is true and note is provided, skips prompts and creates ticket automatically
func ProcessServiceDML(doormanClient doorman.DoormanInterface, serviceName string, deployStmts, rollbackStmts []string, autoCreate bool, note string) {
	if len(deployStmts) == 0 {
		return
	}

	// Auto-create mode: skip prompts and create ticket directly
	if autoCreate {
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
		return
	}

	// Interactive mode: prompt user
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
