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
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/storage"
)

const (
	// dbFileMode is the file permission mode for the BoltDB database file.
	dbFileMode os.FileMode = 0o600
	// dbOpenTimeout is the timeout in seconds for opening the database.
	dbOpenTimeout int64 = 5
	// int64ByteLength is the byte length of an int64 for binary encoding.
	int64ByteLength int = 8
)

// schemaVersion is the current schema version for database migrations.
const schemaVersion int64 = 1

var (
	// bufferPool provides reusable bytes.Buffer instances to reduce allocations
	// during gob encoding operations.
	bufferPool *sync.Pool = &sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	// bucketSystemCPU is the bucket name for system CPU metrics.
	bucketSystemCPU []byte = []byte("system_cpu")
	// bucketSystemMemory is the bucket name for system memory metrics.
	bucketSystemMemory []byte = []byte("system_memory")
	// bucketProcessMetrics is the bucket name for process metrics.
	bucketProcessMetrics []byte = []byte("process_metrics")
	// bucketMetadata is the bucket name for database metadata.
	bucketMetadata []byte = []byte("metadata")

	// keyLastPrune is the metadata key for last prune timestamp.
	keyLastPrune []byte = []byte("last_prune")
	// keyVersion is the metadata key for database version.
	keyVersion []byte = []byte("version")
	// keyCreated is the metadata key for creation timestamp.
	keyCreated []byte = []byte("created")

	// errNotFound indicates that no metrics were found in the database
	errNotFound error = errors.New("metrics not found")
)

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
	db, err := bolt.Open(config.Path, dbFileMode, &bolt.Options{
		Timeout: time.Duration(dbOpenTimeout) * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	store := &Store{
		db:     db,
		config: config,
	}

	if err := store.initSchema(); err != nil {
		// Close database on schema failure to avoid resource leak.
		_ = db.Close()

		return nil, fmt.Errorf("init schema: %w", err)
	}

	return store, nil
}

// initSchema creates the bucket structure.
//
// Returns:
//   - error: bucket creation or metadata initialization errors
func (s *Store) initSchema() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			bucketSystemCPU,
			bucketSystemMemory,
			bucketProcessMetrics,
			bucketMetadata,
		}

		for _, name := range buckets {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}

		// Initialize metadata for new databases.
		meta := tx.Bucket(bucketMetadata)
		if meta.Get(keyCreated) == nil {
			now := time.Now().UnixNano()
			if err := meta.Put(keyCreated, int64ToBytes(now)); err != nil {
				return err
			}
			if err := meta.Put(keyVersion, int64ToBytes(schemaVersion)); err != nil {
				return err
			}
		}

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
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		key := timeToKey(m.Timestamp)
		value, err := encodeSystemCPU(m)
		if err != nil {
			return err
		}

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
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		key := timeToKey(m.Timestamp)
		value, err := encodeSystemMemory(m)
		if err != nil {
			return err
		}

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
	if err := ctx.Err(); err != nil {
		return err
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)

		// Each service has its own nested bucket for metrics isolation.
		serviceBucket, err := parent.CreateBucketIfNotExists([]byte(m.ServiceName))
		if err != nil {
			return fmt.Errorf("create service bucket: %w", err)
		}

		key := timeToKey(m.Timestamp)
		value, err := encodeProcessMetrics(m)
		if err != nil {
			return err
		}

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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.SystemCPU

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.SystemCPU
			if err := decodeSystemCPU(val, &m); err != nil {
				return err
			}
			result = append(result, m)
		}

		return nil
	})

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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.SystemMemory

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.SystemMemory
			if err := decodeSystemMemory(val, &m); err != nil {
				return err
			}
			result = append(result, m)
		}

		return nil
	})

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
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.ProcessMetrics

	err := s.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.ProcessMetrics
			if err := decodeProcessMetrics(val, &m); err != nil {
				return err
			}
			result = append(result, m)
		}

		return nil
	})

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
func (s *Store) GetLatestSystemCPU(ctx context.Context) (metrics.SystemCPU, error) {
	if err := ctx.Err(); err != nil {
		return metrics.SystemCPU{}, err
	}

	var result metrics.SystemCPU
	var found bool

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		c := b.Cursor()
		// Last entry is most recent due to timestamp-based keys.
		k, val := c.Last()
		if k == nil {
			return nil
		}
		found = true

		return decodeSystemCPU(val, &result)
	})

	if err != nil {
		return metrics.SystemCPU{}, err
	}
	if !found {
		return metrics.SystemCPU{}, fmt.Errorf("no system CPU metrics found: %w", errNotFound)
	}

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
func (s *Store) GetLatestSystemMemory(ctx context.Context) (metrics.SystemMemory, error) {
	if err := ctx.Err(); err != nil {
		return metrics.SystemMemory{}, err
	}

	var result metrics.SystemMemory
	var found bool

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		c := b.Cursor()
		// Last entry is most recent due to timestamp-based keys.
		k, val := c.Last()
		if k == nil {
			return nil
		}
		found = true

		return decodeSystemMemory(val, &result)
	})

	if err != nil {
		return metrics.SystemMemory{}, err
	}
	if !found {
		return metrics.SystemMemory{}, fmt.Errorf("no system memory metrics found: %w", errNotFound)
	}

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
	if err := ctx.Err(); err != nil {
		return metrics.ProcessMetrics{}, err
	}

	var result metrics.ProcessMetrics
	var found bool

	err := s.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		// Last entry is most recent due to timestamp-based keys.
		k, val := c.Last()
		if k == nil {
			return nil
		}
		found = true

		return decodeProcessMetrics(val, &result)
	})

	if err != nil {
		return metrics.ProcessMetrics{}, err
	}
	if !found {
		return metrics.ProcessMetrics{}, fmt.Errorf("no process metrics found for %s: %w", serviceName, errNotFound)
	}

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
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	cutoffKey := timeToKey(cutoff)

	return s.pruneTransaction(cutoffKey)
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
	var deleted int

	err := s.db.Update(func(tx *bolt.Tx) error {
		n, err := s.pruneBucketHelper(tx.Bucket(bucketSystemCPU), cutoffKey)
		if err != nil {
			return err
		}
		deleted += n

		n, err = s.pruneBucketHelper(tx.Bucket(bucketSystemMemory), cutoffKey)
		if err != nil {
			return err
		}
		deleted += n

		n, err = s.pruneProcessMetricsBuckets(tx.Bucket(bucketProcessMetrics), cutoffKey)
		if err != nil {
			return err
		}
		deleted += n

		meta := tx.Bucket(bucketMetadata)

		return meta.Put(keyLastPrune, int64ToBytes(time.Now().UnixNano()))
	})

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
	var deleted int

	err := parent.ForEach(func(k, val []byte) error {
		// nil value indicates a nested bucket, not a key-value pair.
		if val != nil {
			return nil
		}

		serviceBucket := parent.Bucket(k)
		if serviceBucket != nil {
			n, err := s.pruneBucketHelper(serviceBucket, cutoffKey)
			if err != nil {
				return err
			}
			deleted += n
		}

		return nil
	})

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
	// Collect keys before deleting - cursor reuses memory, so Clone is required.
	var toDelete [][]byte
	c := b.Cursor()

	for k, _ := c.First(); k != nil && bytes.Compare(k, cutoffKey) < 0; k, _ = c.Next() {
		toDelete = append(toDelete, slices.Clone(k))
	}

	for _, k := range toDelete {
		if err := b.Delete(k); err != nil {
			return 0, err
		}
	}

	return len(toDelete), nil
}

