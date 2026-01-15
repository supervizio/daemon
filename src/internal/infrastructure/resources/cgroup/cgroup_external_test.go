//go:build linux

// Package cgroup_test provides external tests for the cgroup package.
package cgroup_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)

func TestVersion_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		version cgroup.Version
		want    string
	}{
		{cgroup.VersionV1, "v1"},
		{cgroup.VersionV2, "v2"},
		{cgroup.VersionHybrid, "hybrid"},
		{cgroup.VersionUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.version.String())
		})
	}
}

func TestDetectWithPath(t *testing.T) {
	t.Parallel()

	t.Run("detects v2", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cgroup.controllers"), []byte("cpu memory io"), 0o644))

		version := cgroup.DetectWithPath(mockCgroup)
		assert.Equal(t, cgroup.VersionV2, version)
	})

	t.Run("detects v1", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "cpu"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "memory"), 0o755))

		version := cgroup.DetectWithPath(mockCgroup)
		assert.Equal(t, cgroup.VersionV1, version)
	})

	t.Run("detects hybrid", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cgroup.controllers"), []byte("cpu memory"), 0o644))
		require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "cpu"), 0o755))
		require.NoError(t, os.MkdirAll(filepath.Join(mockCgroup, "memory"), 0o755))

		version := cgroup.DetectWithPath(mockCgroup)
		assert.Equal(t, cgroup.VersionHybrid, version)
	})

	t.Run("returns unknown for empty directory", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		version := cgroup.DetectWithPath(mockCgroup)
		assert.Equal(t, cgroup.VersionUnknown, version)
	})
}

func TestV2Reader_CPUUsage(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()
	cpuStat := `usage_usec 1234567
user_usec 800000
system_usec 434567
nr_periods 100
nr_throttled 5
throttled_usec 50000
`
	require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cpu.stat"), []byte(cpuStat), 0o644))

	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	usage, err := reader.CPUUsage(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint64(1234567), usage)
}

func TestV2Reader_CPULimit(t *testing.T) {
	t.Parallel()

	t.Run("parses quota and period", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cpu.max"), []byte("100000 100000\n"), 0o644))

		reader, err := cgroup.NewV2Reader(mockCgroup)
		require.NoError(t, err)

		quota, period, err := reader.CPULimit(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(100000), quota)
		assert.Equal(t, uint64(100000), period)
	})

	t.Run("handles max (unlimited)", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "cpu.max"), []byte("max 100000\n"), 0o644))

		reader, err := cgroup.NewV2Reader(mockCgroup)
		require.NoError(t, err)

		quota, period, err := reader.CPULimit(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(0), quota)
		assert.Equal(t, uint64(100000), period)
	})

	t.Run("handles missing file", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		reader, err := cgroup.NewV2Reader(mockCgroup)
		require.NoError(t, err)

		quota, period, err := reader.CPULimit(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(0), quota)
		assert.Equal(t, uint64(0), period)
	})
}

func TestV2Reader_MemoryUsage(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.current"), []byte("104857600\n"), 0o644))

	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	usage, err := reader.MemoryUsage(context.Background())
	require.NoError(t, err)
	assert.Equal(t, uint64(104857600), usage) // 100MB
}

func TestV2Reader_MemoryLimit(t *testing.T) {
	t.Parallel()

	t.Run("parses limit", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.max"), []byte("1073741824\n"), 0o644))

		reader, err := cgroup.NewV2Reader(mockCgroup)
		require.NoError(t, err)

		limit, err := reader.MemoryLimit(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(1073741824), limit) // 1GB
	})

	t.Run("handles max (unlimited)", func(t *testing.T) {
		t.Parallel()

		mockCgroup := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.max"), []byte("max\n"), 0o644))

		reader, err := cgroup.NewV2Reader(mockCgroup)
		require.NoError(t, err)

		limit, err := reader.MemoryLimit(context.Background())
		require.NoError(t, err)
		assert.Equal(t, uint64(0), limit)
	})
}

func TestV2Reader_MemoryStat(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()
	memoryStat := `anon 52428800
file 41943040
kernel 8388608
slab 2097152
sock 1048576
shmem 4194304
mapped 10485760
dirty 524288
pgfault 12345
pgmajfault 67
`
	require.NoError(t, os.WriteFile(filepath.Join(mockCgroup, "memory.stat"), []byte(memoryStat), 0o644))

	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	stat, err := reader.ReadMemoryStat(context.Background())
	require.NoError(t, err)

	assert.Equal(t, uint64(52428800), stat.Anon)
	assert.Equal(t, uint64(41943040), stat.File)
	assert.Equal(t, uint64(8388608), stat.Kernel)
	assert.Equal(t, uint64(2097152), stat.Slab)
	assert.Equal(t, uint64(4194304), stat.Shmem)
	assert.Equal(t, uint64(12345), stat.Pgfault)
	assert.Equal(t, uint64(67), stat.Pgmajfault)
}

func TestV2Reader_ContextCancellation(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()
	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = reader.CPUUsage(ctx)
	assert.ErrorIs(t, err, context.Canceled)

	_, _, err = reader.CPULimit(ctx)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = reader.MemoryUsage(ctx)
	assert.ErrorIs(t, err, context.Canceled)

	_, err = reader.MemoryLimit(ctx)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestV2Reader_Path(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()
	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	assert.Equal(t, mockCgroup, reader.Path())
}

func TestNewV2Reader_InvalidPath(t *testing.T) {
	t.Parallel()

	_, err := cgroup.NewV2Reader("/nonexistent/path/that/should/not/exist")
	assert.Error(t, err)
}

func TestReader_Interface(t *testing.T) {
	t.Parallel()

	mockCgroup := t.TempDir()

	// V2Reader should implement Reader interface
	reader, err := cgroup.NewV2Reader(mockCgroup)
	require.NoError(t, err)

	// Verify interface compliance
	var _ cgroup.Reader = reader
	assert.Equal(t, mockCgroup, reader.Path())
}

func TestErrors(t *testing.T) {
	t.Parallel()

	assert.Error(t, cgroup.ErrUnknownVersion)
	assert.Error(t, cgroup.ErrPathNotFound)

	assert.Contains(t, cgroup.ErrUnknownVersion.Error(), "unknown")
	assert.Contains(t, cgroup.ErrPathNotFound.Error(), "not found")
}
