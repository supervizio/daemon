//go:build linux

// Package boltdb provides a BoltDB store for metrics persistence.
package boltdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/storage"
)

// File permissions and timeouts for database operations.
const (
	dbFileMode      os.FileMode = 0o600
	dbOpenTimeout   int64       = 5
	int64ByteLength int         = 8
)

// Current schema version for database migrations.
const schemaVersion int64 = 1

// bucketReader defines the interface for reading from a bucket.
// Used to decouple pruning logic from concrete bbolt.Bucket type.
type bucketReader interface {
	Bucket(name []byte) *bolt.Bucket
	ForEach(fn func(k []byte, v []byte) error) error
}

// bucketPruner defines the interface for pruning entries from a bucket.
// Used to decouple pruning logic from concrete bbolt.Bucket type.
type bucketPruner interface {
	Cursor() *bolt.Cursor
	Delete(key []byte) error
}

// Database organization: bucket names and metadata keys.
var (
	// Bucket names for organizing data
	bucketSystemCPU      []byte = []byte("system_cpu")
	bucketSystemMemory   []byte = []byte("system_memory")
	bucketProcessMetrics []byte = []byte("process_metrics")
	bucketMetadata       []byte = []byte("metadata")

	// Metadata keys for storing database metadata
	keyLastPrune []byte = []byte("last_prune")
	keyVersion   []byte = []byte("version")
	keyCreated   []byte = []byte("created")

	// errNotFound indicates that no metrics were found in the database
	errNotFound error = errors.New("metrics not found")
)

// Store implements MetricsStore using BoltDB.
//
// Store provides persistent storage for system and process metrics using
// an embedded BoltDB database with time-series optimizations.
type Store struct {
	db     *bolt.DB
	config storage.StoreConfig
}

// NewStore creates a new BoltDB store.
//
// Params:
//   - config: storage configuration containing database path
//
// Returns:
//   - *Store: initialized store instance
//   - error: database open or schema initialization errors
func NewStore(config storage.StoreConfig) (*Store, error) {
	// Open database with timeout and file permissions
	db, err := bolt.Open(config.Path, dbFileMode, &bolt.Options{
		Timeout: time.Duration(dbOpenTimeout) * time.Second,
	})
	// Check if database open failed
	if err != nil {
		// Database file could not be opened
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	// Create store instance
	store := &Store{
		db:     db,
		config: config,
	}

	// Initialize database schema
	if err := store.initSchema(); err != nil {
		// Schema initialization failed, close database before returning
		_ = db.Close()
		// Return wrapped error
		return nil, fmt.Errorf("init schema: %w", err)
	}

	// Store ready for use
	return store, nil
}

// initSchema creates the bucket structure.
//
// Returns:
//   - error: bucket creation or metadata initialization errors
func (s *Store) initSchema() error {
	// Execute schema setup in a write transaction
	return s.db.Update(func(tx *bolt.Tx) error {
		// Define all required top-level buckets
		buckets := [][]byte{
			bucketSystemCPU,
			bucketSystemMemory,
			bucketProcessMetrics,
			bucketMetadata,
		}

		// Create each bucket if it doesn't exist
		for _, name := range buckets {
			// Attempt to create bucket
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				// Bucket creation failed
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}

		// Get metadata bucket for version tracking
		meta := tx.Bucket(bucketMetadata)
		// Check if this is a new database
		if meta.Get(keyCreated) == nil {
			// Record database creation timestamp
			now := time.Now().UnixNano()
			// Write creation timestamp
			if err := meta.Put(keyCreated, int64ToBytes(now)); err != nil {
				// Failed to write creation timestamp
				return err
			}
			// Write schema version for migrations
			if err := meta.Put(keyVersion, int64ToBytes(schemaVersion)); err != nil {
				// Failed to write schema version
				return err
			}
		}

		// Schema initialization complete
		return nil
	})
}

// WriteSystemCPU persists system CPU metrics.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - m: system CPU metrics to persist
//
// Returns:
//   - error: context cancellation or database write errors
func (s *Store) WriteSystemCPU(ctx context.Context, m *metrics.SystemCPU) error {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort write
		return err
	}

	// Write metrics in a database transaction
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the system CPU bucket
		b := tx.Bucket(bucketSystemCPU)
		// Convert timestamp to sortable key
		key := timeToKey(m.Timestamp)
		// Serialize metrics to bytes
		value, err := encodeSystemCPU(m)
		// Check if encoding failed
		if err != nil {
			// Encoding failed
			return err
		}
		// Store metrics with timestamp key
		return b.Put(key, value)
	})
}

