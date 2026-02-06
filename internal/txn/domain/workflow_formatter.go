package domain

import (
	"fmt"
	"strconv"
	"sync"

	"buddy/internal/config"
	"buddy/internal/constants"
	"buddy/internal/logging"
)

var (
	workflowStates    map[string]map[int]string
	fastAdapterStates map[string]map[int]string
	statesInitOnce    sync.Once
	statesInitError   error
	logger            = logging.NewDefaultLogger("domain")
)

// initializeStates loads workflow and fast adapter states from configuration
func initializeStates() {
	statesInitOnce.Do(func() {
		var err error

		// Load workflow states
		workflowStates, err = config.GetWorkflowStates()
		if err != nil {
			logger.Error("Failed to load workflow states: %v", err)
			statesInitError = err
			return
		}

		// Load fast adapter states (optional)
		fastAdapterStates, err = config.GetFastAdapterStates()
		if err != nil {
			logger.Warn("Failed to load fast adapter states, using empty mappings: %v", err)
			fastAdapterStates = make(map[string]map[int]string)
		}
	})
}

// FormatWorkflowState formats a workflow state using external configuration
func FormatWorkflowState(workflowID, state string) string {
	initializeStates()
	if statesInitError != nil {
		logger.Error("Workflow states not initialized: %v", statesInitError)
		return state
	}

	stateInt, err := strconv.Atoi(state)
	if err != nil {
		return state
	}

	if stateMap, exists := workflowStates[workflowID]; exists {
		if stateName, found := stateMap[stateInt]; found {
			return fmt.Sprintf("%d (%s)", stateInt, stateName)
		}
	}

	return fmt.Sprintf("%d", stateInt)
}

// FormatFastAdapterState formats a fast adapter state using external configuration
func FormatFastAdapterState(adapterType string, statusCode int) string {
	initializeStates()
	if statesInitError != nil {
		logger.Error("Fast adapter states not initialized: %v", statesInitError)
		return fmt.Sprintf("%s:%d", adapterType, statusCode)
	}

	if stateMap, exists := fastAdapterStates[adapterType]; exists {
		if stateName, found := stateMap[statusCode]; found {
			return fmt.Sprintf("%s:%d (%s)", adapterType, statusCode, stateName)
		}
	}

	return fmt.Sprintf("%s:%d", adapterType, statusCode)
}

// GetWorkflowStateMap returns the workflow state map for a specific workflow
func GetWorkflowStateMap(workflowID string) (map[int]string, bool) {
	initializeStates()
	if statesInitError != nil {
		return nil, false
	}

	stateMap, exists := workflowStates[workflowID]
	return stateMap, exists
}

// GetFastAdapterStateMap returns the fast adapter state map for a specific adapter type
func GetFastAdapterStateMap(adapterType string) (map[int]string, bool) {
	initializeStates()
	if statesInitError != nil {
		return nil, false
	}

	stateMap, exists := fastAdapterStates[adapterType]
	return stateMap, exists
}

// IsKnownWorkflow checks if a workflow ID is known in the configuration
func IsKnownWorkflow(workflowID string) bool {
	switch workflowID {
	case constants.WorkflowTransferPayment,
		constants.WorkflowInternalPayment,
		constants.WorkflowExternalPayment,
		constants.WorkflowCashIn,
		constants.WorkflowCashOut:
		return true
	default:
		initializeStates()
		if statesInitError != nil {
			return false
		}
		_, exists := workflowStates[workflowID]
		return exists
	}
}

// IsKnownFastAdapterType checks if a fast adapter type is known in the configuration
func IsKnownFastAdapterType(adapterType string) bool {
	switch adapterType {
	case constants.FastTypeCashIn, constants.FastTypeCashOut:
		return true
	default:
		initializeStates()
		if statesInitError != nil {
			return false
		}
		_, exists := fastAdapterStates[adapterType]
		return exists
	}
}
