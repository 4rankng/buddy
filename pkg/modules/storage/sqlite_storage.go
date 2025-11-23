package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // Pure Go SQLite driver

	"oncall/pkg/config"
	"oncall/pkg/ports"
)

// sqliteStorage implements the StoragePort interface using SQLite with JSON1 extension
type sqliteStorage struct {
	db    *sql.DB
	path  string
	cfg   config.StorageConfig
}

// NewSQLiteStorage creates a new SQLite-based storage module
func NewSQLiteStorage(cfg config.StorageConfig) (ports.StoragePort, error) {
	// Ensure directory exists
	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	dbPath := filepath.Join(cfg.BasePath, "oncall_storage.db")

	// Open database
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_foreign_keys=1")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &sqliteStorage{
		db:   db,
		path: dbPath,
		cfg:  cfg,
	}

	// Initialize schema
	if err := storage.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	// Start cleanup routine
	go storage.startCleanupRoutine()

	return storage, nil
}

// initSchema creates the database schema
func (s *sqliteStorage) initSchema() error {
	schema := `
		CREATE TABLE IF NOT EXISTS storage_items (
			key TEXT PRIMARY KEY,
			data TEXT NOT NULL,                    -- JSON data
			metadata TEXT,                         -- JSON metadata
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			expires_at DATETIME,
			size INTEGER DEFAULT 0,
			tags TEXT                               -- Comma-separated tags for fast lookup
		);

		-- Indexes for performance
		CREATE INDEX IF NOT EXISTS idx_storage_created_at ON storage_items(created_at);
		CREATE INDEX IF NOT EXISTS idx_storage_updated_at ON storage_items(updated_at);
		CREATE INDEX IF NOT EXISTS idx_storage_expires_at ON storage_items(expires_at);
		CREATE INDEX IF NOT EXISTS idx_storage_tags ON storage_items(tags);

		-- Trigger to update updated_at timestamp
		CREATE TRIGGER IF NOT EXISTS update_storage_timestamp
			AFTER UPDATE ON storage_items
			BEGIN
				UPDATE storage_items SET updated_at = CURRENT_TIMESTAMP WHERE key = NEW.key;
			END;
	`

	_, err := s.db.Exec(schema)
	return err
}

// Store stores data with the given key
func (s *sqliteStorage) Store(key string, data interface{}) error {
	return s.StoreWithMetadata(key, data, nil)
}

