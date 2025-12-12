package output

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func LogEvent(event string, data map[string]any) {
	// Filter out authentication logs
	if event == "doorman_auth_attempt" {
		return
	}

	logEntry := map[string]any{
		"event": event,
		"data":  data,
	}
	if jsonBytes, err := json.Marshal(logEntry); err == nil {
		log.Println(string(jsonBytes))
	}
}

func PrintJSON(data interface{}) {
	if jsonBytes, err := json.MarshalIndent(data, "", "  "); err == nil {
		fmt.Println(string(jsonBytes))
	} else {
		fmt.Printf("Error: %v\n", err)
	}
}

func PrintError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}

func PrintMessage(msg string) {
	fmt.Println(msg)
}
