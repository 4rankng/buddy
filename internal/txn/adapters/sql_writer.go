package adapters

import (
	"buddy/internal/txn/domain"
	"bytes"
	"fmt"
	"os"
)

// ClearSQLFiles removes all existing SQL files before batch processing
func ClearSQLFiles() {
	sqlFiles := []string{
		"PC_Deploy.sql", "PC_Rollback.sql",
		"PE_Deploy.sql", "PE_Rollback.sql",
		"PPE_Deploy.sql", "PPE_Rollback.sql",
		"RPP_Deploy.sql", "RPP_Rollback.sql",
	}
	for _, file := range sqlFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to remove %s: %v\n", file, err)
		}
	}
}

// WriteSQLFiles writes the SQL statements to database-specific Deploy.sql and Rollback.sql files
// Returns a list of filenames that were successfully created
func WriteSQLFiles(statements domain.SQLStatements, basePath string) ([]string, error) {
	var filesCreated []string

	// Validate that rollback statements exist when deploy statements exist
	if err := validateDeployRollbackPairs(statements); err != nil {
		return filesCreated, fmt.Errorf("deploy/rollback validation failed: %w", err)
	}

	// Write PC files (always overwrite fixed filenames)
	if len(statements.PCDeployStatements) > 0 {
		deployPath := "PC_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.PCDeployStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, deployPath)
	}
	if len(statements.PCRollbackStatements) > 0 {
		rollbackPath := "PC_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.PCRollbackStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, rollbackPath)
	}

	// Write PE files (always overwrite fixed filenames)
	if len(statements.PEDeployStatements) > 0 {
		deployPath := "PE_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.PEDeployStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, deployPath)
	}
	if len(statements.PERollbackStatements) > 0 {
		rollbackPath := "PE_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.PERollbackStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, rollbackPath)
	}

	// Write PPE files (always overwrite fixed filenames)
	if len(statements.PPEDeployStatements) > 0 {
		deployPath := "PPE_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.PPEDeployStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, deployPath)
	}
	if len(statements.PPERollbackStatements) > 0 {
		rollbackPath := "PPE_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.PPERollbackStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, rollbackPath)
	}

	// Write RPP files (always overwrite fixed filenames)
	if len(statements.RPPDeployStatements) > 0 {
		deployPath := "RPP_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.RPPDeployStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, deployPath)
	}
	if len(statements.RPPRollbackStatements) > 0 {
		rollbackPath := "RPP_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.RPPRollbackStatements); err != nil {
			return filesCreated, err
		}
		filesCreated = append(filesCreated, rollbackPath)
	}

	return filesCreated, nil
}

// WriteSQLFile writes SQL statements to a file
func WriteSQLFile(filePath string, statements []string) error {
	var buffer bytes.Buffer
	for _, stmt := range statements {
		if _, err := buffer.WriteString(stmt); err != nil {
			fmt.Printf("Warning: failed to buffer SQL statement: %v\n", err)
		}
		if _, err := buffer.WriteString("\n\n"); err != nil {
			fmt.Printf("Warning: failed to buffer newline: %v\n", err)
		}
	}

	if err := os.WriteFile(filePath, buffer.Bytes(), 0o644); err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	return nil
}

// AppendSQLFile appends SQL statements to a file
func AppendSQLFile(filePath string, statements []string) error {
	// Check if file exists to determine if we need to add separator
	_, err := os.Stat(filePath)
	fileExists := err == nil

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Printf("Warning: failed to close file %s: %v\n", filePath, closeErr)
		}
	}()

	// Add separator if file is not empty
	if fileExists {
		if _, err := file.WriteString("\n\n"); err != nil {
			fmt.Printf("Warning: failed to write separator: %v\n", err)
		}
	}

	for _, stmt := range statements {
		if _, err := file.WriteString(stmt); err != nil {
			fmt.Printf("Warning: failed to write SQL statement: %v\n", err)
		}
		if _, err := file.WriteString("\n\n"); err != nil {
			fmt.Printf("Warning: failed to write newline: %v\n", err)
		}
	}

	fmt.Printf("SQL statements appended to %s\n", filePath)
	return nil
}

// validateDeployRollbackPairs ensures that rollback statements exist when deploy statements exist
func validateDeployRollbackPairs(statements domain.SQLStatements) error {
	var missingRollbacks []string

	// Check PC
	if len(statements.PCDeployStatements) > 0 && len(statements.PCRollbackStatements) == 0 {
		missingRollbacks = append(missingRollbacks, "PC")
	}

	// Check PE
	if len(statements.PEDeployStatements) > 0 && len(statements.PERollbackStatements) == 0 {
		missingRollbacks = append(missingRollbacks, "PE")
	}

	// Check PPE
	if len(statements.PPEDeployStatements) > 0 && len(statements.PPERollbackStatements) == 0 {
		missingRollbacks = append(missingRollbacks, "PPE")
	}

	// Check RPP
	if len(statements.RPPDeployStatements) > 0 && len(statements.RPPRollbackStatements) == 0 {
		missingRollbacks = append(missingRollbacks, "RPP")
	}

	if len(missingRollbacks) > 0 {
		fmt.Printf("[ERROR] Deploy/rollback validation failed: missing rollback statements for databases: %v\n", missingRollbacks)
		return fmt.Errorf("missing rollback statements for databases: %v", missingRollbacks)
	}

	return nil
}
