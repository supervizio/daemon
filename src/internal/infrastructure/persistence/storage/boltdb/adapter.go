//go:build linux

// Package boltdb provides a BoltDB adapter for metrics persistence.
package boltdb

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/storage"
)

// Bucket names for organizing data.
var (
	bucketSystemCPU      = []byte("system_cpu")
	bucketSystemMemory   = []byte("system_memory")
	bucketProcessMetrics = []byte("process_metrics")
	bucketMetadata       = []byte("metadata")
)

// Metadata keys.
var (
	keyLastPrune = []byte("last_prune")
	keyVersion   = []byte("version")
	keyCreated   = []byte("created")
)

// Current schema version.
const schemaVersion = 1

// Adapter implements MetricsStore using BoltDB.
type Adapter struct {
	db     *bolt.DB
	config storage.StoreConfig
}

// New creates a new BoltDB adapter.
func New(config storage.StoreConfig) (*Adapter, error) {
	db, err := bolt.Open(config.Path, 0o600, &bolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("open boltdb: %w", err)
	}

	adapter := &Adapter{
		db:     db,
		config: config,
	}

	if err := adapter.initSchema(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("init schema: %w", err)
	}

	return adapter, nil
}

// initSchema creates the bucket structure.
func (a *Adapter) initSchema() error {
	return a.db.Update(func(tx *bolt.Tx) error {
		// Create top-level buckets
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

		// Set metadata
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
func (a *Adapter) WriteSystemCPU(ctx context.Context, m *metrics.SystemCPU) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		key := timeToKey(m.Timestamp)
		value, err := encode(m)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

// WriteSystemMemory persists system memory metrics.
func (a *Adapter) WriteSystemMemory(ctx context.Context, m *metrics.SystemMemory) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		key := timeToKey(m.Timestamp)
		value, err := encode(m)
		if err != nil {
			return err
		}
		return b.Put(key, value)
	})
}

// WriteProcessMetrics persists process metrics.
func (a *Adapter) WriteProcessMetrics(ctx context.Context, m *metrics.ProcessMetrics) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return a.db.Update(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)

		// Create nested bucket for service
		serviceBucket, err := parent.CreateBucketIfNotExists([]byte(m.ServiceName))
		if err != nil {
			return fmt.Errorf("create service bucket: %w", err)
		}

		key := timeToKey(m.Timestamp)
		value, err := encode(m)
		if err != nil {
			return err
		}
		return serviceBucket.Put(key, value)
	})
}

// GetSystemCPU retrieves system CPU metrics within the time range.
//
//nolint:dupl // Intentional type-specific implementation for SystemCPU
func (a *Adapter) GetSystemCPU(ctx context.Context, since, until time.Time) ([]metrics.SystemCPU, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.SystemCPU
	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemCPU)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, v := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, v = c.Next() {
			var m metrics.SystemCPU
			if err := decode(v, &m); err != nil {
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
//nolint:dupl // Intentional type-specific implementation for SystemMemory
func (a *Adapter) GetSystemMemory(ctx context.Context, since, until time.Time) ([]metrics.SystemMemory, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.SystemMemory
	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketSystemMemory)
		c := b.Cursor()

		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, v := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, v = c.Next() {
			var m metrics.SystemMemory
			if err := decode(v, &m); err != nil {
				return err
			}
			result = append(result, m)
		}
		return nil
	})
	return result, err
}

// GetProcessMetrics retrieves process metrics for a service within the time range.
func (a *Adapter) GetProcessMetrics(ctx context.Context, serviceName string, since, until time.Time) ([]metrics.ProcessMetrics, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var result []metrics.ProcessMetrics
	err := a.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		if b == nil {
			return nil // No data for this service
		}

		c := b.Cursor()
		sinceKey := timeToKey(since)
		untilKey := timeToKey(until)

		for k, v := c.Seek(sinceKey); k != nil && bytes.Compare(k, untilKey) <= 0; k, v = c.Next() {
			var m metrics.ProcessMetrics
			if err := decode(v, &m); err != nil {
				return err
			}
			result = append(result, m)
		}
		return nil
	})
	return result, err
}

// getLatest is a generic helper to retrieve the most recent entry from a bucket.
func getLatest[T any](ctx context.Context, a *Adapter, bucket []byte, notFoundMsg string) (T, error) {
	var zero T
	if err := ctx.Err(); err != nil {
		return zero, err
	}

	var result T
	var found bool

	err := a.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return nil
		}
		found = true
		return decode(v, &result)
	})

	if err != nil {
		return zero, err
	}
	if !found {
		return zero, fmt.Errorf("%s", notFoundMsg)
	}
	return result, nil
}

