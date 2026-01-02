package adapters

import "buddy/internal/txn/domain"

// TemplateFunc represents a SQL template function
type TemplateFunc func(domain.TransactionResult) *domain.DMLTicket

// sqlTemplates maps SOP cases to their DML tickets
var sqlTemplates = map[domain.Case]TemplateFunc{}

// init initializes all template registrations
func init() {
	registerPCTemplates(sqlTemplates)
	registerPEBasicTemplates(sqlTemplates)
	registerPEAdvancedTemplates(sqlTemplates)
	registerRPPBasicTemplates(sqlTemplates)
	registerRPPAdvancedTemplates(sqlTemplates)
	registerPPETemplates(sqlTemplates)
	registerCrossDomainTemplates(sqlTemplates)
}