// StoreWithMetadata stores data with metadata
func (s *sqliteStorage) StoreWithMetadata(key string, data interface{}, metadata map[string]string) error {
	if err := s.validateKey(key); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	// Serialize data
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Serialize metadata
	var metadataJSON sql.NullString
	if metadata != nil && len(metadata) > 0 {
		metadataBytes, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata: %w", err)
		}
		metadataJSON.String = string(metadataBytes)
		metadataJSON.Valid = true
	}

	// Extract tags from metadata for fast lookup
	tags := s.extractTags(metadata)

	// Calculate size
	size := int64(len(dataJSON))

	query := `
		INSERT OR REPLACE INTO storage_items (key, data, metadata, size, tags)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = s.db.Exec(query, key, string(dataJSON), metadataJSON, size, tags)
	if err != nil {
		return fmt.Errorf("failed to store data: %w", err)
	}

	return nil
}

// StoreWithTTL stores data with a time-to-live
func (s *sqliteStorage) StoreWithTTL(key string, data interface{}, ttl time.Duration) error {
	expiresAt := time.Now().Add(ttl)
	metadata := map[string]string{
		"ttl": ttl.String(),
	}

	return s.StoreWithExpiration(key, data, &expiresAt, metadata)
}

// StoreWithExpiration stores data with an explicit expiration time
func (s *sqliteStorage) StoreWithExpiration(key string, data interface{}, expiresAt *time.Time, metadata map[string]string) error {
	if metadata == nil {
		metadata = make(map[string]string)
	}

	if expiresAt != nil {
		metadata["expires_at"] = expiresAt.Format(time.RFC3339)
	}

	if err := s.StoreWithMetadata(key, data, metadata); err != nil {
		return err
	}

	// Update expiration time
	var expiresAtSQL sql.NullTime
	if expiresAt != nil {
		expiresAtSQL.Time = *expiresAt
		expiresAtSQL.Valid = true
	}

	query := `UPDATE storage_items SET expires_at = ? WHERE key = ?`
	_, err := s.db.Exec(query, expiresAtSQL, key)
	if err != nil {
		return fmt.Errorf("failed to update expiration: %w", err)
	}

	return nil
}

// Retrieve retrieves data by key
func (s *sqliteStorage) Retrieve(key string) (interface{}, error) {
	query := `
		SELECT data, expires_at
		FROM storage_items
		WHERE key = ? AND (expires_at IS NULL OR expires_at > ?)
	`

	var dataStr string
	var expiresAt sql.NullTime
	err := s.db.QueryRow(query, key, time.Now()).Scan(&dataStr, &expiresAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("key not found or expired: %s", key)
		}
		return nil, fmt.Errorf("failed to retrieve data: %w", err)
	}

	var data interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return data, nil
}

// Delete deletes data by key
func (s *sqliteStorage) Delete(key string) error {
	query := `DELETE FROM storage_items WHERE key = ?`
	result, err := s.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete data: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("key not found: %s", key)
	}

	return nil
}

// Exists checks if data exists for the given key
func (s *sqliteStorage) Exists(key string) (bool, error) {
	query := `
		SELECT 1
		FROM storage_items
		WHERE key = ? AND (expires_at IS NULL OR expires_at > ?)
		LIMIT 1
	`

	var exists int
	err := s.db.QueryRow(query, key, time.Now()).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	return exists == 1, nil
}

// List lists keys with the given prefix
func (s *sqliteStorage) List(prefix string) ([]string, error) {
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
func (s *sqliteStorage) ListWithMetadata(prefix string) ([]ports.StorageItem, error) {
	query := `
		SELECT key, data, metadata, created_at, updated_at, expires_at, size
		FROM storage_items
		WHERE key LIKE ? AND (expires_at IS NULL OR expires_at > ?)
		ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, prefix+"%", time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	defer rows.Close()

	var items []ports.StorageItem
	for rows.Next() {
		item, err := s.scanStorageItem(rows)
		if err != nil {
			continue // Skip malformed rows
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

// StoreBatch stores multiple items
func (s *sqliteStorage) StoreBatch(items map[string]interface{}) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for key, data := range items {
		if err := s.Store(key, data); err != nil {
			return fmt.Errorf("failed to store batch item %s: %w", key, err)
		}
	}

	return tx.Commit()
}

// RetrieveBatch retrieves multiple items
func (s *sqliteStorage) RetrieveBatch(keys []string) (map[string]interface{}, error) {
	if len(keys) == 0 {
		return make(map[string]interface{}), nil
	}

	// Create placeholder string for IN clause
	placeholders := strings.Repeat("?,", len(keys))
	placeholders = placeholders[:len(placeholders)-1]

	query := fmt.Sprintf(`
		SELECT key, data
		FROM storage_items
		WHERE key IN (%s) AND (expires_at IS NULL OR expires_at > ?)
	`, placeholders)

	args := make([]interface{}, len(keys)+1)
	for i, key := range keys {
		args[i] = key
	}
	args[len(keys)] = time.Now()

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve batch: %w", err)
	}
	defer rows.Close()

	result := make(map[string]interface{})
	for rows.Next() {
		var key, dataStr string
		if err := rows.Scan(&key, &dataStr); err != nil {
			continue
		}

		var data interface{}
		if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
			continue
		}

		result[key] = data
	}

	return result, rows.Err()
}

