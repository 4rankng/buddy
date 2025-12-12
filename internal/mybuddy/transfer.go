package mybuddy

import (
	"encoding/json"
	"fmt"
	"os"

	"buddy/clients"
	"buddy/output"
	"github.com/spf13/cobra"
)

// TransferCmd represents the transfer command
var TransferCmd = &cobra.Command{
	Use:   "transfer",
	Short: "Query transfer table",
	Long:  `Query the transfer table from the payment engine database`,
	Run:   runTransfer,
}

// TransferResult represents the structured response from the transfer query
type TransferResult struct {
	Code    int    `json:"code"`
	Errors  string `json:"errors"`
	Message string `json:"message"`
	Result  struct {
		Headers []string        `json:"headers"`
		Types   []string        `json:"types"`
		Rows    [][]interface{} `json:"rows"`
	} `json:"result"`
	RequestID string `json:"requestID"`
}

func runTransfer(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Println("Error: query argument is required")
		fmt.Println("Usage: mybuddy transfer \"SELECT * FROM transfer LIMIT 10\"")
		os.Exit(1)
	}

	query := args[0]

	// Create doorman client
	client, err := clients.NewDoormanClient(30)
	if err != nil {
		fmt.Printf("Error creating doorman client: %v\n", err)
		os.Exit(1)
	}

	// Execute transfer query using payment engine
	data, err := client.QueryPaymentEngine(query)
	if err != nil {
		fmt.Printf("Error executing transfer query: %v\n", err)
		os.Exit(1)
	}

	// Convert to structured result for consistent output
	result := TransferResult{
		Code:    200,
		Errors:  "",
		Message: "",
		RequestID: "transfer-query",
	}
	
	// Populate result with actual data
	if len(data) > 0 {
		// Extract headers from first row keys
		var headers []string
		for key := range data[0] {
			headers = append(headers, key)
		}
		
		// Convert data to rows
		var rows [][]interface{}
		for _, row := range data {
			var rowValues []interface{}
			for _, header := range headers {
				rowValues = append(rowValues, row[header])
			}
			rows = append(rows, rowValues)
		}
		
		result.Result.Headers = headers
		result.Result.Rows = rows
	}

	// Output the result
	outputJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling output: %v\n", err)
		os.Exit(1)
	}

	output.PrintJSON(string(outputJSON))
}