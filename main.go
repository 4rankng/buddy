package main

import (
	"fmt"
	"os"

	"mybuddy/cmd"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:               "mybuddy",
	Short:             "Oncall CLI tool for G-Bank operations",
	Long:              `A CLI tool for oncall operations at G-Bank, including transaction queries and other operational tasks.`,
	DisableAutoGenTag: true,
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(cmd.TxnCmd)
}
