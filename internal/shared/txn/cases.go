package txn

// IdentifySOPCases identifies SOP cases for all transactions without generating SQL
// Deprecated: Use SOPRepository.IdentifySOPCase for new implementations
func IdentifySOPCases(results []TransactionResult) {
	IdentifySOPCasesWithEnv(results, "my") // Default to MY for backward compatibility
}

// IdentifySOPCasesWithEnv identifies SOP cases for all transactions with specified environment
// Deprecated: Use SOPRepository.IdentifySOPCase for new implementations
func IdentifySOPCasesWithEnv(results []TransactionResult, env string) {
	sopRepo := NewSOPRepository()
	for i := range results {
		if results[i].CaseType == SOPCaseNone {
			// Use the new rule-based approach
			sopRepo.IdentifySOPCase(&results[i], env)
		}
	}
}
