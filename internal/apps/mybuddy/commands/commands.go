package mybuddy

import (
	"buddy/internal/apps/common"
	"buddy/internal/di"

	"github.com/spf13/cobra"
)

func GetCommands(appCtx *common.Context, clients *di.ClientSet) []*cobra.Command {
	return []*cobra.Command{
		NewTxnCmd(appCtx, clients),
		NewRppCmd(appCtx, clients),
		NewRppResumeCmd(appCtx, clients),
		NewEcoTxnCmd(appCtx, clients),
		NewJiraCmd(appCtx, clients),
	}
}
