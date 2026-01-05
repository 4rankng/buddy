package sgbuddy

import (
	"buddy/internal/apps/common"
	"buddy/internal/di"

	"github.com/spf13/cobra"
)

func GetCommands(appCtx *common.Context, clients *di.ClientSet) []*cobra.Command {
	return []*cobra.Command{
		NewTxnCmd(appCtx, clients),
		NewJiraCmd(appCtx, clients),
		NewEcoTxnCmd(appCtx, clients),
		NewPayNowCmd(appCtx, clients),
		NewDatadogCmd(appCtx, clients),
		NewDoormanCmd(appCtx, clients),
	}
}
