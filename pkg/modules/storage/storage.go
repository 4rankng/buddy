package storage

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// storageModule implements the StoragePort interface
type storageModule struct {
	config      config.StorageConfig
	basePath    string
	compression bool
	encryption  bool
	gcm         cipher.AEAD
}

// storageFile represents a stored file with metadata
type storageFile struct {
	Key        string                 `json:"key"`
	Data       interface{}            `json:"data"`
	Metadata   map[string]string      `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Size       int64                  `json:"size"`
	Compressed bool                   `json:"compressed"`
	Encrypted  bool                   `json:"encrypted"`
}

// NewStorageModule creates a new Storage module (using SQLite for better performance)
func NewStorageModule(cfg config.StorageConfig) (ports.StoragePort, error) {
	return NewSQLiteStorage(cfg)
}

// Store stores data with the given key
func (s *storageModule) Store(key string, data interface{}) error {
	return s.StoreWithMetadata(key, data, nil)
}

// StoreWithMetadata stores data with metadata
func (s *storageModule) StoreWithMetadata(key string, data interface{}, metadata map[string]string) error {
	if err := s.validateKey(key); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	file := storageFile{
		Key:       key,
		Data:      data,
		Metadata:  make(map[string]string),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Size:      s.calculateSize(data),
	}

	// Copy metadata
	if metadata != nil {
		for k, v := range metadata {
			file.Metadata[k] = v
		}
	}

	filePath := s.getFilePath(key)

	// Read existing file to preserve CreatedAt
	if existingData, err := s.readFile(filePath); err == nil {
		file.CreatedAt = existingData.CreatedAt
	}

	// Serialize data
	jsonData, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Compress if enabled
	if s.compression {
		jsonData, err = s.compress(jsonData)
		if err != nil {
			return fmt.Errorf("failed to compress data: %w", err)
		}
		file.Compressed = true
	}

	// Encrypt if enabled
	if s.encryption {
		jsonData, err = s.encrypt(jsonData)
		if err != nil {
			return fmt.Errorf("failed to encrypt data: %w", err)
		}
		file.Encrypted = true
	}

	// Write to temporary file first
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	// Atomic rename
	if err := os.Rename(tempPath, filePath); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to rename file: %w", err)
	}

	return nil
}

// StoreWithTTL stores data with a time-to-live
func (s *storageModule) StoreWithTTL(key string, data interface{}, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	metadata := map[string]string{
		"ttl": ttl.String(),
	}

	return s.StoreWithExpiration(key, data, &expiresAt, metadata)
}

// StoreWithExpiration stores data with an explicit expiration time
func (s *storageModule) StoreWithExpiration(key string, data interface{}, expiresAt *time.Time, metadata map[string]string) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	if expiresAt != nil {
		metadata["expires_at"] = expiresAt.Format(time.RFC3339)
	}

	if err := s.StoreWithMetadata(key, data, metadata); err != nil {
		return err
	}

	// Update the file with expiration
	filePath := s.getFilePath(key)
	fileData, err := s.readFile(filePath)
	if err != nil {
		return err
	}

	fileData.ExpiresAt = expiresAt
	fileData.UpdatedAt = time.Now()

	return s.writeFile(filePath, fileData)
}

// Retrieve retrieves data by key
func (s *storageModule) Retrieve(key string) (interface{}, error) {
	filePath := s.getFilePath(key)

	fileData, err := s.readFile(filePath)
	if err != nil {
		return nil, err
	}

	// Check if expired
	if fileData.ExpiresAt != nil && time.Now().After(*fileData.ExpiresAt) {
		_ = os.Remove(filePath)
		return nil, fmt.Errorf("data has expired")
	}

	return fileData.Data, nil
}

// Delete deletes data by key
func (s *storageModule) Delete(key string) error {
	filePath := s.getFilePath(key)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// Exists checks if data exists for the given key
func (s *storageModule) Exists(key string) (bool, error) {
	filePath := s.getFilePath(key)

	fileData, err := s.readFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	// Check if expired
	if fileData.ExpiresAt != nil && time.Now().After(*fileData.ExpiresAt) {
		_ = os.Remove(filePath)
		return false, nil
	}

	return true, nil
}

// List lists keys with the given prefix
func (s *storageModule) List(prefix string) ([]string, error) {
	items, err := s.ListWithMetadata(prefix)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(items))
	for _, item := range items {
		keys = append(keys, item.Key)
	}

	return keys, nil
}

// ListWithMetadata lists keys and metadata with the given prefix
func (s *storageModule) ListWithMetadata(prefix string) ([]ports.StorageItem, error) {
	prefixPath := s.getFilePath(prefix)
	dir := filepath.Dir(prefixPath)
	basePrefix := filepath.Base(prefixPath)

	var items []ports.StorageItem

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Check if file matches prefix
		if !strings.HasPrefix(info.Name(), basePrefix) {
			return nil
		}

		// Skip temporary files
		if strings.HasSuffix(info.Name(), ".tmp") {
			return nil
		}

		// Read file data
		fileData, err := s.readFile(path)
		if err != nil {
			// Skip files that can't be read (might be expired)
			return nil
		}

		// Check if expired
		if fileData.ExpiresAt != nil && time.Now().After(*fileData.ExpiresAt) {
			_ = os.Remove(path)
			return nil
		}

		// Convert to StorageItem
		item := ports.StorageItem{
			Key:       fileData.Key,
			Data:      fileData.Data,
			Metadata:  fileData.Metadata,
			CreatedAt: fileData.CreatedAt,
			UpdatedAt: fileData.UpdatedAt,
			ExpiresAt: fileData.ExpiresAt,
			Size:      fileData.Size,
		}

		items = append(items, item)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	return items, nil
}

// StoreBatch stores multiple items
func (s *storageModule) StoreBatch(items map[string]interface{}) error {
	for key, data := range items {
		if err := s.Store(key, data); err != nil {
			return fmt.Errorf("failed to store batch item %s: %w", key, err)
		}
	}
	return nil
}

// RetrieveBatch retrieves multiple items
func (s *storageModule) RetrieveBatch(keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, key := range keys {
		data, err := s.Retrieve(key)
		if err != nil {
			// Skip missing items, but don't fail the entire operation
			continue
		}
		result[key] = data
	}

	return result, nil
}

// Query performs a complex query
func (s *storageModule) Query(criteria *ports.QueryCriteria) ([]ports.StorageItem, error) {
	// Get all items with prefix
	items, err := s.ListWithMetadata(criteria.Prefix)
	if err != nil {
		return nil, err
	}

	// Filter items
	var filtered []ports.StorageItem
	for _, item := range items {
		if s.matchesCriteria(item, criteria) {
			filtered = append(filtered, item)
		}
	}

	// Sort items
	if criteria.SortBy != "" {
		filtered = s.sortItems(filtered, criteria.SortBy, criteria.SortOrder)
	}

	// Apply limit and offset
	if criteria.Offset >= len(filtered) {
		return []ports.StorageItem{}, nil
	}

	end := len(filtered)
	if criteria.Limit > 0 && criteria.Offset+criteria.Limit < end {
		end = criteria.Offset + criteria.Limit
	}

	return filtered[criteria.Offset:end], nil
}

// Cleanup removes expired files
func (s *storageModule) Cleanup() error {
	return filepath.Walk(s.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Skip temporary files
		if strings.HasSuffix(info.Name(), ".tmp") {
			return nil
		}

		fileData, err := s.readFile(path)
		if err != nil {
			// Remove files that can't be read
			_ = os.Remove(path)
			return nil
		}

		// Check if expired
		if fileData.ExpiresAt != nil && time.Now().After(*fileData.ExpiresAt) {
			_ = os.Remove(path)
		}

		return nil
	})
}

// Compact optimizes storage by removing old files and cleaning up
func (s *storageModule) Compact() error {
	// First, cleanup expired files
	if err := s.Cleanup(); err != nil {
		return fmt.Errorf("cleanup failed during compaction: %w", err)
	}

	// Add more compaction logic here if needed
	// For example: reorganize files, consolidate small files, etc.

	return nil
}

// HealthCheck performs a health check on the storage
func (s *storageModule) HealthCheck() error {
	// Try to create a test file
	testKey := ".health-check-" + time.Now().Format("20060102-150405")
	testData := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "ok",
	}

	if err := s.Store(testKey, testData); err != nil {
		return fmt.Errorf("failed to write health check file: %w", err)
	}

	// Try to read it back
	if _, err := s.Retrieve(testKey); err != nil {
		return fmt.Errorf("failed to read health check file: %w", err)
	}

	// Clean up
	_ = s.Delete(testKey)

	return nil
}

// Helper methods

func (s *storageModule) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if strings.Contains(key, "..") {
		return fmt.Errorf("key cannot contain '..'")
	}

	if strings.HasPrefix(key, "/") || strings.HasPrefix(key, "\\") {
		return fmt.Errorf("key cannot start with '/' or '\\'")
	}

	return nil
}

func (s *storageModule) getFilePath(key string) string {
	// Replace slashes with dashes to avoid directory traversal
	safeKey := strings.ReplaceAll(key, "/", "-")
	return filepath.Join(s.basePath, safeKey+".json")
}

func (s *storageModule) readFile(path string) (*storageFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Decrypt if needed
	if s.encryption {
		data, err = s.decrypt(data)
		if err != nil {
			return nil, err
		}
	}

	// Decompress if needed
	if s.compression {
		data, err = s.decompress(data)
		if err != nil {
			return nil, err
		}
	}

	var file storageFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, err
	}

	return &file, nil
}

func (s *storageModule) writeFile(path string, file *storageFile) error {
	data, err := json.Marshal(file)
	if err != nil {
		return err
	}

	// Compress if enabled
	if s.compression {
		data, err = s.compress(data)
		if err != nil {
			return err
		}
	}

	// Encrypt if enabled
	if s.encryption {
		data, err = s.encrypt(data)
		if err != nil {
			return err
		}
	}

	return os.WriteFile(path, data, 0644)
}

func (s *storageModule) calculateSize(data interface{}) int64 {
	jsonData, _ := json.Marshal(data)
	return int64(len(jsonData))
}

func (s *storageModule) startCleanupRoutine() {
	ticker := time.NewTicker(s.config.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.Cleanup()
	}
}

// Placeholder implementations for compression and encryption
func (s *storageModule) compress(data []byte) ([]byte, error) {
	// TODO: Implement compression (e.g., using gzip)
	return data, nil
}

func (s *storageModule) decompress(data []byte) ([]byte, error) {
	// TODO: Implement decompression
	return data, nil
}

func (s *storageModule) encrypt(data []byte) ([]byte, error) {
	if s.gcm == nil {
		return nil, fmt.Errorf("encryption not initialized")
	}

	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return s.gcm.Seal(nonce, nonce, data, nil), nil
}

func (s *storageModule) decrypt(data []byte) ([]byte, error) {
	if s.gcm == nil {
		return nil, fmt.Errorf("encryption not initialized")
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return s.gcm.Open(nil, nonce, ciphertext, nil)
}

func (s *storageModule) matchesCriteria(item ports.StorageItem, criteria *ports.QueryCriteria) bool {
	// Check metadata matches
	for key, value := range criteria.Tags {
		if itemValue, ok := item.Metadata[key]; !ok || itemValue != value {
			return false
		}
	}

	// Check time ranges
	if criteria.CreatedAfter != nil && item.CreatedAt.Before(*criteria.CreatedAfter) {
		return false
	}

	if criteria.CreatedBefore != nil && item.CreatedAt.After(*criteria.CreatedBefore) {
		return false
	}

	if criteria.ExpiresAfter != nil && (item.ExpiresAt == nil || item.ExpiresAt.Before(*criteria.ExpiresAfter)) {
		return false
	}

	if criteria.ExpiresBefore != nil && (item.ExpiresAt != nil && item.ExpiresAt.After(*criteria.ExpiresBefore)) {
		return false
	}

	return true
}

func (s *storageModule) sortItems(items []ports.StorageItem, sortBy, sortOrder string) []ports.StorageItem {
	// TODO: Implement sorting logic
	// For now, return items as-is
	return items
}