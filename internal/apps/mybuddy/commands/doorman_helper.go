package mybuddy

import (
	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/txn/domain"

	commondoorman "buddy/internal/apps/common/doorman"
)

// PromptForDoormanTicket prompts user to create Doorman DML tickets for all services
// Delegates to the common implementation to avoid code duplication
func PromptForDoormanTicket(appCtx *common.Context, clients *di.ClientSet, statements domain.SQLStatements, autoCreate bool, note string) {
	commondoorman.PromptForDoormanTicket(clients.Doorman, statements, autoCreate, note)
}