// WriteSystemMemory persists system memory metrics.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - m: system memory metrics to persist
//
// Returns:
//   - error: context cancellation or database write errors
func (s *Store) WriteSystemMemory(ctx context.Context, m *metrics.SystemMemory) error {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort write
		return err
	}

	// Write metrics in a database transaction
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the system memory bucket
		b := tx.Bucket(bucketSystemMemory)
		// Convert timestamp to sortable key
		key := timeToKey(m.Timestamp)
		// Serialize metrics to bytes
		value, err := encodeSystemMemory(m)
		// Check if encoding failed
		if err != nil {
			// Encoding failed
			return err
		}
		// Store metrics with timestamp key
		return b.Put(key, value)
	})
}

// WriteProcessMetrics persists process metrics.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - m: process metrics to persist
//
// Returns:
//   - error: context cancellation or database write errors
func (s *Store) WriteProcessMetrics(ctx context.Context, m *metrics.ProcessMetrics) error {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort write
		return err
	}

	// Write metrics in a database transaction
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get parent bucket for all process metrics
		parent := tx.Bucket(bucketProcessMetrics)

		// Create nested bucket for this service's metrics
		serviceBucket, err := parent.CreateBucketIfNotExists([]byte(m.ServiceName))
		// Check if bucket creation failed
		if err != nil {
			// Service bucket creation failed
			return fmt.Errorf("create service bucket: %w", err)
		}

		// Convert timestamp to sortable key
		key := timeToKey(m.Timestamp)
		// Serialize metrics to bytes
		value, err := encodeProcessMetrics(m)
		// Check if encoding failed
		if err != nil {
			// Encoding failed
			return err
		}
		// Store metrics in service-specific bucket
		return serviceBucket.Put(key, value)
	})
}

// GetSystemCPU retrieves system CPU metrics within the time range.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - since: start of time range (inclusive)
//   - until: end of time range (inclusive)
//
// Returns:
//   - []metrics.SystemCPU: metrics within the time range
//   - error: context cancellation or database read errors
//
//nolint:dupl // Intentional type-specific implementation for SystemCPU
func (s *Store) GetSystemCPU(ctx context.Context, since, until time.Time) ([]metrics.SystemCPU, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return nil, err
	}

	// Collect matching metrics
	var result []metrics.SystemCPU
	// Read metrics in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the system CPU bucket
		b := tx.Bucket(bucketSystemCPU)
		// Create cursor for range iteration
		c := b.Cursor()

		// Convert time range to sortable keys
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// Iterate through metrics in time range
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			// Deserialize metric from stored bytes
			var m metrics.SystemCPU
			// Attempt to decode the value
			if err := decodeSystemCPU(val, &m); err != nil {
				// Decoding failed
				return err
			}
			// Add metric to result set
			result = append(result, m)
		}
		// All metrics retrieved successfully
		return nil
	})
	// Return collected metrics and any error
	return result, err
}

// GetSystemMemory retrieves system memory metrics within the time range.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - since: start of time range (inclusive)
//   - until: end of time range (inclusive)
//
// Returns:
//   - []metrics.SystemMemory: metrics within the time range
//   - error: context cancellation or database read errors
//
//nolint:dupl // Intentional type-specific implementation for SystemMemory
func (s *Store) GetSystemMemory(ctx context.Context, since, until time.Time) ([]metrics.SystemMemory, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return nil, err
	}

	// Collect matching metrics
	var result []metrics.SystemMemory
	// Read metrics in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the system memory bucket
		b := tx.Bucket(bucketSystemMemory)
		// Create cursor for range iteration
		c := b.Cursor()

		// Convert time range to sortable keys
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// Iterate through metrics in time range
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			// Deserialize metric from stored bytes
			var m metrics.SystemMemory
			// Attempt to decode the value
			if err := decodeSystemMemory(val, &m); err != nil {
				// Decoding failed
				return err
			}
			// Add metric to result set
			result = append(result, m)
		}
		// All metrics retrieved successfully
		return nil
	})
	// Return collected metrics and any error
	return result, err
}

