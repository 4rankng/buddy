package cobra

import (
	"buddy/internal/apps/common"
	"fmt"
	"github.com/spf13/cobra"
)

func NewRootCommand(appCtx *common.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:               appCtx.BinaryName,
		Short:             "Oncall CLI tool for G-Bank operations",
		Long:              `A CLI tool for oncall operations at G-Bank, including transaction queries and other operational tasks.`,
		DisableAutoGenTag: true,
		Version:           "1.0.0",
	}

	// Add version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Display the version of " + appCtx.BinaryName,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s version 1.0.0 (Environment: %s)\n", appCtx.BinaryName, appCtx.GetPrefix()[1:3])
		},
	})

	return rootCmd
}
