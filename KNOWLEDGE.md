# Finding Related Records in Payment-Engine Given an RPP Record

This document explains how to find related records in payment-engine given an RPP (Real-time Payment Processing) record in the payment system.

## Overview

The payment system maintains records across two domains:
- **RPP Records**: Stored in the `credit_transfer` table, representing payment transactions in the RPP network
- **Payment-Engine Records**: External transaction records from the payment-engine service

These records are linked through specific identifier fields that enable correlation and status updates.

## Primary Linking Fields

The following table shows the field mappings between RPP and payment-engine records:

| RPP Field | Payment-Engine Field | Purpose | Example Usage |
|-----------|---------------------|---------|---------------|
| `EndToEndID` | `ExternalID` | Primary linkage for payment-engine stream events | Used when processing payment-engine status updates |
| `PartnerTxID` | `TransactionID` | Transaction identifier from payment-engine | Used for precise transaction matching |
| `ReqBizMsgID` | N/A | Unique Business Message ID for RPP request-response matching | Used to correlate RPP request and response messages |

### Field Definitions

**EndToEndID** (`end_to_end_id` in database)
- Type: `varchar(36)`
- Purpose: End-to-end identifier used by RPP for the entire transaction lifecycle
- Index: `index_req_end_to_end_id` (optimized for queries)

**PartnerTxID** (`partner_tx_id` in database)
- Type: `varchar(36)`
- Purpose: Transaction ID given by the service that initiates the transfer (payment-engine)
- Index: `index_partner_tx_id` (optimized for queries)

**ReqBizMsgID** (`req_biz_msg_id` in database)
- Type: `varchar(36)`
- Purpose: Unique Business Message ID for/from RPP request
- Index: `index_biz_msg_id` (unique constraint)
- Note: This field has a unique constraint (`uk_req_msg_id`)

## Query Patterns

### Pattern 1: Query by EndToEndID (Most Common)

Use this pattern when processing payment-engine stream events. This is the most common query pattern for finding RPP records based on payment-engine events.

```go
import (
    "context"
    "gitlab.myteksi.net/dakota/servus/v2/data"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/dbhelper"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/storage"
)

// Find credit transfer by EndToEndID (from payment-engine ExternalID)
func FindByEndToEndID(ctx context.Context, dbHelper storage.ICreditTransferDBHelper, externalID string) (*storage.CreditTransfer, error) {
    return dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, externalID),
    )
}
```

**Example Usage:**
```go
// In payment-engine stream handler
ct, err := c.CreditTransferDBHelper.FindOne(ctx,
    data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, entityData.ExternalID),
)
if err != nil {
    // Handle error
}
// Update record with payment-engine status
```

### Pattern 2: Query by ReqBizMsgID (Request-Response Matching)

Use this pattern when matching RPP request-response pairs within the RPP network.

```go
// Find credit transfer by ReqBizMsgID (for RPP request-response correlation)
func FindByReqBizMsgID(ctx context.Context, dbHelper storage.ICreditTransferDBHelper, reqBizMsgID string) (*storage.CreditTransfer, error) {
    return dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnReqBizMsgID, reqBizMsgID),
    )
}
```

**Example Usage:**
```go
// In RPP response handler
creditTransfer, err := w.CreditTransferDBHelper.FindOne(ctx,
    data.EqualTo(dbhelper.CreditTransferColumnReqBizMsgID, extractor.GetMessageReference()),
)
if err != nil {
    // Handle error
}
// Process RPP response
```

### Pattern 3: Query by Both EndToEndID and PartnerTxID (Precise Matching)

Use this pattern for the most precise matching when both identifiers are available. This ensures you match the exact transaction record.

```go
// Find credit transfer by both EndToEndID and PartnerTxID
func FindByEndToEndIDAndPartnerTxID(
    ctx context.Context,
    dbHelper storage.ICreditTransferDBHelper,
    externalID string,
    transactionID string,
) (*storage.CreditTransfer, error) {
    return dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, externalID),
        data.EqualTo(dbhelper.CreditTransferColumnPartnerTxID, transactionID),
    )
}
```

**Example Usage:**
```go
// In payment-engine stream handler (onPayment function)
ct, err := c.CreditTransferDBHelper.FindOne(ctx,
    data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, entityData.ExternalID),
    data.EqualTo(dbhelper.CreditTransferColumnPartnerTxID, entityData.TransactionID),
)
if ct == nil {
    return err
}
// Update record with payment-engine status
nextCt := *ct
nextCt.PartnerTxID = entityData.TransactionID
nextCt.PartnerTxStatus = entityData.Status
nextCt.PartnerTxStatusReason = entityData.StatusReason
_, errSave := c.CreditTransferDBHelper.UpdateEntity(ctx, ct, &nextCt)
```

## Key Files Reference