// GetProcessMetrics retrieves process metrics for a service within the time range.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - serviceName: name of the service to query
//   - since: start of time range (inclusive)
//   - until: end of time range (inclusive)
//
// Returns:
//   - []metrics.ProcessMetrics: metrics within the time range
//   - error: context cancellation or database read errors
func (s *Store) GetProcessMetrics(ctx context.Context, serviceName string, since, until time.Time) ([]metrics.ProcessMetrics, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return nil, err
	}

	// Collect matching metrics
	var result []metrics.ProcessMetrics
	// Read metrics in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get parent bucket for all process metrics
		parent := tx.Bucket(bucketProcessMetrics)
		// Get service-specific bucket
		b := parent.Bucket([]byte(serviceName))
		// Check if service bucket exists
		if b == nil {
			// No data exists for this service, return empty result
			return nil
		}

		// Create cursor for range iteration
		c := b.Cursor()
		// Convert time range to sortable keys
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// Iterate through metrics in time range
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			// Deserialize metric from stored bytes
			var m metrics.ProcessMetrics
			// Attempt to decode the value
			if err := decodeProcessMetrics(val, &m); err != nil {
				// Decoding failed
				return err
			}
			// Add metric to result set
			result = append(result, m)
		}
		// All metrics retrieved successfully
		return nil
	})
	// Return collected metrics and any error
	return result, err
}

// GetLatestSystemCPU retrieves the most recent system CPU metrics.
//
// Params:
//   - ctx: context for cancellation and timeout
//
// Returns:
//   - metrics.SystemCPU: most recent CPU metrics
//   - error: context cancellation, database errors, or not found
//
//nolint:dupl // Intentional type-specific implementation for SystemCPU
func (s *Store) GetLatestSystemCPU(ctx context.Context) (metrics.SystemCPU, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return metrics.SystemCPU{}, err
	}

	// Storage for retrieved metric
	var result metrics.SystemCPU
	// Track whether any data was found
	var found bool

	// Read latest metric in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the target bucket
		b := tx.Bucket(bucketSystemCPU)
		// Create cursor for navigation
		c := b.Cursor()
		// Get last entry (most recent due to timestamp keys)
		k, val := c.Last()
		// Check if bucket has any data
		if k == nil {
			// Bucket is empty, no data available
			return nil
		}
		// Mark that we found data
		found = true
		// Deserialize the metric
		return decodeSystemCPU(val, &result)
	})

	// Check for database errors
	if err != nil {
		// Database operation failed
		return metrics.SystemCPU{}, err
	}
	// Check if any data was found
	if !found {
		// No data in bucket, return error with context
		return metrics.SystemCPU{}, fmt.Errorf("no system CPU metrics found: %w", errNotFound)
	}
	// Return the retrieved metric
	return result, nil
}

// GetLatestSystemMemory retrieves the most recent system memory metrics.
//
// Params:
//   - ctx: context for cancellation and timeout
//
// Returns:
//   - metrics.SystemMemory: most recent memory metrics
//   - error: context cancellation, database errors, or not found
//
//nolint:dupl // Intentional type-specific implementation for SystemMemory
func (s *Store) GetLatestSystemMemory(ctx context.Context) (metrics.SystemMemory, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return metrics.SystemMemory{}, err
	}

	// Storage for retrieved metric
	var result metrics.SystemMemory
	// Track whether any data was found
	var found bool

	// Read latest metric in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the target bucket
		b := tx.Bucket(bucketSystemMemory)
		// Create cursor for navigation
		c := b.Cursor()
		// Get last entry (most recent due to timestamp keys)
		k, val := c.Last()
		// Check if bucket has any data
		if k == nil {
			// Bucket is empty, no data available
			return nil
		}
		// Mark that we found data
		found = true
		// Deserialize the metric
		return decodeSystemMemory(val, &result)
	})

	// Check for database errors
	if err != nil {
		// Database operation failed
		return metrics.SystemMemory{}, err
	}
	// Check if any data was found
	if !found {
		// No data in bucket, return error with context
		return metrics.SystemMemory{}, fmt.Errorf("no system memory metrics found: %w", errNotFound)
	}
	// Return the retrieved metric
	return result, nil
}

