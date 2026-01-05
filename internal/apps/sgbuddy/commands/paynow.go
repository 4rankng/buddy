package sgbuddy

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"buddy/internal/apps/common"
	"buddy/internal/clients/jira"
	"buddy/internal/di"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func NewPayNowCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "paynow",
		Short: "PayNow operations",
	}

	cmd.AddCommand(NewPayNowUnlinkCmd(appCtx, clients))
	return cmd
}

func NewPayNowUnlinkCmd(appCtx *common.Context, clients *di.ClientSet) *cobra.Command {
	var safeID string
	var shiprm bool

	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Unlink PayNow for a SafeID",
		Long:  "Trigger unlink (deregistration) flow for a user SafeID in Pairing Service.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if safeID == "" {
				return errors.New("missing --safe-id")
			}

			// 1. Query Doorman
			fmt.Printf("Querying Doorman for SafeID: %s...\n", safeID)

			// We need to access the pairing service directly.
			// The Doorman client in buddy exposes generic query methods.
			// However, QueryPairingService is not explicitly in the interface method list we saw,
			// but we can check if we can cast or just use ExecuteQuery equivalent if available,
			// OR we assume the Doorman client has it (the interface had QueryPaymentEngine etc).
			// Let's check if there is a generic ExecuteQuery exposed or we add a specific one.
			// Actually, the interface `DoormanInterface` had `QueryPaymentEngine`, `QueryPaymentCore`, etc.
			// usage: clients.Doorman.QueryPaymentEngine(...)
			// But for Pairing Service, we need to know if `QueryPairingService` exists on the client.
			// I will assume likely NOT based on the interface file I read earlier (lines 1-36 of doorman_interface.go didn't show it).
			// Wait, I read `doorman_interface.go` step 68, it has:
			// QueryPaymentEngine, QueryPaymentCore, QueryFastAdapter, QueryRppAdapter, QueryPartnerpayEngine.
			// It DOES NOT have QueryPairingService.
			// I should use `ExecuteQuery` which IS in the interface: `ExecuteQuery(cluster, instance, schema, query string)`

			query := fmt.Sprintf("SELECT registration_id, status FROM pay_now_account where user_id='%s' and status='LINKED'", safeID)

			// Cluster/Instance details from oncall-app:
			// "sg-prd-m-pairing-service", "sg-prd-m-pairing-service", "prod_pairing_service_db01"
			rows, err := clients.Doorman.ExecuteQuery(
				"sg-prd-m-pairing-service",
				"sg-prd-m-pairing-service",
				"prod_pairing_service_db01",
				query,
			)
			if err != nil {
				return fmt.Errorf("doorman query failed: %w", err)
			}

			if len(rows) == 0 {
				fmt.Printf("No PayNow account found for SafeID: %s\n", safeID)
				return nil
			}

			// Assuming Doorman returns map[string]interface{}.
			// Keys might be uppercase or lowercase depending on how `ExecuteQuery` handles headers.
			// oncall-app client seemed to preserve them. Buddy's DoormanClient probably does too.
			// I'll check first row.
			r := rows[0]

			// Handle potential case sensitivity or type assertion
			var reg, status string
			if v, ok := r["registration_id"].(string); ok {
				reg = v
			}
			if v, ok := r["status"].(string); ok {
				status = v
			}

			if reg == "" || status == "" {
				fmt.Printf("Invalid PayNow account data for SafeID: %s (Raw: %v)\n", safeID, r)
				return nil
			}

			if status == "UNLINKED" {
				fmt.Printf("PayNow account for SafeID %s is already UNLINKED\n", safeID)
				return nil
			}

			curl := generateCurl(safeID, reg)

			if !shiprm {
				fmt.Printf("Prepared PayNow deregistration for SafeID %s.\n", safeID)
				fmt.Println("Run:")
				fmt.Println(curl)
				return nil
			}

			// 2. Create SHIPRM Ticket
			fmt.Println("Creating SHIPRM ticket...")

			jiraConfig := jira.GetJiraConfig("sg") // Force SG for now as this is sgbuddy
			if jiraConfig.Auth.Username == "" || jiraConfig.Auth.APIKey == "" {
				return errors.New("JIRA credentials not found in env")
			}

			title := fmt.Sprintf("[pairing-service] Deregister PayNow For SafeID: %s", safeID)
			desc := fmt.Sprintf("Delink PayNow account for user SafeID: %s, Registration ID: %s", safeID, reg)

			serviceObj := []map[string]any{{
				"workspaceId": "246c1bd0-99bf-42c1-b124-a30c70c816d1",
				"id":          "246c1bd0-99bf-42c1-b124-a30c70c816d1:232", // 232 is pairing-service
				"objectId":    "232",
			}}

			// "319" is default change type CURL, per oncall-app
			err = createShiprmTicket(jiraConfig, title, desc, curl, "319", serviceObj)
			if err != nil {
				return fmt.Errorf("failed to create SHIPRM ticket: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&safeID, "safe-id", "", "User SafeID")
	cmd.Flags().BoolVar(&shiprm, "shiprm", false, "Create SHIPRM ticket")
	return cmd
}

func generateCurl(safeID, registrationID string) string {
	// Hardcoded base URL from oncall-app default implies we might want to config it
	// But for now I'll stick to the logic provided.
	url := fmt.Sprintf("https://debug.sgbank.pr/pairing-service/v1/proxy/%s", registrationID)
	idempotency := uuid.New().String()
	return fmt.Sprintf("curl --location --request DELETE '%s' --header 'X-Grab-Id-Userid: %s' --header 'Content-Type: application/json' --data '{\"idempotencyKey\": \"%s\"}'", url, safeID, idempotency)
}

// --- Jira Helpers duplicated from oncall-app because generic client is insufficient ---

type adfDoc struct {
	Version int              `json:"version"`
	Type    string           `json:"type"`
	Content []map[string]any `json:"content"`
}

type jiraIssueReq struct {
	Fields map[string]any `json:"fields"`
}

type jiraIssueResp struct {
	Key string `json:"key"`
}

func adf(text string) adfDoc {
	return adfDoc{Version: 1, Type: "doc", Content: []map[string]any{{
		"type":    "paragraph",
		"content": []map[string]any{{"type": "text", "text": text}},
	}}}
}

func adfCode(text string) adfDoc {
	return adfDoc{Version: 1, Type: "doc", Content: []map[string]any{{
		"type":    "codeBlock",
		"attrs":   map[string]any{"language": "bash"},
		"content": []map[string]any{{"type": "text", "text": text}},
	}}}
}

func createShiprmTicket(cfg jira.JiraConfig, title, description, curl string, changeType string, serviceObj []map[string]any) error {
	baseURL := os.Getenv("JIRA_BASE_URL")
	if baseURL == "" {
		baseURL = "https://gxsbank.atlassian.net"
	}
	u := baseURL + "/rest/api/3/issue"

	fields := map[string]any{
		"project":           map[string]any{"key": "SHIPRM"},
		"summary":           title,
		"description":       adf(description),
		"customfield_11290": []map[string]any{{"workspaceId": "246c1bd0-99bf-42c1-b124-a30c70c816d1", "id": "246c1bd0-99bf-42c1-b124-a30c70c816d1:" + changeType, "objectId": changeType}},
		"customfield_10925": adf("NA"),
		"customfield_11181": serviceObj,
		"customfield_11183": time.Now().Format("2006-01-02"),
		"customfield_11187": adfCode(curl),
		"customfield_10042": adf("NA"),
		"customfield_11188": adf("Execute curl to deregister PayNow for requested accounts"),
		"customfield_11189": adf("No"),
		"customfield_11190": adf("No"),
		"customfield_11191": adf("No"),
		"customfield_11186": adf("Similar to SHIPRM-6187"),
		"customfield_11192": adf("Validate paynow is unlinked"),
		"issuetype":         map[string]any{"id": "10005"},
	}

	body := jiraIssueReq{Fields: fields}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	if err != nil {
		return err
	}

	req.SetBasicAuth(cfg.Auth.Username, cfg.Auth.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		rb, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("JIRA create failed: %s (%s)", resp.Status, string(rb))
	}

	var out jiraIssueResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}

	fmt.Printf("Created SHIPRM ticket: %s/browse/%s\n", baseURL, out.Key)

	// Add comment
	return addComment(client, baseURL, cfg, out.Key)
}

func addComment(client *http.Client, baseURL string, cfg jira.JiraConfig, issueKey string) error {
	u := baseURL + "/rest/api/3/issue/" + issueKey + "/comment"

	// Hardcoded comment body structure from oncall-app
	comment := map[string]any{
		"body": map[string]any{
			"version": 1,
			"type":    "doc",
			"content": []any{
				map[string]any{
					"type": "paragraph",
					"content": []any{
						map[string]any{"type": "text", "text": "Generated by sgbuddy paynow unlink"},
					},
				},
			},
		},
		"properties": []any{
			map[string]any{
				"key": "sd.public.comment",
				"value": map[string]any{
					"internal": true,
				},
			},
		},
	}

	b, _ := json.Marshal(comment)
	req, _ := http.NewRequest(http.MethodPost, u, bytes.NewReader(b))
	req.SetBasicAuth(cfg.Auth.Username, cfg.Auth.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != 201 {
		// Log but don't fail hard if comment fails
		fmt.Printf("Warning: Failed to add comment: %s\n", resp.Status)
	}
	return nil
}
