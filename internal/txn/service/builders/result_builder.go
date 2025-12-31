package builders

import "buddy/internal/txn/domain"

// TransactionResultBuilder provides a fluent API for building TransactionResult
type TransactionResultBuilder struct {
	result *domain.TransactionResult
}

// NewResultBuilder creates a new ResultBuilder with initialized TransactionResult
func NewResultBuilder() *TransactionResultBuilder {
	return &TransactionResultBuilder{
		result: &domain.TransactionResult{
			CaseType: domain.CaseNone,
		},
	}
}

// SetInputID sets the input ID
func (b *TransactionResultBuilder) SetInputID(id string) *TransactionResultBuilder {
	b.result.InputID = id
	return b
}

// SetPaymentEngine sets the PaymentEngine info (nil-safe)
func (b *TransactionResultBuilder) SetPaymentEngine(info *domain.PaymentEngineInfo) *TransactionResultBuilder {
	if info != nil {
		b.result.PaymentEngine = info
	}
	return b
}

// SetPaymentCore sets the PaymentCore info (nil-safe)
func (b *TransactionResultBuilder) SetPaymentCore(info *domain.PaymentCoreInfo) *TransactionResultBuilder {
	if info != nil {
		b.result.PaymentCore = info
	}
	return b
}

// SetRPPAdapter sets the RPP adapter info (nil-safe)
func (b *TransactionResultBuilder) SetRPPAdapter(info *domain.RPPAdapterInfo) *TransactionResultBuilder {
	if info != nil {
		b.result.RPPAdapter = info
	}
	return b
}

// SetFastAdapter sets the Fast adapter info (nil-safe)
func (b *TransactionResultBuilder) SetFastAdapter(info *domain.FastAdapterInfo) *TransactionResultBuilder {
	if info != nil {
		b.result.FastAdapter = info
	}
	return b
}

// SetPartnerpayEngine sets the PartnerpayEngine info (nil-safe)
func (b *TransactionResultBuilder) SetPartnerpayEngine(info *domain.PartnerpayEngineInfo) *TransactionResultBuilder {
	if info != nil {
		b.result.PartnerpayEngine = info
	}
	return b
}

// SetError sets the error message
func (b *TransactionResultBuilder) SetError(err string) *TransactionResultBuilder {
	b.result.Error = err
	return b
}

// SetCaseType sets the SOP case type
func (b *TransactionResultBuilder) SetCaseType(caseType domain.Case) *TransactionResultBuilder {
	b.result.CaseType = caseType
	return b
}

// Build returns the constructed TransactionResult
func (b *TransactionResultBuilder) Build() *domain.TransactionResult {
	return b.result
}