// Query performs a complex query using SQLite JSON functions
func (s *sqliteStorage) Query(criteria *ports.QueryCriteria) ([]ports.StorageItem, error) {
	// Build dynamic SQL query using JSON functions
	query := `
		SELECT key, data, metadata, created_at, updated_at, expires_at, size
		FROM storage_items
		WHERE 1=1
	`

	args := []interface{}{}

	// Add prefix filter
	if criteria.Prefix != "" {
		query += " AND key LIKE ?"
		args = append(args, criteria.Prefix+"%")
	}

	// Filter out expired items
	query += " AND (expires_at IS NULL OR expires_at > ?)"
	args = append(args, time.Now())

	// Add metadata tag filters using JSON functions
	for key, value := range criteria.Tags {
		query += " AND json_extract(metadata, '$." + key + "') = ?"
		args = append(args, value)
	}

	// Add time range filters
	if criteria.CreatedAfter != nil {
		query += " AND created_at >= ?"
		args = append(args, criteria.CreatedAfter)
	}

	if criteria.CreatedBefore != nil {
		query += " AND created_at <= ?"
		args = append(args, criteria.CreatedBefore)
	}

	if criteria.ExpiresAfter != nil {
		query += " AND expires_at >= ?"
		args = append(args, criteria.ExpiresAfter)
	}

	if criteria.ExpiresBefore != nil {
		query += " AND expires_at <= ?"
		args = append(args, criteria.ExpiresBefore)
	}

	// Add sorting
	if criteria.SortBy != "" {
		orderClause := "ORDER BY "
		switch criteria.SortBy {
		case "key":
			orderClause += "key"
		case "created_at":
			orderClause += "created_at"
		case "updated_at":
			orderClause += "updated_at"
		case "size":
			orderClause += "size"
		default:
			orderClause += "created_at"
		}

		if criteria.SortOrder == "desc" {
			orderClause += " DESC"
		} else {
			orderClause += " ASC"
		}

		query += " " + orderClause
	} else {
		query += " ORDER BY created_at DESC"
	}

	// Add limit and offset
	if criteria.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, criteria.Limit)
	}

	if criteria.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, criteria.Offset)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	var items []ports.StorageItem
	for rows.Next() {
		item, err := s.scanStorageItem(rows)
		if err != nil {
			continue
		}
		items = append(items, *item)
	}

	return items, rows.Err()
}

// Cleanup removes expired items
func (s *sqliteStorage) Cleanup() error {
	query := `DELETE FROM storage_items WHERE expires_at IS NOT NULL AND expires_at <= ?`
	result, err := s.db.Exec(query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired items: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired items\n", rowsAffected)
	}

	return nil
}

// Compact optimizes the database
func (s *sqliteStorage) Compact() error {
	// First cleanup expired items
	if err := s.Cleanup(); err != nil {
		return fmt.Errorf("cleanup failed during compaction: %w", err)
	}

	// Vacuum to optimize database file
	_, err := s.db.Exec("VACUUM")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	// Analyze to update statistics
	_, err = s.db.Exec("ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to analyze database: %w", err)
	}

	return nil
}

// HealthCheck performs a health check
func (s *sqliteStorage) HealthCheck() error {
	// Test database connection
	if err := s.db.Ping(); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Test read/write operation
	testKey := fmt.Sprintf(".health-check-%d", time.Now().Unix())
	testData := map[string]interface{}{
		"timestamp": time.Now(),
		"status":    "ok",
	}

	if err := s.Store(testKey, testData); err != nil {
		return fmt.Errorf("failed to write health check data: %w", err)
	}

	if _, err := s.Retrieve(testKey); err != nil {
		return fmt.Errorf("failed to read health check data: %w", err)
	}

	if err := s.Delete(testKey); err != nil {
		return fmt.Errorf("failed to delete health check data: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *sqliteStorage) Close() error {
	return s.db.Close()
}

// Helper methods

func (s *sqliteStorage) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	return nil
}

func (s *sqliteStorage) extractTags(metadata map[string]string) string {
	if metadata == nil {
		return ""
	}

	var tags []string
	for k, v := range metadata {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}

	return strings.Join(tags, ",")
}

func (s *sqliteStorage) scanStorageItem(rows *sql.Rows) (*ports.StorageItem, error) {
	var (
		key, dataStr, metadataStr string
		createdAt, updatedAt      time.Time
		expiresAt                 sql.NullTime
		size                      int64
	)

	err := rows.Scan(&key, &dataStr, &metadataStr, &createdAt, &updatedAt, &expiresAt, &size)
	if err != nil {
		return nil, err
	}

	// Parse data
	var data interface{}
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to parse data: %w", err)
	}

	// Parse metadata
	metadata := make(map[string]string)
	if metadataStr != "" {
		if err := json.Unmarshal([]byte(metadataStr), &metadata); err != nil {
			return nil, fmt.Errorf("failed to parse metadata: %w", err)
		}
	}

	item := &ports.StorageItem{
		Key:       key,
		Data:      data,
		Metadata:  metadata,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Size:      size,
	}

	if expiresAt.Valid {
		item.ExpiresAt = &expiresAt.Time
	}

	return item, nil
}

func (s *sqliteStorage) startCleanupRoutine() {
	ticker := time.NewTicker(s.cfg.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		_ = s.Cleanup()
	}
}