package jira

import (
	"encoding/csv"
	"strings"

	"buddy/internal/errors"
)

// ParseCSVAttachment parses CSV content from JIRA attachments
func (c *JiraClient) ParseCSVAttachment(content string) ([]CSVRow, error) {
	reader := csv.NewReader(strings.NewReader(content))
	records, err := reader.ReadAll()
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorTypeValidation, "failed to parse CSV")
	}

	if len(records) == 0 {
		return nil, nil
	}

	fieldMappings := c.getFieldMappings()
	headerRowIndex := c.findHeaderRowIndex(records, fieldMappings)

	if headerRowIndex < 0 {
		return nil, errors.Validation("header row not found in CSV")
	}

	c.mapColumnIndices(records[headerRowIndex], fieldMappings)

	var rows []CSVRow
	for _, row := range records[headerRowIndex+1:] {
		if c.isEmptyRow(row) || c.isSummaryRow(row) {
			continue
		}

		csvRow := c.processCSVRow(row, fieldMappings)
		if csvRow != nil {
			rows = append(rows, *csvRow)
		}
	}

	return rows, nil
}

// csvFieldMapping represents a mapping between CSV fields and their column indices
type csvFieldMapping struct {
	Index  int
	Fields []string
}

// getFieldMappings returns the mapping configuration for CSV fields
func (c *JiraClient) getFieldMappings() map[string]*csvFieldMapping {
	return map[string]*csvFieldMapping{
		"transaction_date": {
			Fields: []string{"date"},
		},
		"batch_id": {
			Fields: []string{"batch id", "partner_tx_id"},
		},
		"end_to_end_id": {
			Fields: []string{"tar02 bmid", "original_bizmsgid"},
		},
		"transaction_id": {
			Fields: []string{"transaction id"},
		},
		"req_biz_msg_id": {
			Fields: []string{"req_biz_msg_id"},
		},
		"internal_status": {
			Fields: []string{"dbmy status"},
		},
		"paynet_status": {
			Fields: []string{"column_status", "tar02 sts", "rpp_status"},
		},
	}
}

// findHeaderRowIndex finds the index of the header row in CSV records
func (c *JiraClient) findHeaderRowIndex(records [][]string, mappings map[string]*csvFieldMapping) int {
	for i, row := range records {
		if c.findHeaderRow(row, mappings) {
			return i
		}
	}
	return -1
}

// findHeaderRow determines if a row is the header row
func (c *JiraClient) findHeaderRow(row []string, mappings map[string]*csvFieldMapping) bool {
	if len(row) == 0 {
		return false
	}

	lowerRow := make([]string, len(row))
	for i, cell := range row {
		lowerRow[i] = strings.ToLower(strings.TrimSpace(cell))
	}

	matches := 0
	for _, mapping := range mappings {
		for _, field := range mapping.Fields {
			for _, cell := range lowerRow {
				if cell == field {
					matches++
					break
				}
			}
		}
	}

	// Consider it a header if we match at least 3 fields
	return matches >= 3
}

// mapColumnIndices maps column headers to their indices
func (c *JiraClient) mapColumnIndices(headerRow []string, mappings map[string]*csvFieldMapping) {
	for i, header := range headerRow {
		lowerHeader := strings.ToLower(strings.TrimSpace(header))

		for _, mapping := range mappings {
			for _, field := range mapping.Fields {
				if lowerHeader == field {
					mapping.Index = i
					break
				}
			}
		}
	}
}

// isEmptyRow checks if a CSV row is empty
func (c *JiraClient) isEmptyRow(row []string) bool {
	for _, cell := range row {
		if strings.TrimSpace(cell) != "" {
			return false
		}
	}
	return true
}

// isSummaryRow checks if a CSV row is a summary row
func (c *JiraClient) isSummaryRow(row []string) bool {
	if len(row) == 0 {
		return false
	}
	return strings.Contains(strings.ToLower(strings.TrimSpace(row[0])), "total")
}

// processCSVRow processes a single CSV row into a CSVRow struct
func (c *JiraClient) processCSVRow(row []string, mappings map[string]*csvFieldMapping) *CSVRow {
	csvRow := &CSVRow{}
	hasData := false

	for fieldName, mapping := range mappings {
		if mapping.Index >= 0 && mapping.Index < len(row) {
			value := strings.TrimSpace(row[mapping.Index])
			if value != "" && value != "-" {
				ptr := &value
				switch fieldName {
				case "transaction_date":
					csvRow.TransactionDate = ptr
				case "batch_id":
					csvRow.BatchID = ptr
				case "end_to_end_id":
					csvRow.EndToEndID = ptr
				case "transaction_id":
					csvRow.TransactionID = ptr
				case "req_biz_msg_id":
					csvRow.ReqBizMsgID = ptr
				case "internal_status":
					csvRow.InternalStatus = ptr
				case "paynet_status":
					csvRow.PaynetStatus = ptr
				}
				hasData = true
			}
		}
	}

	if !hasData {
		return nil
	}

	return csvRow
}
