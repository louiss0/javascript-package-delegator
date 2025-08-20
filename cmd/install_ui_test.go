package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/services"
)

func TestNewPackageMultiSelectUI(t *testing.T) {
	t.Run("creates MultiUISelecter with package info", func(t *testing.T) {
		packageInfo := []services.PackageInfo{
			{Name: "react", Version: "18.2.0"},
			{Name: "vue", Version: "3.2.45"},
			{Name: "angular", Version: "15.0.0"},
		}

		result := newPackageMultiSelectUI(packageInfo)

		assert.NotNil(t, result, "newPackageMultiSelectUI should return non-nil")

		// Type assert to verify it's the correct type
		ui, ok := result.(*packageMultiSelectUI)
		assert.True(t, ok, "Result should be type *packageMultiSelectUI")
		assert.NotNil(t, ui.multiSelectUI, "Internal multiSelectUI should be initialized")
	})

	t.Run("creates empty MultiUISelecter with no packages", func(t *testing.T) {
		packageInfo := []services.PackageInfo{}

		result := newPackageMultiSelectUI(packageInfo)

		assert.NotNil(t, result, "newPackageMultiSelectUI should return non-nil even with empty input")

		ui, ok := result.(*packageMultiSelectUI)
		assert.True(t, ok, "Result should be type *packageMultiSelectUI")
		assert.NotNil(t, ui.multiSelectUI, "Internal multiSelectUI should be initialized")
	})
}

func TestPackageMultiSelectUI_Values(t *testing.T) {
	t.Run("returns empty slice by default", func(t *testing.T) {
		packageInfo := []services.PackageInfo{
			{Name: "react", Version: "18.2.0"},
		}

		result := newPackageMultiSelectUI(packageInfo)
		ui := result.(*packageMultiSelectUI)

		values := ui.Values()
		assert.Empty(t, values, "Values() should return empty slice by default")
	})

	t.Run("returns set values", func(t *testing.T) {
		packageInfo := []services.PackageInfo{
			{Name: "react", Version: "18.2.0"},
			{Name: "vue", Version: "3.2.45"},
		}

		result := newPackageMultiSelectUI(packageInfo)
		ui := result.(*packageMultiSelectUI)

		// Manually set values to simulate user selection
		ui.value = []string{"react@18.2.0", "vue@3.2.45"}

		values := ui.Values()
		expected := []string{"react@18.2.0", "vue@3.2.45"}
		assert.Equal(t, expected, values, "Values() should return the set values")
	})
}

func TestPackageMultiSelectUI_Run(t *testing.T) {
	t.Run("returns error in non-interactive environment", func(t *testing.T) {
		packageInfo := []services.PackageInfo{
			{Name: "react", Version: "18.2.0"},
		}

		result := newPackageMultiSelectUI(packageInfo)
		ui := result.(*packageMultiSelectUI)

		// In a non-interactive test environment (no TTY), Run() should return an error
		// This exercises the code path without requiring user input
		err := ui.Run()
		assert.Error(t, err, "Run() should return error in non-interactive environment")
		// We don't check the specific error message as it may vary based on the TUI library
	})
}
