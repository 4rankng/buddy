package sgbuddy

import (
	"buddy/internal/apps/common"
	"github.com/spf13/cobra"
)

func GetCommands(appCtx *common.Context) []*cobra.Command {
	return []*cobra.Command{
		NewTxnCmd(appCtx),
		NewJiraCmd(appCtx),
		NewEcoTxnCmd(appCtx),
	}
}
