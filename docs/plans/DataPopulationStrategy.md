# Implementation Plan: Transaction Data Population Refactoring

## Objective
Refactor `transaction_service.go` to solve file size violations (>500 lines) and separate concerns by implementing a Strategy Pattern for data population.

## Context

- **Current Service**: `internal/txn/service/transaction_service.go` (monolithic)
- **Target Structure**: `internal/txn/service/population/` (strategies & populators)

---

## Phase 1: Foundation & Interfaces

Establish the directory structure and type definitions to decouple dependencies.

### Create Directory Structure

Create folders:

- `internal/txn/service/population/`
- `internal/txn/service/strategies/`
- `internal/txn/service/builders/`

### Create Constants & Errors

**File**: `internal/txn/service/constants.go`

**Action**: Define constants for time windows:

```go
PEExternalIDWindow = 30 * time.Minute
PCRegularTxnWindow = 1 * time.Hour
RPPProcessRegistryWindow = 5 * time.Minute
FastInstructionIDWindow = 1 * time.Hour
```

**Action**: Define `PopulationError` struct and `AdapterType` constants.

### Define Core Interfaces

**File**: `internal/txn/service/population/interfaces.go`

**Action**: Define the following interfaces:

- `PaymentEnginePopulator` (Methods: `QueryByTransactionID`, `QueryByExternalID`, `QueryWorkflow`)
- `PaymentCorePopulator` (Methods: `QueryInternal`, `QueryExternal`)
- `AdapterPopulator` (Methods: `QueryByInputID`, `GetAdapterType`)
- `PopulationStrategy` (Methods: `Populate(input string) (*TransactionResult, error)`)

### Create Result Builder

**File**: `internal/txn/service/builders/result_builder.go`

**Action**: Create a `ResultBuilder` struct that wraps `domain.TransactionResult`.

**Action**: Implement setters (`SetPaymentEngine`, `SetPaymentCore`, `SetAdapterInfo`) that handle nil checks safely.

---

## Phase 2: Populator Implementation

Move raw data fetching/parsing logic out of the service and into dedicated, single-responsibility structs.

### Implement Payment Engine Populator

**File**: `internal/txn/service/population/payment_engine_populator.go`

**Action**: Create struct `pePopulator` implementing `PaymentEnginePopulator`.

**Action**: Move logic from `fillPaymentEngineFromTransactionID` and `populatePaymentEngineFromAdapters` here.

**Requirement**: Ensure `QueryWorkflow` is called automatically if `reference_id` is found.

### Implement Payment Core Populator

**File**: `internal/txn/service/population/payment_core_populator.go`

**Action**: Create struct `pcPopulator` implementing `PaymentCorePopulator`.

**Action**: Move logic from `populatePaymentCoreInfo` here.

**Requirement**: Use the `PCRegularTxnWindow` constant for time filtering.

### Implement Adapter Populators

**File**: `internal/txn/service/population/adapter_populator.go`

**Action**: Create `rppPopulator` (Malaysia) and `fastPopulator` (Singapore) implementing `AdapterPopulator`.

**Action**: Move logic from `fillAdapterFromE2EID` here.

**Requirement**: Handle RPP's dual workflow query strategy within `rppPopulator`.

---

## Phase 3: Strategy Implementation

Implement the orchestration logic that ties populators together based on environment.

### Implement Malaysia Strategy

**File**: `internal/txn/service/strategies/malaysia_strategy.go`

**Action**: Implement `MalaysiaPopulationStrategy`.

**Logic**:

1. Check if input is E2E ID → Call `rppPopulator`.
2. If RPP found → Extract External ID → Call `pePopulator`.
3. If Input is TxnID → Call `pePopulator` directly.
4. Once PE is populated → Call `pcPopulator` using PE Transaction ID.

### Implement Singapore Strategy

**File**: `internal/txn/service/strategies/singapore_strategy.go`

**Action**: Implement `SingaporePopulationStrategy`.

**Logic**: Similar to Malaysia, but use `fastPopulator` and query by Instruction ID.

### Create Strategy Factory

**File**: `internal/txn/service/population/factory.go`

**Action**: Create `NewPopulationStrategy(env string, adapters ...)` that returns the correct strategy implementation based on the env string ("my" vs "sg").

---

## Phase 4: Service Refactoring & Cleanup

Wire everything up in the main service, effectively deleting the old monolithic code.

### Refactor Transaction Service

**File**: `internal/txn/service/transaction_service.go`

**Action**: Inject `PopulationStrategy` into `TransactionQueryService`.

**Action**: Rewrite `QueryTransactionWithEnv` to strictly call `strategy.Populate(input)`.

**Action**: Remove all private population methods (`fillPaymentEngine...`, `populatePaymentCore...`, etc.).

**Goal**: File size should drop to < 200 lines.

### Verify & Test

**Action**: Run existing tests for `TransactionQueryService`.

**Action**: Verify that `SOPRepository` still receives the correct data structure for rule evaluation.