### RPP Storage Models
**File:** [`rpp-adapter/storage/credit_transfer.go`](rpp-adapter/storage/credit_transfer.go)
- Defines the [`CreditTransfer`](rpp-adapter/storage/credit_transfer.go:21) struct with all field mappings
- Contains column name constants for querying:
  - [`CreditTransferColumnEndToEndID`](rpp-adapter/storage/credit_transfer.go:71)
  - [`CreditTransferColumnReqBizMsgID`](rpp-adapter/storage/credit_transfer.go:70)
  - [`CreditTransferColumnPartnerTxID`](rpp-adapter/storage/credit_transfer.go:69)
- Implements [`ICreditTransferDBHelper`](rpp-adapter/storage/credit_transfer.go:76) interface with query methods

### Database Schema
**File:** [`rpp-adapter/db/mysql/deploy/0001-credit_transfer.sql`](rpp-adapter/db/mysql/deploy/0001-credit_transfer.sql)
- Defines the `credit_transfer` table structure
- Contains all column definitions with comments
- Includes indexes for optimized queries:
  - `index_req_end_to_end_id` on `end_to_end_id` column (line 47)
  - `index_biz_msg_id` on `req_biz_msg_id` column (line 45)
  - Unique constraint `uk_req_msg_id` on `req_biz_msg_id` (line 44)

### Payment-Engine Stream Handler
**File:** [`rpp-adapter/streams/consumers/payment_engine_stream_handler.go`](rpp-adapter/streams/consumers/payment_engine_stream_handler.go)
- Handles incoming payment-engine stream events
- Demonstrates query patterns in:
  - [`onPayment()`](rpp-adapter/streams/consumers/payment_engine_stream_handler.go:90) - Uses both EndToEndID and PartnerTxID (lines 98-99)
  - [`onCollection()`](rpp-adapter/streams/consumers/payment_engine_stream_handler.go:74) - Uses EndToEndID for workflow resumption (line 85)
- Shows how to update credit transfer records with payment-engine status (lines 104-108)

### DB Helper Constants
**File:** [`rpp-adapter/dbhelper/credit_transfer.go`](rpp-adapter/dbhelper/credit_transfer.go)
- Provides helper functions for querying credit transfer records
- Contains column name constants:
  - [`CreditTransferColumnEndToEndID`](rpp-adapter/dbhelper/credit_transfer.go:17)
  - [`CreditTransferColumnReqBizMsgID`](rpp-adapter/dbhelper/credit_transfer.go:19)
  - [`CreditTransferColumnPartnerTxID`](rpp-adapter/dbhelper/credit_transfer.go:22)
- Helper functions:
  - [`FindCreditTransferByCondition()`](rpp-adapter/dbhelper/credit_transfer.go:29) - Find single record
  - [`FindCreditTransfersByCondition()`](rpp-adapter/dbhelper/credit_transfer.go:48) - Find multiple records

## Database Indexes

The following indexes are defined on the `credit_transfer` table to optimize query performance:

| Index Name | Column(s) | Purpose |
|------------|-----------|---------|
| `index_req_end_to_end_id` | `end_to_end_id` | Optimizes queries by EndToEndID (Pattern 1) |
| `index_biz_msg_id` | `req_biz_msg_id` | Optimizes queries by ReqBizMsgID (Pattern 2) |
| `index_partner_tx_id` | `partner_tx_id` | Optimizes queries by PartnerTxID (Pattern 3) |
| `index_partner_msg_id` | `partner_msg_id` | Optimizes queries by PartnerMsgID |

**Note:** The `req_biz_msg_id` column also has a unique constraint (`uk_req_msg_id`), ensuring each RPP request has a unique business message ID.

## Complete Example Usage

### Example 1: Processing Payment-Engine Stream Events

```go
package main

import (
    "context"
    "gitlab.myteksi.net/dakota/servus/v2/data"
    "gitlab.myteksi.net/dakota/servus/v2/slog"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/dbhelper"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/storage"
)

// ProcessPaymentEngineEvent processes a payment-engine stream event
func ProcessPaymentEngineEvent(
    ctx context.Context,
    dbHelper storage.ICreditTransferDBHelper,
    externalID string,
    transactionID string,
    status string,
    statusReason string,
) error {
    // Pattern 3: Query by both EndToEndID and PartnerTxID for precise matching
    ct, err := dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, externalID),
        data.EqualTo(dbhelper.CreditTransferColumnPartnerTxID, transactionID),
    )
    if err != nil {
        slog.FromContext(ctx).Error("Failed to find credit transfer", slog.Error(err))
        return err
    }
    if ct == nil {
        slog.FromContext(ctx).Warn("Credit transfer not found",
            slog.String("externalID", externalID),
            slog.String("transactionID", transactionID))
        return nil
    }

    // Update the record with payment-engine status
    nextCt := *ct
    nextCt.PartnerTxID = transactionID
    nextCt.PartnerTxStatus = status
    nextCt.PartnerTxStatusReason = statusReason

    _, err = dbHelper.UpdateEntity(ctx, ct, &nextCt)
    if err != nil {
        slog.FromContext(ctx).Error("Failed to update credit transfer", slog.Error(err))
        return err
    }

    slog.FromContext(ctx).Info("Successfully updated credit transfer",
        slog.String("reqBizMsgID", ct.ReqBizMsgID),
        slog.String("status", status))
    return nil
}
```