// GetLatestProcessMetrics retrieves the most recent process metrics for a service.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - serviceName: name of the service to query
//
// Returns:
//   - metrics.ProcessMetrics: most recent process metrics
//   - error: context cancellation, database errors, or not found
func (s *Store) GetLatestProcessMetrics(ctx context.Context, serviceName string) (metrics.ProcessMetrics, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort read
		return metrics.ProcessMetrics{}, err
	}

	// Storage for retrieved metric
	var result metrics.ProcessMetrics
	// Track whether any data was found
	var found bool

	// Read latest metric in a read-only transaction
	err := s.db.View(func(tx *bolt.Tx) error {
		// Get parent bucket for all process metrics
		parent := tx.Bucket(bucketProcessMetrics)
		// Get service-specific bucket
		b := parent.Bucket([]byte(serviceName))
		// Check if service bucket exists
		if b == nil {
			// No bucket exists for this service
			return nil
		}

		// Create cursor for navigation
		c := b.Cursor()
		// Get last entry (most recent due to timestamp keys)
		k, val := c.Last()
		// Check if bucket has any data
		if k == nil {
			// Bucket is empty, no data available
			return nil
		}
		// Mark that we found data
		found = true
		// Deserialize the metric
		return decodeProcessMetrics(val, &result)
	})

	// Check for database errors
	if err != nil {
		// Database operation failed
		return metrics.ProcessMetrics{}, err
	}
	// Check if any data was found
	if !found {
		// No data for this service, return error with context
		return metrics.ProcessMetrics{}, fmt.Errorf("no process metrics found for %s: %w", serviceName, errNotFound)
	}
	// Return the retrieved metric
	return result, nil
}

// Prune removes metrics older than the specified duration.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - olderThan: age threshold for deletion
//
// Returns:
//   - int: number of entries deleted
//   - error: context cancellation or database errors
func (s *Store) Prune(ctx context.Context, olderThan time.Duration) (int, error) {
	// Check if context is already cancelled
	if err := ctx.Err(); err != nil {
		// Context cancelled, abort prune
		return 0, err
	}

	// Calculate cutoff timestamp for deletion
	cutoff := time.Now().Add(-olderThan)
	cutoffKey := timeToKey(cutoff)

	// Execute pruning in a write transaction
	deleted, err := s.pruneTransaction(cutoffKey)
	// Return deletion count and any error
	return deleted, err
}

// pruneTransaction executes the pruning logic within a transaction.
//
// Params:
//   - cutoffKey: timestamp key threshold for deletion
//
// Returns:
//   - int: number of entries deleted
//   - error: database errors
func (s *Store) pruneTransaction(cutoffKey []byte) (int, error) {
	// Track total number of deleted entries
	var deleted int

	// Execute pruning in a write transaction
	err := s.db.Update(func(tx *bolt.Tx) error {
		// Prune system CPU metrics
		n, err := s.pruneBucketHelper(tx.Bucket(bucketSystemCPU), cutoffKey)
		// Check if pruning failed
		if err != nil {
			// System CPU pruning failed
			return err
		}
		// Add to deletion count
		deleted += n

		// Prune system memory metrics
		n, err = s.pruneBucketHelper(tx.Bucket(bucketSystemMemory), cutoffKey)
		// Check if pruning failed
		if err != nil {
			// System memory pruning failed
			return err
		}
		// Add to deletion count
		deleted += n

		// Prune process metrics across all services
		n, err = s.pruneProcessMetricsBuckets(tx.Bucket(bucketProcessMetrics), cutoffKey)
		// Check if pruning failed
		if err != nil {
			// Process metrics pruning failed
			return err
		}
		// Add to deletion count
		deleted += n

		// Update last prune timestamp in metadata
		meta := tx.Bucket(bucketMetadata)
		// Record when pruning completed
		return meta.Put(keyLastPrune, int64ToBytes(time.Now().UnixNano()))
	})

	// Return deletion count and any error
	return deleted, err
}

// pruneProcessMetricsBuckets prunes all service-specific process metrics buckets.
//
// Params:
//   - parent: parent bucket containing service buckets
//   - cutoffKey: timestamp key threshold for deletion
//
// Returns:
//   - int: number of entries deleted
//   - error: database errors
func (s *Store) pruneProcessMetricsBuckets(parent bucketReader, cutoffKey []byte) (int, error) {
	// Track deletion count
	var deleted int

	// Iterate through all service buckets
	err := parent.ForEach(func(k, val []byte) error {
		// Check if this is a bucket (not a key-value pair)
		if val != nil {
			// This is a key-value pair, not a bucket, skip it
			return nil
		}
		// Get the service-specific bucket
		serviceBucket := parent.Bucket(k)
		// Verify bucket exists
		if serviceBucket != nil {
			// Prune this service's metrics
			n, err := s.pruneBucketHelper(serviceBucket, cutoffKey)
			// Check if pruning failed
			if err != nil {
				// Service bucket pruning failed
				return err
			}
			// Add to deletion count
			deleted += n
		}
		// Continue to next service
		return nil
	})

	// Return deletion count and any error
	return deleted, err
}

