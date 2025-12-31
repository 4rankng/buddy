package mybuddy

import (
	"buddy/internal/apps/common"
	"github.com/spf13/cobra"
)

func NewRppCmd(appCtx *common.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpp",
		Short: "RPP (Real-time Payment Processing) specific commands",
		Long: `Commands for handling RPP-specific operations.
Useful for managing workflows in the RPP adapter system.`,
	}

	// Add subcommands
	cmd.AddCommand(NewRppResumeCmd(appCtx))
	cmd.AddCommand(NewRppRtpCashinCmd(appCtx))

	return cmd
}
