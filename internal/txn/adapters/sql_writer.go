package adapters

import (
	"buddy/internal/txn/domain"
	"bytes"
	"fmt"
	"os"
)

// WriteSQLFiles writes the SQL statements to database-specific Deploy.sql and Rollback.sql files
func WriteSQLFiles(statements domain.SQLStatements, basePath string) error {
	// Write PC files (always append to fixed filenames)
	if len(statements.PCDeployStatements) > 0 {
		deployPath := "PC_Deploy.sql"
		if err := AppendSQLFile(deployPath, statements.PCDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		rollbackPath := "PC_Rollback.sql"
		if err := AppendSQLFile(rollbackPath, statements.PCRollbackStatements); err != nil {
			return err
		}
	}

	// Write PE files (always append to fixed filenames)
	if len(statements.PEDeployStatements) > 0 {
		deployPath := "PE_Deploy.sql"
		if err := AppendSQLFile(deployPath, statements.PEDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		rollbackPath := "PE_Rollback.sql"
		if err := AppendSQLFile(rollbackPath, statements.PERollbackStatements); err != nil {
			return err
		}
	}

	// Write PPE files (always append to fixed filenames)
	if len(statements.PPEDeployStatements) > 0 {
		deployPath := "PPE_Deploy.sql"
		if err := AppendSQLFile(deployPath, statements.PPEDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PPERollbackStatements) > 0 {
		rollbackPath := "PPE_Rollback.sql"
		if err := AppendSQLFile(rollbackPath, statements.PPERollbackStatements); err != nil {
			return err
		}
	}

	// Write RPP files (always append to fixed filenames)
	if len(statements.RPPDeployStatements) > 0 {
		deployPath := "RPP_Deploy.sql"
		if err := AppendSQLFile(deployPath, statements.RPPDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		rollbackPath := "RPP_Rollback.sql"
		if err := AppendSQLFile(rollbackPath, statements.RPPRollbackStatements); err != nil {
			return err
		}
	}

	return nil
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

	fmt.Printf("SQL statements written to %s\n", filePath)
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
