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
	// propagate database open failures immediately
	if err != nil {
		// return error with context wrapping
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	store := &Store{
		db:     db,
		config: config,
	}

	// initialize buckets and metadata on first use
	if err := store.initSchema(); err != nil {
		// Close database on schema failure to avoid resource leak.
		_ = db.Close()

		// return error with context wrapping
		return nil, fmt.Errorf("init schema: %w", err)
	}

	// return initialized store
	return store, nil
}

// initSchema creates the bucket structure.
//
// Returns:
//   - error: bucket creation or metadata initialization errors
func (s *Store) initSchema() error {
	// create all required buckets and metadata atomically
	return s.db.Update(func(tx *bolt.Tx) error {
		buckets := [][]byte{
			bucketSystemCPU,
			bucketSystemMemory,
			bucketProcessMetrics,
			bucketMetadata,
		}

		// ensure all buckets exist before writing data
		for _, name := range buckets {
			// propagate bucket creation errors to abort transaction
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				// return error with bucket name context
				return fmt.Errorf("create bucket %s: %w", name, err)
			}
		}

		// Initialize metadata for new databases.
		meta := tx.Bucket(bucketMetadata)
		// detect first-time initialization by checking for creation timestamp
		if meta.Get(keyCreated) == nil {
			now := time.Now().UnixNano()
			// store creation timestamp for database lifecycle tracking
			if err := meta.Put(keyCreated, int64ToBytes(now)); err != nil {
				// propagate metadata write error
				return err
			}
			// record schema version for future migration compatibility
			if err := meta.Put(keyVersion, int64ToBytes(schemaVersion)); err != nil {
				// propagate metadata write error
				return err
			}
		}

		// signal transaction success
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
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return err
	}

	// write metrics atomically to prevent partial updates
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		key := timeToKey(m.Timestamp)
		value, err := encodeSystemCPU(m)
		// abort transaction if encoding fails
		if err != nil {
			// propagate encoding error
			return err
		}

		// persist encoded metrics
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
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return err
	}

	// write metrics atomically to prevent partial updates
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		key := timeToKey(m.Timestamp)
		value, err := encodeSystemMemory(m)
		// abort transaction if encoding fails
		if err != nil {
			// propagate encoding error
			return err
		}

		// persist encoded metrics
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
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return err
	}

	// write metrics atomically to prevent partial updates
	return s.db.Update(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)

		// Each service has its own nested bucket for metrics isolation.
		serviceBucket, err := parent.CreateBucketIfNotExists([]byte(m.ServiceName))
		// abort transaction if bucket creation fails
		if err != nil {
			// return error with service context
			return fmt.Errorf("create service bucket: %w", err)
		}

		key := timeToKey(m.Timestamp)
		value, err := encodeProcessMetrics(m)
		// abort transaction if encoding fails
		if err != nil {
			// propagate encoding error
			return err
		}

		// persist encoded metrics
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
//nolint:dupl // Type-specific implementation; same structure as GetSystemMemory but different types.
func (s *Store) GetSystemCPU(ctx context.Context, since, until time.Time) ([]metrics.SystemCPU, error) {
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return nil, err
	}

	var result []metrics.SystemCPU

	// read metrics snapshot without blocking writers
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// seek to start of range and iterate through matching entries
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.SystemCPU
			// abort read if any entry cannot be decoded
			if err := decodeSystemCPU(val, &m); err != nil {
				// propagate decoding error
				return err
			}
			result = append(result, m)
		}

		// signal transaction success
		return nil
	})

	// return collected metrics or error
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
//nolint:dupl // Type-specific implementation; same structure as GetSystemCPU but different types.
func (s *Store) GetSystemMemory(ctx context.Context, since, until time.Time) ([]metrics.SystemMemory, error) {
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return nil, err
	}

	var result []metrics.SystemMemory

	// read metrics snapshot without blocking writers
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// seek to start of range and iterate through matching entries
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.SystemMemory
			// abort read if any entry cannot be decoded
			if err := decodeSystemMemory(val, &m); err != nil {
				// propagate decoding error
				return err
			}
			result = append(result, m)
		}

		// signal transaction success
		return nil
	})

	// return collected metrics or error
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
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return nil, err
	}

	var result []metrics.ProcessMetrics

	// read metrics snapshot without blocking writers
	err := s.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		// return empty result if service has no metrics yet
		if b == nil {
			// signal no error but empty result
			return nil
		}

		c := b.Cursor()
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		// seek to start of range and iterate through matching entries
		for k, val := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, val = c.Next() {
			var m metrics.ProcessMetrics
			// abort read if any entry cannot be decoded
			if err := decodeProcessMetrics(val, &m); err != nil {
				// propagate decoding error
				return err
			}
			result = append(result, m)
		}

		// signal transaction success
		return nil
	})

	// return collected metrics or error
	return result, err
}