// Close closes the database.
//
// Returns:
//   - error: database close errors
func (s *Store) Close() error {
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
	var buf [int64ByteLength]byte
	//nolint:gosec // G115: Safe conversion - timestamps are positive since Unix epoch (1970)
	binary.BigEndian.PutUint64(buf[:], uint64(n))

	return buf[:]
}

// encodeSystemCPU serializes system CPU metrics using gob.
//
// Params:
//   - data: system CPU metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors (unreachable with current types)
//
// Note: gob.Encode cannot fail for this struct (primitive types only: uint64, time.Time).
// The error handling is retained for robustness if struct types change in the future.
// Coverage gap on error branch is accepted as theoretically unreachable with current types.
func encodeSystemCPU(data *metrics.SystemCPU) ([]byte, error) {
	buf, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
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
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// encodeSystemMemory serializes system memory metrics using gob.
//
// Params:
//   - data: system memory metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors (unreachable with current types)
//
// Note: gob.Encode cannot fail for this struct (primitive types only: uint64, time.Time).
// The error handling is retained for robustness if struct types change in the future.
// Coverage gap on error branch is accepted as theoretically unreachable with current types.
func encodeSystemMemory(data *metrics.SystemMemory) ([]byte, error) {
	buf, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
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
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// encodeProcessMetrics serializes process metrics using gob.
//
// Params:
//   - data: process metrics to encode
//
// Returns:
//   - []byte: encoded bytes
//   - error: encoding errors (unreachable with current types)
//
// Note: gob.Encode cannot fail for this struct (primitive types only: uint64, int, time.Time).
// The error handling is retained for robustness if struct types change in the future.
// Coverage gap on error branch is accepted as theoretically unreachable with current types.
func encodeProcessMetrics(data *metrics.ProcessMetrics) ([]byte, error) {
	buf, ok := bufferPool.Get().(*bytes.Buffer)
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
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
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// Db returns the underlying BoltDB instance for testing purposes.
//
// Returns:
//   - *bolt.DB: the underlying database instance.
func (s *Store) Db() *bolt.DB {
	return s.db
}