// GetLatestSystemCPU retrieves the most recent system CPU metrics.
func (a *Adapter) GetLatestSystemCPU(ctx context.Context) (metrics.SystemCPU, error) {
	return getLatest[metrics.SystemCPU](ctx, a, bucketSystemCPU, "no system CPU metrics found")
}

// GetLatestSystemMemory retrieves the most recent system memory metrics.
func (a *Adapter) GetLatestSystemMemory(ctx context.Context) (metrics.SystemMemory, error) {
	return getLatest[metrics.SystemMemory](ctx, a, bucketSystemMemory, "no system memory metrics found")
}

// GetLatestProcessMetrics retrieves the most recent process metrics for a service.
func (a *Adapter) GetLatestProcessMetrics(ctx context.Context, serviceName string) (metrics.ProcessMetrics, error) {
	if err := ctx.Err(); err != nil {
		return metrics.ProcessMetrics{}, err
	}

	var result metrics.ProcessMetrics
	var found bool

	err := a.db.View(func(tx *bolt.Tx) error {
		parent := tx.Bucket(bucketProcessMetrics)
		b := parent.Bucket([]byte(serviceName))
		if b == nil {
			return nil
		}

		c := b.Cursor()
		k, v := c.Last()
		if k == nil {
			return nil
		}
		found = true
		return decode(v, &result)
	})

	if err != nil {
		return metrics.ProcessMetrics{}, err
	}
	if !found {
		return metrics.ProcessMetrics{}, fmt.Errorf("no process metrics found for %s", serviceName)
	}
	return result, nil
}

// Prune removes metrics older than the specified duration.
func (a *Adapter) Prune(ctx context.Context, olderThan time.Duration) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	cutoffKey := timeToKey(cutoff)
	var deleted int

	err := a.db.Update(func(tx *bolt.Tx) error {
		// Prune system CPU
		n, err := pruneBucket(tx.Bucket(bucketSystemCPU), cutoffKey)
		if err != nil {
			return err
		}
		deleted += n

		// Prune system memory
		n, err = pruneBucket(tx.Bucket(bucketSystemMemory), cutoffKey)
		if err != nil {
			return err
		}
		deleted += n

		// Prune process metrics (nested buckets)
		parent := tx.Bucket(bucketProcessMetrics)
		err = parent.ForEach(func(k, v []byte) error {
			if v != nil {
				return nil // Skip non-bucket entries
			}
			serviceBucket := parent.Bucket(k)
			if serviceBucket != nil {
				n, err := pruneBucket(serviceBucket, cutoffKey)
				if err != nil {
					return err
				}
				deleted += n
			}
			return nil
		})
		if err != nil {
			return err
		}

		// Update last prune timestamp
		meta := tx.Bucket(bucketMetadata)
		return meta.Put(keyLastPrune, int64ToBytes(time.Now().UnixNano()))
	})

	return deleted, err
}

// pruneBucket removes entries older than cutoffKey from a bucket.
func pruneBucket(b *bolt.Bucket, cutoffKey []byte) (int, error) {
	var toDelete [][]byte
	c := b.Cursor()

	for k, _ := c.First(); k != nil && bytes.Compare(k, cutoffKey) < 0; k, _ = c.Next() {
		toDelete = append(toDelete, append([]byte{}, k...))
	}

	for _, k := range toDelete {
		if err := b.Delete(k); err != nil {
			return 0, err
		}
	}

	return len(toDelete), nil
}

// Close closes the database.
func (a *Adapter) Close() error {
	return a.db.Close()
}

// timeToKey converts a time to a sortable byte key.
func timeToKey(t time.Time) []byte {
	return int64ToBytes(t.UnixNano())
}

// int64ToBytes converts an int64 to big-endian bytes.
// Used for timestamps which are always positive since Unix epoch.
func int64ToBytes(n int64) []byte {
	buf := make([]byte, 8)
	//nolint:gosec // G115: Safe conversion - timestamps are positive since Unix epoch (1970)
	binary.BigEndian.PutUint64(buf, uint64(n))
	return buf
}

// encode serializes a value using gob.
func encode(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(v); err != nil {
		return nil, fmt.Errorf("gob encode: %w", err)
	}
	return buf.Bytes(), nil
}

// decode deserializes a value using gob.
func decode(data []byte, v any) error {
	return gob.NewDecoder(bytes.NewReader(data)).Decode(v)
}
