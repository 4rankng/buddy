package clients

import (
	"time"

	"buddy/config"
)

// NewDoormanClient creates a Doorman client using the appropriate environment
// Now returns the singleton DoormanClient
func NewDoormanClient(timeout time.Duration) (DoormanInterface, error) {
	env := config.GetEnvironment()
	return GetDoormanClient(env)
}
