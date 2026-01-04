package mybuddy

import (
	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/txn/domain"

	commondoorman "buddy/internal/apps/common/doorman"
)

// PromptForDoormanTicket prompts user to create Doorman DML tickets for all services
func PromptForDoormanTicket(appCtx *common.Context, clients *di.ClientSet, statements domain.SQLStatements) {
	if clients.Doorman == nil {
		return
	}

	commondoorman.ProcessServiceDML(clients.Doorman, "payment_core", statements.PCDeployStatements, statements.PCRollbackStatements)
	commondoorman.ProcessServiceDML(clients.Doorman, "rpp_adapter", statements.RPPDeployStatements, statements.RPPRollbackStatements)
	commondoorman.ProcessServiceDML(clients.Doorman, "payment_engine", statements.PEDeployStatements, statements.PERollbackStatements)
	commondoorman.ProcessServiceDML(clients.Doorman, "partnerpay_engine", statements.PPEDeployStatements, statements.PPERollbackStatements)
}
