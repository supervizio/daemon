package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvertResultToError(t *testing.T) {
	tests := []struct {
		name    string
		success bool
		code    int
		message string
		wantErr bool
		wantIs  error
	}{
		{
			name:    "ReturnsNilOnSuccess",
			success: true,
			code:    0,
			message: "",
			wantErr: false,
			wantIs:  nil,
		},
		{
			name:    "ReturnsNotSupportedForCode1",
			success: false,
			code:    1,
			message: "",
			wantErr: true,
			wantIs:  ErrNotSupported,
		},
		{
			name:    "ReturnsPermissionForCode2",
			success: false,
			code:    2,
			message: "",
			wantErr: true,
			wantIs:  ErrPermission,
		},
		{
			name:    "ReturnsNotFoundForCode3",
			success: false,
			code:    3,
			message: "",
			wantErr: true,
			wantIs:  ErrNotFound,
		},
		{
			name:    "ReturnsInvalidParamForCode4",
			success: false,
			code:    4,
			message: "",
			wantErr: true,
			wantIs:  ErrInvalidParam,
		},
		{
			name:    "ReturnsIOErrorForCode5",
			success: false,
			code:    5,
			message: "",
			wantErr: true,
			wantIs:  ErrIO,
		},
		{
			name:    "ReturnsInternalForCode99",
			success: false,
			code:    99,
			message: "",
			wantErr: true,
			wantIs:  ErrInternal,
		},
		{
			name:    "ReturnsCustomErrorForUnknownCodeWithMessage",
			success: false,
			code:    999,
			message: "custom error",
			wantErr: true,
			wantIs:  nil,
		},
		{
			name:    "ReturnsInternalForUnknownCodeWithoutMessage",
			success: false,
			code:    999,
			message: "",
			wantErr: true,
			wantIs:  ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := convertResultToError(tt.success, tt.code, tt.message)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantIs != nil {
					assert.ErrorIs(t, err, tt.wantIs)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBytesToStringWithNull(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		wantStr string
	}{
		{
			name:    "ConvertsNullTerminatedString",
			input:   []byte{'h', 'e', 'l', 'l', 'o', 0, 'x'},
			wantStr: "hello",
		},
		{
			name:    "ConvertsFullArrayWithoutNull",
			input:   []byte{'a', 'b', 'c'},
			wantStr: "abc",
		},
		{
			name:    "ConvertsEmptyString",
			input:   []byte{0, 'x', 'y'},
			wantStr: "",
		},
		{
			name:    "ConvertsSingleChar",
			input:   []byte{'x', 0},
			wantStr: "x",
		},
		{
			name:    "ConvertsEmptySlice",
			input:   []byte{},
			wantStr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := bytesToStringWithNull(tt.input)
			assert.Equal(t, tt.wantStr, result)
		})
	}
}

func TestProbeCodeConstants(t *testing.T) {
	tests := []struct {
		name      string
		code      int
		wantValue int
	}{
		{
			name:      "ProbeCodeOKIs0",
			code:      probeCodeOK,
			wantValue: 0,
		},
		{
			name:      "ProbeCodeNotSupportedIs1",
			code:      probeCodeNotSupported,
			wantValue: 1,
		},
		{
			name:      "ProbeCodePermissionIs2",
			code:      probeCodePermission,
			wantValue: 2,
		},
		{
			name:      "ProbeCodeNotFoundIs3",
			code:      probeCodeNotFound,
			wantValue: 3,
		},
		{
			name:      "ProbeCodeInvalidParamIs4",
			code:      probeCodeInvalidParam,
			wantValue: 4,
		},
		{
			name:      "ProbeCodeIOIs5",
			code:      probeCodeIO,
			wantValue: 5,
		},
		{
			name:      "ProbeCodeInternalIs99",
			code:      probeCodeInternal,
			wantValue: 99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantValue, tt.code)
		})
	}
}
