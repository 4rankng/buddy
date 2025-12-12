package txn

import (
	"bytes"
	"fmt"
	"os"
)

// WriteSQLFiles writes the SQL statements to database-specific Deploy.sql and Rollback.sql files
func WriteSQLFiles(statements SQLStatements, basePath string) error {
	// Write PC files
	if len(statements.PCDeployStatements) > 0 {
		deployPath := basePath + "_PC_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.PCDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PCRollbackStatements) > 0 {
		rollbackPath := basePath + "_PC_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.PCRollbackStatements); err != nil {
			return err
		}
	}

	// Write PE files
	if len(statements.PEDeployStatements) > 0 {
		deployPath := basePath + "_PE_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.PEDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.PERollbackStatements) > 0 {
		rollbackPath := basePath + "_PE_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.PERollbackStatements); err != nil {
			return err
		}
	}

	// Write RPP files
	if len(statements.RPPDeployStatements) > 0 {
		deployPath := basePath + "_RPP_Deploy.sql"
		if err := WriteSQLFile(deployPath, statements.RPPDeployStatements); err != nil {
			return err
		}
	}
	if len(statements.RPPRollbackStatements) > 0 {
		rollbackPath := basePath + "_RPP_Rollback.sql"
		if err := WriteSQLFile(rollbackPath, statements.RPPRollbackStatements); err != nil {
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
