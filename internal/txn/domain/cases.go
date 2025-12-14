package domain

// IdentifySOPCases identifies SOP cases for all transactions without generating SQL
// Deprecated: This function has been moved to adapters package to maintain clean architecture
// Use adapters.NewSOPRepository().IdentifySOPCase() instead
func IdentifySOPCases(results []TransactionResult) {
	// This function is deprecated. Please use the adapters package directly.
}
