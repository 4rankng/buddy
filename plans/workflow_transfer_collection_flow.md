# Workflow Transfer Collection State Flow

## State Flow Diagram

```mermaid
graph TD
    A[stTransferPersisted 100] --> B[stProcessingPublished 101]
    B --> C[stPreTransactionLimitCheck 102]
    C --> D[stPreRiskCheck 103]
    
    D --> E[stTransferProcessing 210]
    E --> F[stTransferStreamPersisted 211]
    
    D --> G[stAuthProcessing 220]
    G --> H[stAuthStreamPersisted 221]
    H --> I[stAuthSuccess 300]
    
    I --> J[stCapturePrepared 230]
    J --> K[stCaptureProcessing 231]
    K --> L[stCaptureStreamPersisted 232]
    L --> M[stCaptureCompleted 901]
    
    I --> N[stCancelPrepared 240]
    N --> O[stCancelProcessing 241]
    O --> P[stCancelStreamPersisted 242]
    P --> Q[stCancelFailed 702]
    
    D --> R[stResumePrepared 250]
    
    C --> S[stTransactionLimitCheckFailed 502]
    D --> T[stRiskCheckError 503]
    D --> U[stRiskCheckDeny 504]
    
    V[stPrepareFailureHandling 501] --> W[stFailurePublished 505]
    W --> X[stFailureNotified 510]
    
    Y[stCaptureFailed 701] --> Z[stInvestigationRequiredPublished 721]
    Z --> AA[stInvestigationRequiredNotified 722]
    
    M --> AB[stTransferCompleted 900]
    AB --> AC[stTransferCompletedAutoPublish 902]
    AC --> AD[stCompletedPublished 905]
    AD --> AE[stCompletedNotified 910]
    
    AF[stCanceled 600] --> AG[stCanceledPublished 610]
    
    AH[stValidateSuccess 800]
```

## State Categories

### Initial States (100-103)
- `stTransferPersisted` (100) - Transfer has been persisted
- `stProcessingPublished` (101) - Processing event has been published
- `stPreTransactionLimitCheck` (102) - Pre-transaction limit check in progress
- `stPreRiskCheck` (103) - Pre-risk assessment in progress

### Processing States (210-250)
- `stTransferProcessing` (210) - Transfer is being processed
- `stTransferStreamPersisted` (211) - Transfer stream has been persisted
- `stAuthProcessing` (220) - Authentication in progress
- `stAuthStreamPersisted` (221) - Authentication stream has been persisted
- `stAuthSuccess` (300) - Authentication successful
- `stCapturePrepared` (230) - Capture has been prepared
- `stCaptureProcessing` (231) - Capture is being processed
- `stCaptureStreamPersisted` (232) - Capture stream has been persisted
- `stCancelPrepared` (240) - Cancellation has been prepared
- `stCancelProcessing` (241) - Cancellation is being processed
- `stCancelStreamPersisted` (242) - Cancellation stream has been persisted
- `stResumePrepared` (250) - Resume has been prepared
- `stValidateSuccess` (800) - Validation successful

### Failure Handling States (501-510)
- `stPrepareFailureHandling` (501) - Asynchronous flow failure handling
- `stTransactionLimitCheckFailed` (502) - Transaction limit check failed
- `stRiskCheckError` (503) - Risk check service internal error
- `stRiskCheckDeny` (504) - Risk check returned Deny response
- `stFailurePublished` (505) - Failure has been published
- `stFailureNotified` (510) - Failure has been notified

### Cancellation States (600-610)
- `stCanceled` (600) - Transfer has been canceled
- `stCanceledPublished` (610) - Cancellation has been published

### Investigation Required States (701-722)
- `stCaptureFailed` (701) - Capture failed (investigation required)
- `stCancelFailed` (702) - Cancel failed (investigation required)
- `stInvestigationRequiredPublished` (721) - Investigation required has been published
- `stInvestigationRequiredNotified` (722) - Investigation required has been notified

### Completion States (900-910)
- `stTransferCompleted` (900) - Transfer has been completed
- `stCaptureCompleted` (901) - Capture has been completed
- `stTransferCompletedAutoPublish` (902) - Transfer completion auto-published
- `stCompletedPublished` (905) - Completion has been published
- `stCompletedNotified` (910) - Completion has been notified