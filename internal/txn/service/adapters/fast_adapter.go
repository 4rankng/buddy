package adapters

import (
	"buddy/internal/txn/domain"
	"buddy/internal/txn/ports"
	"buddy/internal/txn/utils"
	"fmt"
	"strconv"
	"time"
)

// FastAdapter implements the FastAdapterPort interface
type FastAdapter struct {
	client ports.ClientPort
}

// NewFastAdapter creates a new FastAdapter
func NewFastAdapter(client ports.ClientPort) *FastAdapter {
	return &FastAdapter{
		client: client,
	}
}

func (f *FastAdapter) QueryByInstructionID(instructionID, createdAt string) (*domain.FastAdapterInfo, error) {
	if instructionID == "" {
		return nil, nil
	}
	query := fmt.Sprintf("SELECT type, instruction_id, status, cancel_reason_code, reject_reason_code, created_at FROM transactions WHERE instruction_id = '%s'", instructionID)
	if createdAt != "" {
		startTime, err := time.Parse(time.RFC3339, createdAt)
		if err == nil {
			endTime := startTime.Add(1 * time.Hour)
			query += fmt.Sprintf(" AND created_at >= '%s' AND created_at <= '%s'", createdAt, endTime.Format(time.RFC3339))
		}
	}
	query += " LIMIT 1"
	results, err := f.client.QueryFastAdapter(query)
	if err != nil || len(results) == 0 {
		return nil, err
	}
	row := results[0]
	info := &domain.FastAdapterInfo{
		InstructionID:    utils.GetStringValue(row, "instruction_id"),
		Type:             utils.GetStringValue(row, "type"),
		CreatedAt:        utils.GetStringValue(row, "created_at"),
		CancelReasonCode: utils.GetStringValue(row, "cancel_reason_code"),
		RejectReasonCode: utils.GetStringValue(row, "reject_reason_code"),
	}
	if statusVal, ok := row["status"]; ok {
		var statusNum int
		if str, ok := statusVal.(string); ok {
			if num, err := strconv.Atoi(str); err == nil {
				statusNum = num
			}
		} else if num, ok := statusVal.(int); ok {
			statusNum = num
		} else if fnum, ok := statusVal.(float64); ok {
			statusNum = int(fnum)
		}

		// Look up status name from domain.FastAdapterStateMaps
		if stateMap, exists := domain.FastAdapterStateMaps[info.Type]; exists {
			if statusName, found := stateMap[statusNum]; found {
				info.Status = statusName
			} else {
				info.Status = "UNKNOWN"
			}
		} else {
			// Fallback for unknown adapter types
			switch statusNum {
			case 0:
				info.Status = "INITIATED"
			case 1:
				info.Status = "PENDING"
			case 2:
				info.Status = "PROCESSING"
			case 3:
				info.Status = "SUCCESS"
			case 4:
				info.Status = "FAILED"
			case 5:
				info.Status = "CANCELLED"
			case 6:
				info.Status = "REJECTED"
			case 7:
				info.Status = "TIMEOUT"
			case 8:
				info.Status = "ERROR"
			default:
				info.Status = "UNKNOWN"
			}
		}
		info.StatusCode = statusNum
	}
	return info, nil
}
