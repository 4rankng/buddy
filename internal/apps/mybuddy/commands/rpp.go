package mybuddy

import (
	"buddy/internal/apps/common"
	"buddy/internal/di"

	"github.com/spf13/cobra"
)

func NewRppCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpp",
		Short: "RPP (Real-time Payment Processing) specific commands",
		Long: `Commands for handling RPP-specific operations.
Useful for managing workflows in the RPP adapter system.`,
	}

	// Add subcommands
	cmd.AddCommand(NewRppResumeCmd(appCtx, clients))
	cmd.AddCommand(NewRppRtpCashinCmd(appCtx, clients))

	return cmd
}
