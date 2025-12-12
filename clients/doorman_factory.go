package clients

import (
	"time"

	"buddy/config"
)

// DoormanStrategy defines the interface for different Doorman configurations
type DoormanStrategy interface {
	GetBaseURL() string
	GetUsernameKey() string
	GetPasswordKey() string
	GetDatabaseCluster(service string) string
}

// MalaysiaDoormanStrategy implements the strategy for Malaysia
type MalaysiaDoormanStrategy struct{}

func (s *MalaysiaDoormanStrategy) GetBaseURL() string {
	return config.Get("DOORMAN_BASE_URL", "https://doorman.infra.prd.g-bank.app")
}

func (s *MalaysiaDoormanStrategy) GetUsernameKey() string {
	return "DOORMAN_USERNAME"
}

func (s *MalaysiaDoormanStrategy) GetPasswordKey() string {
	return "DOORMAN_PASSWORD"
}

func (s *MalaysiaDoormanStrategy) GetDatabaseCluster(service string) string {
	switch service {
	case "payment-engine":
		return "prd-payments-payment-engine-rds-mysql"
	case "payment-core":
		return "prd-payments-payment-core-rds-mysql"
	case "rpp-adapter":
		return "prd-payments-rpp-adapter-rds-mysql"
	case "partnerpay-engine":
		return "prd-payments-partnerpay-engine-rds-mysql"
	default:
		return service
	}
}

// SingaporeDoormanStrategy implements the strategy for Singapore
type SingaporeDoormanStrategy struct{}

func (s *SingaporeDoormanStrategy) GetBaseURL() string {
	return config.Get("DOORMAN_SG_BASE_URL", "https://doorman.sgbank.pr/rds/query")
}

func (s *SingaporeDoormanStrategy) GetUsernameKey() string {
	return "DOORMAN_SG_USERNAME"
}

func (s *SingaporeDoormanStrategy) GetPasswordKey() string {
	return "DOORMAN_SG_PASSWORD"
}

func (s *SingaporeDoormanStrategy) GetDatabaseCluster(service string) string {
	switch service {
	case "payment-engine":
		return "sg-prd-m-payment-engine"
	case "payment-core":
		return "sg-prd-m-payment-core"
	case "fast-adapter":
		return "sg-prd-m-fast-adapter"
	default:
		return service
	}
}

// DoormanClientFactory creates Doorman clients based on the environment
type DoormanClientFactory struct {
	strategy DoormanStrategy
}

// NewDoormanClientFactory creates a new factory with the appropriate strategy
func NewDoormanClientFactory(env string) *DoormanClientFactory {
	var strategy DoormanStrategy
	if env == "sg" {
		strategy = &SingaporeDoormanStrategy{}
	} else {
		strategy = &MalaysiaDoormanStrategy{}
	}
	return &DoormanClientFactory{strategy: strategy}
}

// CreateClient creates a new Doorman client using the factory's strategy
func (f *DoormanClientFactory) CreateClient(timeout time.Duration) (*DoormanClient, error) {
	baseURL := f.strategy.GetBaseURL()
	username := config.Get(f.strategy.GetUsernameKey(), "")
	password := config.Get(f.strategy.GetPasswordKey(), "")

	return NewDoormanClientWithConfig(baseURL, username, password, timeout)
}

// QueryDatabase executes a query against the specified database service
func (f *DoormanClientFactory) QueryDatabase(service, schema, query string) ([]map[string]interface{}, error) {
	client, err := f.CreateClient(30 * time.Second)
	if err != nil {
		return nil, err
	}

	cluster := f.strategy.GetDatabaseCluster(service)
	return client.ExecuteQuery(cluster, cluster, schema, query)
}
