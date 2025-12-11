package cmd

import (
	"buddy/internal/app"
	"github.com/spf13/cobra"
)

func NewRppCmd(appCtx *app.Context) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rpp",
		Short: "RPP (Real-time Payment Processing) specific commands",
		Long: `Commands for handling RPP-specific operations.
Useful for managing workflows in the RPP adapter system.`,
	}

	// Add subcommands
	cmd.AddCommand(NewRppResumeCmd(appCtx))

	return cmd
}
