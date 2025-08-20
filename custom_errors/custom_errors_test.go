package custom_errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestFlagName_Error(t *testing.T) {
	tests := []struct {
		name        string
		flag        FlagName
		wantErr     bool
		expectedErr error
	}{
		{
			name:    "valid flag name (alphanumeric)",
			flag:    "validflag123",
			wantErr: false,
		},
		{
			name:    "valid flag name (all letters)",
			flag:    "anotherflag",
			wantErr: false,
		},
		{
			name:    "valid flag name (all numbers)",
			flag:    "12345",
			wantErr: false,
		},
		{
			name:        "invalid flag name (with hyphen)",
			flag:        "invalid-flag",
			wantErr:     true,
			expectedErr: fmt.Errorf("%w: invalid-flag must be alphanumeric: invalid-flag", ErrInvalidFlag),
		},
		{
			name:        "invalid flag name (with underscore)",
			flag:        "invalid_flag",
			wantErr:     true,
			expectedErr: fmt.Errorf("%w: invalid_flag must be alphanumeric: invalid_flag", ErrInvalidFlag),
		},
		{
			name:        "invalid flag name (with special chars)",
			flag:        "flag!@#",
			wantErr:     true,
			expectedErr: fmt.Errorf("%w: flag!@# must be alphanumeric: flag!@#", ErrInvalidFlag),
		},
		{
			name:        "empty flag name",
			flag:        "",
			wantErr:     true,
			expectedErr: fmt.Errorf("%w:  must be alphanumeric: ", ErrInvalidFlag),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flag.Error()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Error() expected an error, but got nil")
				} else if err.Error() != tt.expectedErr.Error() {
					t.Errorf("Error() got = %v, want %v", err, tt.expectedErr)
				}
			} else if err != nil {
				t.Errorf("Error() expected no error, but got %v", err)
			}
		})
	}
}

func TestCreateInvalidFlagErrorWithMessage(t *testing.T) {
	tests := []struct {
		name     string
		flagName FlagName
		message  string
		wantErr  error
	}{
		{
			name:     "valid flag with custom message",
			flagName: "validflag",
			message:  "custom error message",
			wantErr:  fmt.Errorf("%w: validflag custom error message", ErrInvalidFlag),
		},
		{
			name:     "invalid flag with custom message",
			flagName: "invalid-flag",
			message:  "this should not appear",
			wantErr:  fmt.Errorf("%w: invalid-flag must be alphanumeric: invalid-flag", ErrInvalidFlag),
		},
		{
			name:     "empty message",
			flagName: "anotherflag",
			message:  "",
			wantErr:  fmt.Errorf("%w: anotherflag ", ErrInvalidFlag),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateInvalidFlagErrorWithMessage(tt.flagName, tt.message)
			if err.Error() != tt.wantErr.Error() {
				t.Errorf("CreateInvalidFlagErrorWithMessage() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestCreateInvalidArgumentErrorWithMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr error
	}{
		{
			name:    "standard argument error",
			message: "test message",
			wantErr: fmt.Errorf("%w: test message", ErrInvalidArgument),
		},
		{
			name:    "empty message",
			message: "",
			wantErr: fmt.Errorf("%w: ", ErrInvalidArgument),
		},
		{
			name:    "message with special characters",
			message: "message with !@#$",
			wantErr: fmt.Errorf("%w: message with !@#$", ErrInvalidArgument),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CreateInvalidArgumentErrorWithMessage(tt.message)
			if err.Error() != tt.wantErr.Error() {
				t.Errorf("CreateInvalidArgumentErrorWithMessage() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	t.Run("ErrInvalidFlag should be detectable with errors.Is", func(t *testing.T) {
		err := CreateInvalidFlagErrorWithMessage("myflag", "some message")
		if !errors.Is(err, ErrInvalidFlag) {
			t.Errorf("Expected error to be ErrInvalidFlag, but it was not")
		}
	})

	t.Run("ErrInvalidArgument should be detectable with errors.Is", func(t *testing.T) {
		err := CreateInvalidArgumentErrorWithMessage("another message")
		if !errors.Is(err, ErrInvalidArgument) {
			t.Errorf("Expected error to be ErrInvalidArgument, but it was not")
		}
	})

	t.Run("A generic error should not match custom error sentinels", func(t *testing.T) {
		genericErr := errors.New("some other error")
		if errors.Is(genericErr, ErrInvalidFlag) {
			t.Errorf("Generic error should not be ErrInvalidFlag")
		}
		if errors.Is(genericErr, ErrInvalidArgument) {
			t.Errorf("Generic error should not be ErrInvalidArgument")
		}
	})
}
