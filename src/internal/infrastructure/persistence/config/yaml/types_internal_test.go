// Package yaml provides internal tests for private helper methods.
package yaml

import (
	"testing"
	"time"
)

// Test_ProbeDTO_getThresholdDefaults tests getThresholdDefaults helper method.
func Test_ProbeDTO_getThresholdDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		probe                    *ProbeDTO
		expectedSuccessThreshold int
		expectedFailureThreshold int
	}{
		{
			name:                     "default values when zero",
			probe:                    &ProbeDTO{},
			expectedSuccessThreshold: 1,
			expectedFailureThreshold: 3,
		},
		{
			name:                     "custom values preserved",
			probe:                    &ProbeDTO{SuccessThreshold: 2, FailureThreshold: 5},
			expectedSuccessThreshold: 2,
			expectedFailureThreshold: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			success, failure := tt.probe.getThresholdDefaults()

			if success != tt.expectedSuccessThreshold {
				t.Errorf("expected success threshold %d, got %d", tt.expectedSuccessThreshold, success)
			}
			if failure != tt.expectedFailureThreshold {
				t.Errorf("expected failure threshold %d, got %d", tt.expectedFailureThreshold, failure)
			}
		})
	}
}

// Test_ProbeDTO_getTimingDefaults tests getTimingDefaults helper method.
func Test_ProbeDTO_getTimingDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		probe            *ProbeDTO
		expectedInterval time.Duration
		expectedTimeout  time.Duration
	}{
		{
			name:             "default values when zero",
			probe:            &ProbeDTO{},
			expectedInterval: 10 * time.Second,
			expectedTimeout:  5 * time.Second,
		},
		{
			name:             "custom values preserved",
			probe:            &ProbeDTO{Interval: Duration(30 * time.Second), Timeout: Duration(10 * time.Second)},
			expectedInterval: 30 * time.Second,
			expectedTimeout:  10 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			interval, timeout := tt.probe.getTimingDefaults()

			if interval != tt.expectedInterval {
				t.Errorf("expected interval %v, got %v", tt.expectedInterval, interval)
			}
			if timeout != tt.expectedTimeout {
				t.Errorf("expected timeout %v, got %v", tt.expectedTimeout, timeout)
			}
		})
	}
}

// Test_ProbeDTO_getHTTPDefaults tests getHTTPDefaults helper method.
func Test_ProbeDTO_getHTTPDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		probe              *ProbeDTO
		expectedMethod     string
		expectedStatusCode int
	}{
		{
			name:               "default values when empty",
			probe:              &ProbeDTO{},
			expectedMethod:     "GET",
			expectedStatusCode: 200,
		},
		{
			name:               "custom values preserved",
			probe:              &ProbeDTO{Method: "POST", StatusCode: 201},
			expectedMethod:     "POST",
			expectedStatusCode: 201,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			method, statusCode := tt.probe.getHTTPDefaults()

			if method != tt.expectedMethod {
				t.Errorf("expected method %s, got %s", tt.expectedMethod, method)
			}
			if statusCode != tt.expectedStatusCode {
				t.Errorf("expected status code %d, got %d", tt.expectedStatusCode, statusCode)
			}
		})
	}
}
