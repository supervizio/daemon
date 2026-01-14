package probe_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/probe"
)

func TestSystemCPU_Total(t *testing.T) {
	cpu := probe.SystemCPU{
		User:      1000,
		Nice:      100,
		System:    500,
		Idle:      8000,
		IOWait:    200,
		IRQ:       50,
		SoftIRQ:   50,
		Steal:     0,
		Guest:     0,
		GuestNice: 0,
	}

	total := cpu.Total()
	assert.Equal(t, uint64(9900), total)
}

func TestSystemCPU_Active(t *testing.T) {
	cpu := probe.SystemCPU{
		User:   1000,
		Nice:   100,
		System: 500,
		Idle:   8000,
		IOWait: 200,
	}

	active := cpu.Active()
	expected := uint64(1000 + 100 + 500) // Total - Idle - IOWait
	assert.Equal(t, expected, active)
}

func TestProcessCPU_Total(t *testing.T) {
	proc := probe.ProcessCPU{
		PID:    1234,
		Name:   "test",
		User:   500,
		System: 300,
	}

	total := proc.Total()
	assert.Equal(t, uint64(800), total)
}

func TestProcessCPU_TotalWithChildren(t *testing.T) {
	proc := probe.ProcessCPU{
		PID:            1234,
		Name:           "test",
		User:           500,
		System:         300,
		ChildrenUser:   100,
		ChildrenSystem: 50,
	}

	total := proc.TotalWithChildren()
	assert.Equal(t, uint64(950), total)
}

func TestSystemCPU_Timestamp(t *testing.T) {
	now := time.Now()
	cpu := probe.SystemCPU{
		Timestamp: now,
	}

	assert.Equal(t, now, cpu.Timestamp)
}

func TestProcessCPU_Fields(t *testing.T) {
	proc := probe.ProcessCPU{
		PID:          1234,
		Name:         "myprocess",
		User:         100,
		System:       50,
		StartTime:    12345678,
		UsagePercent: 5.5,
		Timestamp:    time.Now(),
	}

	assert.Equal(t, 1234, proc.PID)
	assert.Equal(t, "myprocess", proc.Name)
	assert.Equal(t, uint64(100), proc.User)
	assert.Equal(t, uint64(50), proc.System)
	assert.Equal(t, uint64(12345678), proc.StartTime)
	assert.Equal(t, 5.5, proc.UsagePercent)
}