### Example 2: Handling RPP Response Messages

```go
package main

import (
    "context"
    "gitlab.myteksi.net/dakota/servus/v2/data"
    "gitlab.myteksi.net/dakota/servus/v2/slog"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/dbhelper"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/storage"
)

// HandleRPPResponse processes an RPP response message
func HandleRPPResponse(
    ctx context.Context,
    dbHelper storage.ICreditTransferDBHelper,
    reqBizMsgID string,
    txStatus string,
    txStatusReason string,
) error {
    // Pattern 2: Query by ReqBizMsgID for request-response matching
    creditTransfer, err := dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnReqBizMsgID, reqBizMsgID),
    )
    if err != nil {
        slog.FromContext(ctx).Error("Failed to find credit transfer", slog.Error(err))
        return err
    }
    if creditTransfer == nil {
        slog.FromContext(ctx).Warn("Credit transfer not found",
            slog.String("reqBizMsgID", reqBizMsgID))
        return nil
    }

    // Update the record with RPP response status
    nextCt := *creditTransfer
    nextCt.TxStatus = txStatus
    nextCt.TxStatusReason = txStatusReason

    _, err = dbHelper.UpdateEntity(ctx, creditTransfer, &nextCt)
    if err != nil {
        slog.FromContext(ctx).Error("Failed to update credit transfer", slog.Error(err))
        return err
    }

    slog.FromContext(ctx).Info("Successfully updated credit transfer with RPP response",
        slog.String("reqBizMsgID", reqBizMsgID),
        slog.String("txStatus", txStatus))
    return nil
}
```

### Example 3: Querying by EndToEndID Only

```go
package main

import (
    "context"
    "gitlab.myteksi.net/dakota/servus/v2/data"
    "gitlab.myteksi.net/dakota/servus/v2/slog"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/dbhelper"
    "gitlab.com/gx-regional/dbmy/payment/rpp-adapter/storage"
)

// FindByEndToEndIDOnly finds a credit transfer by EndToEndID only
func FindByEndToEndIDOnly(
    ctx context.Context,
    dbHelper storage.ICreditTransferDBHelper,
    endToEndID string,
) (*storage.CreditTransfer, error) {
    // Pattern 1: Query by EndToEndID (most common)
    ct, err := dbHelper.FindOne(ctx,
        data.EqualTo(dbhelper.CreditTransferColumnEndToEndID, endToEndID),
    )
    if err != nil {
        slog.FromContext(ctx).Error("Failed to find credit transfer", slog.Error(err))
        return nil, err
    }

    slog.FromContext(ctx).Info("Found credit transfer",
        slog.String("endToEndID", endToEndID),
        slog.String("reqBizMsgID", ct.ReqBizMsgID))
    return ct, nil
}
```

## Best Practices

1. **Use the most specific query pattern available:**
   - If you have both `EndToEndID` and `PartnerTxID`, use Pattern 3 (both fields)
   - If you only have `EndToEndID`, use Pattern 1
   - For RPP request-response correlation, use Pattern 2 with `ReqBizMsgID`

2. **Always handle nil results:**
   - Check if the returned record is `nil` before accessing its fields
   - Log warnings when records are not found for debugging

3. **Use appropriate logging:**
   - Include relevant identifiers in log messages for traceability
   - Use different log levels (Info, Warn, Error) appropriately

4. **Leverage database indexes:**
   - The queries shown above use indexed columns for optimal performance
   - Avoid querying on non-indexed columns when possible

5. **Handle errors gracefully:**
   - Check for `data.ErrNoData` when records are not found
   - Distinguish between "not found" and actual errors

## Related Documentation

- [RPP Adapter README](README.md) - General documentation for the RPP adapter
- [Database Migration Files](rpp-adapter/db/mysql/deploy/) - All database schema changes
- [Payment-Engine Stream Handler](rpp-adapter/streams/consumers/payment_engine_stream_handler.go) - Implementation details

## Summary

Finding related records in payment-engine given an RPP record involves:

1. **Understanding the field mappings** between RPP and payment-engine systems
2. **Using the appropriate query pattern** based on available identifiers
3. **Leveraging optimized database indexes** for performance
4. **Following best practices** for error handling and logging

The three main query patterns are:
- **Pattern 1:** Query by `EndToEndID` (most common for payment-engine events)
- **Pattern 2:** Query by `ReqBizMsgID` (for RPP request-response matching)
- **Pattern 3:** Query by both `EndToEndID` and `PartnerTxID` (for precise matching)

All queries are optimized through database indexes on the respective columns, ensuring efficient record lookups in high-volume payment processing scenarios.
