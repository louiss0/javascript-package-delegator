package env

import (
	"testing"

	"github.com/louiss0/javascript-package-delegator/build_info"
)

func TestNewGoEnv(t *testing.T) {
	goEnv := NewGoEnv()

	// Test that the GoEnv is initialized with build_info.GO_MODE
	expected := build_info.GO_MODE.String()
	if goEnv.goEnv != expected {
		t.Errorf("Expected GoEnv.goEnv to be %s, got %s", expected, goEnv.goEnv)
	}
}

func TestGoEnv_Mode(t *testing.T) {
	tests := []struct {
		name     string
		goEnv    GoEnv
		expected string
	}{
		{
			name:     "production mode",
			goEnv:    GoEnv{goEnv: "production"},
			expected: "production",
		},
		{
			name:     "development mode",
			goEnv:    GoEnv{goEnv: "development"},
			expected: "development",
		},
		{
			name:     "debug mode",
			goEnv:    GoEnv{goEnv: "debug"},
			expected: "debug",
		},
		{
			name:     "empty mode",
			goEnv:    GoEnv{goEnv: ""},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goEnv.Mode()
			if result != tt.expected {
				t.Errorf("Mode() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestGoEnv_IsDebugMode(t *testing.T) {
	tests := []struct {
		name     string
		goEnv    GoEnv
		expected bool
	}{
		{
			name:     "debug mode returns true",
			goEnv:    GoEnv{goEnv: "debug"},
			expected: true,
		},
		{
			name:     "production mode returns false",
			goEnv:    GoEnv{goEnv: "production"},
			expected: false,
		},
		{
			name:     "development mode returns false",
			goEnv:    GoEnv{goEnv: "development"},
			expected: false,
		},
		{
			name:     "empty mode returns false",
			goEnv:    GoEnv{goEnv: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goEnv.IsDebugMode()
			if result != tt.expected {
				t.Errorf("IsDebugMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGoEnv_IsDevelopmentMode(t *testing.T) {
	tests := []struct {
		name     string
		goEnv    GoEnv
		expected bool
	}{
		{
			name:     "development mode returns true",
			goEnv:    GoEnv{goEnv: "development"},
			expected: true,
		},
		{
			name:     "empty mode returns true (default)",
			goEnv:    GoEnv{goEnv: ""},
			expected: true,
		},
		{
			name:     "production mode returns false",
			goEnv:    GoEnv{goEnv: "production"},
			expected: false,
		},
		{
			name:     "debug mode returns false",
			goEnv:    GoEnv{goEnv: "debug"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goEnv.IsDevelopmentMode()
			if result != tt.expected {
				t.Errorf("IsDevelopmentMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGoEnv_IsProductionMode(t *testing.T) {
	tests := []struct {
		name     string
		goEnv    GoEnv
		expected bool
	}{
		{
			name:     "production mode returns true",
			goEnv:    GoEnv{goEnv: "production"},
			expected: true,
		},
		{
			name:     "development mode returns false",
			goEnv:    GoEnv{goEnv: "development"},
			expected: false,
		},
		{
			name:     "debug mode returns false",
			goEnv:    GoEnv{goEnv: "debug"},
			expected: false,
		},
		{
			name:     "empty mode returns false",
			goEnv:    GoEnv{goEnv: ""},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.goEnv.IsProductionMode()
			if result != tt.expected {
				t.Errorf("IsProductionMode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGoEnv_ExecuteIfModeIsProduction(t *testing.T) {
	tests := []struct {
		name      string
		goEnv     GoEnv
		shouldRun bool
	}{
		{
			name:      "executes callback in production mode",
			goEnv:     GoEnv{goEnv: "production"},
			shouldRun: true,
		},
		{
			name:      "does not execute callback in development mode",
			goEnv:     GoEnv{goEnv: "development"},
			shouldRun: false,
		},
		{
			name:      "does not execute callback in debug mode",
			goEnv:     GoEnv{goEnv: "debug"},
			shouldRun: false,
		},
		{
			name:      "does not execute callback in empty mode",
			goEnv:     GoEnv{goEnv: ""},
			shouldRun: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executed := false
			callback := func() {
				executed = true
			}

			tt.goEnv.ExecuteIfModeIsProduction(callback)

			if executed != tt.shouldRun {
				t.Errorf("ExecuteIfModeIsProduction() executed callback = %v, want %v", executed, tt.shouldRun)
			}
		})
	}
}

func TestGoEnv_ExecuteIfModeIsProduction_CallbackPanic(t *testing.T) {
	// Test that panics in callback are properly propagated
	goEnv := GoEnv{goEnv: "production"}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic but none occurred")
		} else if r != "test panic" {
			t.Errorf("Expected panic with 'test panic', got %v", r)
		}
	}()

	goEnv.ExecuteIfModeIsProduction(func() {
		panic("test panic")
	})
}