// getLatestFromBucket retrieves the most recent entry from a bucket.
//
// Params:
//   - ctx: context for cancellation and timeout
//   - bucket: the bucket to query
//   - decodeFn: function to decode the value
//   - metricName: name for error messages
//
// Returns:
//   - error: context cancellation, database errors, or not found
func (s *Store) getLatestFromBucket(ctx context.Context, bucket []byte, decodeFn func([]byte) error, metricName string) error {
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return err
	}

	var found bool

	// read latest entry without blocking writers
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		c := b.Cursor()
		// Last entry is most recent due to timestamp-based keys.
		k, val := c.Last()
		// return empty if no metrics exist yet
		if k == nil {
			// signal no error but not found
			return nil
		}
		found = true

		// decode and return latest metrics
		return decodeFn(val)
	})

	// propagate decoding errors
	if err != nil {
		// return with error
		return err
	}
	// return not found error if no metrics exist
	if !found {
		// return sentinel error for missing metrics
		return fmt.Errorf("no %s metrics found: %w", metricName, errNotFound)
	}

	// return found
	return nil
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
	var result metrics.SystemCPU
	decodeFn := func(val []byte) error {
		return decodeSystemCPU(val, &result)
	}
	err := s.getLatestFromBucket(ctx, bucketSystemCPU, decodeFn, "system CPU")
	return result, err
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
	var result metrics.SystemMemory
	decodeFn := func(val []byte) error {
		return decodeSystemMemory(val, &result)
	}
	err := s.getLatestFromBucket(ctx, bucketSystemMemory, decodeFn, "system memory")
	return result, err
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
	// respect context cancellation before starting database transaction
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return metrics.ProcessMetrics{}, err
	}

	var result metrics.ProcessMetrics
	var found bool

	// read latest entry without blocking writers
	err := s.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		// return empty if service has no metrics yet
		if b == nil {
			// signal no error but not found
			return nil
		}

		c := b.Cursor()
		// Last entry is most recent due to timestamp-based keys.
		k, val := c.Last()
		// return empty if bucket is empty
		if k == nil {
			// signal no error but not found
			return nil
		}
		found = true

		// decode and return latest metrics
		return decodeProcessMetrics(val, &result)
	})

	// propagate decoding errors
	if err != nil {
		// return zero value with error
		return metrics.ProcessMetrics{}, err
	}
	// return not found error if no metrics exist
	if !found {
		// return sentinel error for missing metrics
		return metrics.ProcessMetrics{}, fmt.Errorf("no process metrics found for %s: %w", serviceName, errNotFound)
	}

	// return found metrics
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
	// respect context cancellation before starting deletion
	if err := ctx.Err(); err != nil {
		// propagate cancellation error
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	cutoffKey := timeToKey(cutoff)

	// execute deletion and return count
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

	// delete old entries atomically across all buckets
	err := s.db.Update(func(tx *bolt.Tx) error {
		n, err := s.pruneBucketHelper(tx.Bucket(bucketSystemCPU), cutoffKey)
		// abort transaction if deletion fails
		if err != nil {
			// propagate deletion error
			return err
		}
		deleted += n

		n, err = s.pruneBucketHelper(tx.Bucket(bucketSystemMemory), cutoffKey)
		// abort transaction if deletion fails
		if err != nil {
			// propagate deletion error
			return err
		}
		deleted += n

		n, err = s.pruneProcessMetricsBuckets(tx.Bucket(bucketProcessMetrics), cutoffKey)
		// abort transaction if deletion fails
		if err != nil {
			// propagate deletion error
			return err
		}
		deleted += n

		meta := tx.Bucket(bucketMetadata)

		// update last prune timestamp
		return meta.Put(keyLastPrune, int64ToBytes(time.Now().UnixNano()))
	})

	// return deletion count and error
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

	// iterate through all service buckets to delete old metrics
	err := parent.ForEach(func(k, val []byte) error {
		// nil value indicates a nested bucket, not a key-value pair.
		// skip regular key-value pairs to only process buckets
		if val != nil {
			// continue iteration for non-bucket entries
			return nil
		}

		serviceBucket := parent.Bucket(k)
		// prune this service bucket if it exists
		if serviceBucket != nil {
			n, err := s.pruneBucketHelper(serviceBucket, cutoffKey)
			// abort iteration if deletion fails
			if err != nil {
				// propagate deletion error
				return err
			}
			deleted += n
		}

		return nil
	})

	// return deletion count and error
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

	// collect all keys older than cutoff before deletion
	for k, _ := c.First(); k != nil && bytes.Compare(k, cutoffKey) < 0; k, _ = c.Next() {
		toDelete = append(toDelete, slices.Clone(k))
	}

	// delete collected keys in separate pass to avoid cursor issues
	for _, k := range toDelete {
		// abort deletion on first error to prevent partial cleanup
		if err := b.Delete(k); err != nil {
			// return error immediately to abort deletion
			return 0, err
		}
	}

	// return count of deleted entries
	return len(toDelete), nil
}

// Close closes the database.
//
// Returns:
//   - error: database close errors
func (s *Store) Close() error {
	// release file lock and flush pending writes
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
	// convert timestamp to sortable bytes
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
	binary.BigEndian.PutUint64(buf[:], uint64(n))

	// return byte slice representation
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
	// allocate new buffer if pool returns unexpected type
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	// abort encoding if serialization fails
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		// return error with context
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	// return encoded bytes
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
	// deserialize metrics from encoded bytes
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
	// allocate new buffer if pool returns unexpected type
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	// abort encoding if serialization fails
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		// return error with context
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	// return encoded bytes
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
	// deserialize metrics from encoded bytes
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
	// allocate new buffer if pool returns unexpected type
	if !ok {
		buf = new(bytes.Buffer)
	}
	buf.Reset()
	defer bufferPool.Put(buf)

	// abort encoding if serialization fails
	if err := gob.NewEncoder(buf).Encode(data); err != nil {
		// return error with context
		return nil, fmt.Errorf("gob encode: %w", err)
	}

	// Copy bytes - buffer will be reused by pool.
	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	// return encoded bytes
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
	// deserialize metrics from encoded bytes
	return gob.NewDecoder(bytes.NewReader(data)).Decode(dest)
}

// Db returns the underlying BoltDB instance for testing purposes.
//
// Returns:
//   - *bolt.DB: the underlying database instance.
func (s *Store) Db() *bolt.DB {
	// expose database for testing purposes
	return s.db
}
