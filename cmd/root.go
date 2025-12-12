package cmd

import (
	"buddy/cmd/mybuddy"
	"buddy/cmd/sgbuddy"
	"buddy/internal/app"
	"fmt"
	"github.com/spf13/cobra"
)

func Execute(appCtx *app.Context) error {
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

	// Add all subcommands
	rootCmd.AddCommand(mybuddy.NewTxnCmd(appCtx))
	rootCmd.AddCommand(mybuddy.NewRppCmd(appCtx))
	rootCmd.AddCommand(mybuddy.NewRppResumeCmd(appCtx))
	rootCmd.AddCommand(mybuddy.NewEcoTxnCmd(appCtx))

	// Add SG-specific commands only for sgbuddy
	if appCtx.IsSG() {
		rootCmd.AddCommand(sgbuddy.NewTxnCmd(appCtx))
	}

	return rootCmd.Execute()
}
