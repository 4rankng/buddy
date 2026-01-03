package commands

import (
	"context"
	"fmt"
	"os"

	"buddy/internal/apps/common"
	"buddy/internal/di"
	"buddy/internal/errors"
	"buddy/internal/logging"
	"buddy/internal/txn/domain"

	"github.com/spf13/cobra"
)

// BaseCommand provides common functionality for all commands
type BaseCommand struct {
	AppCtx  *common.Context
	Clients *di.ClientSet
	Logger  *logging.Logger
}

// NewBaseCommand creates a new base command
func NewBaseCommand(appCtx *common.Context, clients *di.ClientSet) *BaseCommand {
	return &BaseCommand{
		AppCtx:  appCtx,
		Clients: clients,
		Logger:  logging.NewDefaultLogger(fmt.Sprintf("%s-cmd", appCtx.Environment)),
	}
}

// HandleError provides consistent error handling across commands
func (bc *BaseCommand) HandleError(err error) {
	if err == nil {
		return
	}

	// Handle BuddyError types with context
	if buddyErr, ok := err.(*errors.BuddyError); ok {
		bc.Logger.Error("%s: %s", buddyErr.Type, buddyErr.Message)
		if len(buddyErr.Context) > 0 {
			bc.Logger.Debug("Error context: %+v", buddyErr.Context)
		}
		if buddyErr.Cause != nil {
			bc.Logger.Debug("Caused by: %v", buddyErr.Cause)
		}
	} else {
		bc.Logger.Error("Unexpected error: %v", err)
	}

	os.Exit(1)
}

// HandleTransactionResult provides consistent transaction result handling
func (bc *BaseCommand) HandleTransactionResult(result *domain.TransactionResult) {
	if result == nil {
		bc.HandleError(errors.Internal("received nil transaction result"))
		return
	}

	if result.Error != "" {
		fmt.Printf("%sError: %s\n", bc.AppCtx.GetPrefix(), result.Error)
		return
	}

	bc.Logger.Info("Transaction query completed successfully for ID: %s", result.InputID)
}

// ExecuteWithContext provides a wrapper for command execution with context
func (bc *BaseCommand) ExecuteWithContext(fn func(context.Context) error) error {
	ctx := context.Background()

	if err := fn(ctx); err != nil {
		bc.HandleError(err)
		return err
	}

	return nil
}

// ValidateRequiredFlags validates that required flags are provided
func (bc *BaseCommand) ValidateRequiredFlags(cmd *cobra.Command, required []string) error {
	for _, flag := range required {
		if value, _ := cmd.Flags().GetString(flag); value == "" {
			return errors.Validation(fmt.Sprintf("required flag --%s not provided", flag))
		}
	}
	return nil
}

// GetStringFlag safely gets a string flag value
func (bc *BaseCommand) GetStringFlag(cmd *cobra.Command, name string) (string, error) {
	value, err := cmd.Flags().GetString(name)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorTypeValidation, fmt.Sprintf("failed to get flag %s", name))
	}
	return value, nil
}

// GetBoolFlag safely gets a bool flag value
func (bc *BaseCommand) GetBoolFlag(cmd *cobra.Command, name string) (bool, error) {
	value, err := cmd.Flags().GetBool(name)
	if err != nil {
		return false, errors.Wrap(err, errors.ErrorTypeValidation, fmt.Sprintf("failed to get flag %s", name))
	}
	return value, nil
}

// PrintSuccess prints a success message with consistent formatting
func (bc *BaseCommand) PrintSuccess(message string, args ...any) {
	fmt.Printf("%s%s\n", bc.AppCtx.GetPrefix(), fmt.Sprintf(message, args...))
}

// PrintInfo prints an info message with consistent formatting
func (bc *BaseCommand) PrintInfo(message string, args ...any) {
	fmt.Printf("%s%s\n", bc.AppCtx.GetPrefix(), fmt.Sprintf(message, args...))
}
