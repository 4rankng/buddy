package ports

import "time"

// StoragePort defines the interface for lightweight file-based storage operations
type StoragePort interface {
	// Basic storage operations
	Store(key string, data interface{}) error
	Retrieve(key string) (interface{}, error)
	Delete(key string) error
	Exists(key string) (bool, error)

	// List operations
	List(prefix string) ([]string, error)
	ListWithMetadata(prefix string) ([]StorageItem, error)

	// Advanced operations
	StoreWithTTL(key string, data interface{}, ttl time.Duration) error
	StoreWithMetadata(key string, data interface{}, metadata map[string]string) error

	// Batch operations
	StoreBatch(items map[string]interface{}) error
	RetrieveBatch(keys []string) (map[string]interface{}, error)

	// Query operations
	Query(criteria *QueryCriteria) ([]StorageItem, error)

	// Maintenance operations
	Cleanup() error
	Compact() error

	// Health check
	HealthCheck() error
}

// StorageItem represents a stored item with metadata
type StorageItem struct {
	Key       string                 `json:"key"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]string      `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	ExpiresAt *time.Time             `json:"expires_at,omitempty"`
	Size      int64                  `json:"size"`
}

// QueryCriteria defines criteria for querying stored items
type QueryCriteria struct {
	Prefix      string            `json:"prefix,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	CreatedAfter  *time.Time      `json:"created_after,omitempty"`
	CreatedBefore *time.Time      `json:"created_before,omitempty"`
	ExpiresAfter  *time.Time      `json:"expires_after,omitempty"`
	ExpiresBefore *time.Time      `json:"expires_before,omitempty"`
	Limit        int               `json:"limit,omitempty"`
	Offset       int               `json:"offset,omitempty"`
	SortBy       string            `json:"sort_by,omitempty"` // key, created_at, updated_at, size
	SortOrder    string            `json:"sort_order,omitempty"` // asc, desc
}

// StorageConfig contains configuration for the storage module
type StorageConfig struct {
	BasePath     string        `json:"base_path"`
	MaxSize      int64         `json:"max_size"`       // Maximum storage size in bytes
	MaxFileSize  int64         `json:"max_file_size"` // Maximum file size in bytes
	Compression  bool          `json:"compression"`   // Enable compression
	Encryption   bool          `json:"encryption"`    // Enable encryption
	DefaultTTL   time.Duration `json:"default_ttl"`   // Default TTL for items
	CleanupInterval time.Duration `json:"cleanup_interval"` // Cleanup interval
}

// StorageStats provides statistics about the storage
type StorageStats struct {
	TotalItems   int64         `json:"total_items"`
	TotalSize    int64         `json:"total_size"`
	UsedSpace    int64         `json:"used_space"`
	FreeSpace    int64         `json:"free_space"`
	ExpiredItems int64         `json:"expired_items"`
	LastCleanup  time.Time     `json:"last_cleanup"`
}

// IndexDefinition defines an index for faster queries
type IndexDefinition struct {
	Name   string   `json:"name"`
	Fields []string `json:"fields"`
	Type   string   `json:"type"` // hash, range, text
}