// pruneBucketHelper removes entries older than cutoffKey from a bucket.
//
// Params:
//   - b: bucket to prune
//   - cutoffKey: timestamp key threshold
//
// Returns:
//   - int: number of entries deleted
//   - error: database deletion errors
func (s *Store) pruneBucketHelper(b bucketPruner, cutoffKey []byte) (int, error) {
	// Collect keys to delete (can't delete during iteration)
	var toDelete [][]byte
	// Create cursor for iteration
	c := b.Cursor()

	// Iterate through all entries older than cutoff
	for k, _ := c.First(); k != nil && bytes.Compare(k, cutoffKey) < 0; k, _ = c.Next() {
		// Clone key (cursor reuses memory) and add to deletion list
		toDelete = append(toDelete, slices.Clone(k))
	}

	// Delete all collected keys
	for _, k := range toDelete {
		// Remove this entry from the bucket
		if err := b.Delete(k); err != nil {
			// Deletion failed
			return 0, err
		}
	}

	// Return count of deleted entries
	return len(toDelete), nil
}

// Close closes the database.
//
// Returns:
//   - error: database close errors
func (s *Store) Close() error {
	// Close the underlying BoltDB connection
	return s.db.Close()
}

// timeToKey converts a time to a sortable byte key.
//
// Params:
//   - t: timestamp to convert
//
// Returns:
//   - []byte: sortable byte representation
func timeToKey(t time.Time) []byte {
	// Convert to nanoseconds and encode as sortable bytes
	return int64ToBytes(t.UnixNano())
}

// int64ToBytes converts an int64 to big-endian bytes.
// Used for timestamps which are always positive since Unix epoch.
//
// Params:
//   - n: integer to convert
//
// Returns:
//   - []byte: big-endian byte representation
func int64ToBytes(n int64) []byte {
	// Use fixed-size array for known length
	var buf [int64ByteLength]byte
	//nolint:gosec // G115: Safe conversion - timestamps are positive since Unix epoch (1970)
	binary.BigEndian.PutUint64(buf[:], uint64(n))
	// Return slice of the array
	return buf[:]
}

// encodeSystemCPU serializes system CPU metrics using gob.
//
// Params:
//   - data: system CPU metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors
func encodeSystemCPU(data *metrics.SystemCPU) ([]byte, error) {
	// Create buffer for encoded data
	var buf bytes.Buffer
	// Encode value using gob
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		// Encoding failed
		return nil, fmt.Errorf("gob encode: %w", err)
	}
	// Return encoded bytes
	return buf.Bytes(), nil
}

// decodeSystemCPU deserializes system CPU metrics using gob.
//
// Params:
//   - data: encoded bytes to decode
//   - dest: destination for decoded metrics
//
// Returns:
//   - error: decoding errors
func decodeSystemCPU(data []byte, dest *metrics.SystemCPU) error {
	// Decode from bytes into the provided pointer
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// encodeSystemMemory serializes system memory metrics using gob.
//
// Params:
//   - data: system memory metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors
func encodeSystemMemory(data *metrics.SystemMemory) ([]byte, error) {
	// Create buffer for encoded data
	var buf bytes.Buffer
	// Encode value using gob
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		// Encoding failed
		return nil, fmt.Errorf("gob encode: %w", err)
	}
	// Return encoded bytes
	return buf.Bytes(), nil
}

// decodeSystemMemory deserializes system memory metrics using gob.
//
// Params:
//   - data: encoded bytes to decode
//   - dest: destination for decoded metrics
//
// Returns:
//   - error: decoding errors
func decodeSystemMemory(data []byte, dest *metrics.SystemMemory) error {
	// Decode from bytes into the provided pointer
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// encodeProcessMetrics serializes process metrics using gob.
//
// Params:
//   - data: process metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors
func encodeProcessMetrics(data *metrics.ProcessMetrics) ([]byte, error) {
	// Create buffer for encoded data
	var buf bytes.Buffer
	// Encode value using gob
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		// Encoding failed
		return nil, fmt.Errorf("gob encode: %w", err)
	}
	// Return encoded bytes
	return buf.Bytes(), nil
}

// decodeProcessMetrics deserializes process metrics using gob.
//
// Params:
//   - data: encoded bytes to decode
//   - dest: destination for decoded metrics
//
// Returns:
//   - error: decoding errors
func decodeProcessMetrics(data []byte, dest *metrics.ProcessMetrics) error {
	// Decode from bytes into the provided pointer
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}
