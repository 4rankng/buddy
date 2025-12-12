package sgbuddy

import (
	"buddy/internal/app"
	"github.com/spf13/cobra"
)

func GetCommands(appCtx *app.Context) []*cobra.Command {
	return []*cobra.Command{
		NewTxnCmd(appCtx),
	}
}
