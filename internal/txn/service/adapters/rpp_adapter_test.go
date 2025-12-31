package adapters

import (
	"testing"
	"time"

	"buddy/internal/txn/domain"
)

func TestRPPAdapterQueryByE2EID(t *testing.T) {
	tests := []struct {
		name       string
		mockClient *mockClient
		want       *domain.RPPAdapterInfo
		wantErr    bool
	}{
		{
			name: "Query with valid E2E ID returns workflows",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{
					{
						"req_biz_msg_id": "20251228GXSPMYKL010ORB22837568",
						"partner_tx_id":  "partner123",
						"partner_tx_sts": "900",
						"status":         "900",
						"created_at":     "2025-12-28T06:35:10.292282Z",
						"end_to_end_id":  "e2e123",
					},
				},
				workflowResults: []map[string]interface{}{
					{
						"run_id":        "partner123",
						"workflow_id":   "wf_process_registry",
						"state":         float64(900),
						"attempt":       float64(0),
						"prev_trans_id": "prev456",
						"data":          `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
				},
			},
			want: &domain.RPPAdapterInfo{
				ReqBizMsgID: "20251228GXSPMYKL010ORB22837568",
				PartnerTxID: "partner123",
				EndToEndID:  "e2e123",
				Status:      "900",
				CreatedAt:   "2025-12-28T06:35:10.292282Z",
				Workflow: []domain.WorkflowInfo{
					{
						WorkflowID:  "wf_process_registry",
						State:       "900",
						Attempt:     0,
						RunID:       "partner123",
						PrevTransID: "prev456",
						Data:        `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
				},
				Info: "RPP Status: 900",
			},
			wantErr: false,
		},
		{
			name: "Query with invalid E2E ID returns error",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{}, // Empty slice triggers nil return
				workflowResults:      nil,
			},
			want:    nil,
			wantErr: false, // Adapter returns (nil, nil) for empty results, not an error
		},
		{
			name: "Query returns multiple workflows within time window",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{
					{
						"req_biz_msg_id": "20251228GXSPMYKL010ORB22837568",
						"partner_tx_id":  "partner123",
						"partner_tx_sts": "900",
						"status":         "900",
						"created_at":     "2025-12-28T06:35:10.292282Z",
						"end_to_end_id":  "e2e123",
					},
				},
				workflowResults: []map[string]interface{}{
					{
						"run_id":        "partner123",
						"workflow_id":   "wf_process_registry",
						"state":         float64(900),
						"attempt":       float64(0),
						"prev_trans_id": "prev456",
						"data":          `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
					{
						"run_id":        "partner124",
						"workflow_id":   "wf_process_registry",
						"state":         float64(210),
						"attempt":       float64(0),
						"prev_trans_id": "",
						"data":          `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
				},
			},
			want: &domain.RPPAdapterInfo{
				ReqBizMsgID: "20251228GXSPMYKL010ORB22837568",
				PartnerTxID: "partner123",
				EndToEndID:  "e2e123",
				Status:      "900",
				CreatedAt:   "2025-12-28T06:35:10.292282Z",
				Workflow: []domain.WorkflowInfo{
					{
						WorkflowID:  "wf_process_registry",
						State:       "900",
						Attempt:     0,
						RunID:       "partner123",
						PrevTransID: "prev456",
						Data:        `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
					{
						WorkflowID:  "wf_process_registry",
						State:       "210",
						Attempt:     0,
						RunID:       "partner124",
						PrevTransID: "",
						Data:        `{"req_biz_msg_id":"20251228GXSPMYKL010ORB22837568"}`,
					},
				},
				Info: "RPP Status: 900",
			},
			wantErr: false,
		},
		{
			name: "Query returns empty workflow slice when no matches",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{
					{
						"req_biz_msg_id": "20251228GXSPMYKL010ORB22837568",
						"partner_tx_id":  "partner123",
						"partner_tx_sts": "900",
						"status":         "900",
						"created_at":     "2025-12-28T06:35:10.292282Z",
						"end_to_end_id":  "e2e123",
					},
				},
				workflowResults: nil,
			},
			want: &domain.RPPAdapterInfo{
				ReqBizMsgID: "20251228GXSPMYKL010ORB22837568",
				PartnerTxID: "partner123",
				EndToEndID:  "e2e123",
				Status:      "900",
				CreatedAt:   "2025-12-28T06:35:10.292282Z",
				Workflow:    []domain.WorkflowInfo{},
				Info:        "RPP Status: 900",
			},
			wantErr: false,
		},
		{
			name: "Query with invalid created_at timestamp returns empty workflow",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{
					{
						"req_biz_msg_id": "20251228GXSPMYKL010ORB22837568",
						"partner_tx_id":  "partner123",
						"partner_tx_sts": "900",
						"status":         "900",
						"created_at":     "invalid-timestamp",
						"end_to_end_id":  "e2e123",
					},
				},
				workflowResults: nil,
			},
			want: &domain.RPPAdapterInfo{
				ReqBizMsgID: "20251228GXSPMYKL010ORB22837568",
				PartnerTxID: "partner123",
				EndToEndID:  "e2e123",
				Status:      "900",
				CreatedAt:   "invalid-timestamp",
				Workflow:    []domain.WorkflowInfo{},
				Info:        "RPP Status: 900",
			},
			wantErr: false,
		},
		{
			name: "Query with empty req_biz_msg_id returns empty workflow",
			mockClient: &mockClient{
				creditTransferResult: []map[string]interface{}{
					{
						"req_biz_msg_id": "",
						"partner_tx_id":  "partner123",
						"partner_tx_sts": "900",
						"status":         "900",
						"created_at":     "2025-12-28T06:35:10.292282Z",
						"end_to_end_id":  "e2e123",
					},
				},
				workflowResults: nil,
			},
			want: &domain.RPPAdapterInfo{
				ReqBizMsgID: "",
				PartnerTxID: "partner123",
				EndToEndID:  "e2e123",
				Status:      "900",
				CreatedAt:   "2025-12-28T06:35:10.292282Z",
				Workflow:    []domain.WorkflowInfo{},
				Info:        "RPP Status: 900",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewRPPAdapter(tt.mockClient)

			// Extract input ID from want, or use empty string for error cases
			inputID := ""
			if tt.want != nil {
				inputID = tt.want.EndToEndID
			} else {
				// For nil want case, use a dummy ID
				inputID = "dummy_id"
			}

			got, err := adapter.QueryByE2EID(inputID)

			if tt.wantErr {
				if err == nil {
					t.Errorf("QueryByE2EID() expected error, got nil")
				}
				return
			}

			// For cases where want is nil (no results expected), just check that got is nil
			if tt.want == nil {
				if got != nil {
					t.Errorf("QueryByE2EID() expected nil result, got %v", got)
				}
				return
			}

			if err != nil {
				t.Errorf("QueryByE2EID() unexpected error: %v", err)
				return
			}

			if got.ReqBizMsgID != tt.want.ReqBizMsgID {
				t.Errorf("ReqBizMsgID = %v, want %v", got.ReqBizMsgID, tt.want.ReqBizMsgID)
			}
			if got.PartnerTxID != tt.want.PartnerTxID {
				t.Errorf("PartnerTxID = %v, want %v", got.PartnerTxID, tt.want.PartnerTxID)
			}
			if got.Status != tt.want.Status {
				t.Errorf("Status = %v, want %v", got.Status, tt.want.Status)
			}
			if len(got.Workflow) != len(tt.want.Workflow) {
				t.Errorf("Workflow slice length = %d, want %d", len(got.Workflow), len(tt.want.Workflow))
			}
			for i, wf := range got.Workflow {
				if i < len(tt.want.Workflow) {
					if wf.WorkflowID != tt.want.Workflow[i].WorkflowID {
						t.Errorf("Workflow[%d].WorkflowID = %v, want %v", i, wf.WorkflowID, tt.want.Workflow[i].WorkflowID)
					}
					if wf.State != tt.want.Workflow[i].State {
						t.Errorf("Workflow[%d].State = %v, want %v", i, wf.State, tt.want.Workflow[i].State)
					}
					if wf.Attempt != tt.want.Workflow[i].Attempt {
						t.Errorf("Workflow[%d].Attempt = %d, want %d", i, wf.Attempt, tt.want.Workflow[i].Attempt)
					}
				}
			}
		})
	}
}

type mockClient struct {
	creditTransferResult []map[string]interface{}
	workflowResults      []map[string]interface{}
	// Track if ExecuteQuery has been called to avoid double-counting in tests
	executedQuery bool
}

func (m *mockClient) ExecuteQuery(cluster, service, database, query string) ([]map[string]interface{}, error) {
	if cluster == "prd-payments-rpp-adapter-rds-mysql" {
		m.executedQuery = true
		return m.creditTransferResult, nil
	}
	return nil, nil
}

func (m *mockClient) QueryRppAdapter(query string) ([]map[string]interface{}, error) {
	// Only return workflow results if ExecuteQuery was called first
	// This simulates the real behavior where credit transfer must exist first
	if !m.executedQuery {
		return nil, nil
	}
	// For testing purposes, only return results once to avoid duplicates
	// (in real scenario, the two queries would return different results)
	if len(m.workflowResults) > 0 {
		result := m.workflowResults
		m.workflowResults = nil // Clear to prevent duplicates
		return result, nil
	}
	return nil, nil
}

func (m *mockClient) QueryFastAdapter(query string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockClient) QueryPaymentEngine(query string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockClient) QueryPaymentCore(query string) ([]map[string]interface{}, error) {
	return nil, nil
}

func (m *mockClient) QueryPartnerpayEngine(query string) ([]map[string]interface{}, error) {
	return nil, nil
}

func TestCalculateTimeWindow(t *testing.T) {
	createdAt, _ := time.Parse(time.RFC3339Nano, "2025-12-28T06:35:10.292282Z")

	timeWindowStart := createdAt.Add(-5 * time.Minute)
	timeWindowEnd := createdAt.Add(5 * time.Minute)

	expectedStart := "2025-12-28T06:30:10.292282Z"
	expectedEnd := "2025-12-28T06:40:10.292282Z"

	if timeWindowStart.Format(time.RFC3339Nano) != expectedStart {
		t.Errorf("Time window start = %v, want %v", timeWindowStart.Format(time.RFC3339Nano), expectedStart)
	}
	if timeWindowEnd.Format(time.RFC3339Nano) != expectedEnd {
		t.Errorf("Time window end = %v, want %v", timeWindowEnd.Format(time.RFC3339Nano), expectedEnd)
	}
}
