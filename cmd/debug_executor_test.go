package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugExecutor_ExecuteIfDebugIsTrue(t *testing.T) {
	t.Run("does not execute when debug is false", func(t *testing.T) {
		executor := newDebugExecutor(false)
		counter := 0

		executor.ExecuteIfDebugIsTrue(func() {
			counter++
		})

		assert.Equal(t, 0, counter, "Function should not be executed when debug is false")
	})

	t.Run("executes when debug is true", func(t *testing.T) {
		executor := newDebugExecutor(true)
		counter := 0

		executor.ExecuteIfDebugIsTrue(func() {
			counter++
		})

		assert.Equal(t, 1, counter, "Function should be executed when debug is true")
	})

	t.Run("multiple executions work correctly", func(t *testing.T) {
		executorTrue := newDebugExecutor(true)
		executorFalse := newDebugExecutor(false)
		counter := 0

		incrementFunc := func() { counter++ }

		executorTrue.ExecuteIfDebugIsTrue(incrementFunc)
		executorFalse.ExecuteIfDebugIsTrue(incrementFunc)
		executorTrue.ExecuteIfDebugIsTrue(incrementFunc)

		assert.Equal(t, 2, counter, "Only debug=true executions should increment counter")
	})
}

func TestDebugExecutor_LogDebugMessageIfDebugIsTrue(t *testing.T) {
	t.Run("does not panic when debug is false", func(t *testing.T) {
		executor := newDebugExecutor(false)

		assert.NotPanics(t, func() {
			executor.LogDebugMessageIfDebugIsTrue("test message", "key", "value")
		})
	})

	t.Run("does not panic when debug is true", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogDebugMessageIfDebugIsTrue("test message", "key", "value")
		})
	})

	t.Run("handles empty keyvals", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogDebugMessageIfDebugIsTrue("test message")
		})
	})

	t.Run("handles odd number of keyvals", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogDebugMessageIfDebugIsTrue("test message", "key")
		})
	})
}

func TestDebugExecutor_LogJSCommandIfDebugIsTrue(t *testing.T) {
	t.Run("does not panic when debug is false", func(t *testing.T) {
		executor := newDebugExecutor(false)

		assert.NotPanics(t, func() {
			executor.LogJSCommandIfDebugIsTrue("npm", "install", "react")
		})
	})

	t.Run("does not panic when debug is true", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogJSCommandIfDebugIsTrue("npm", "install", "react")
		})
	})

	t.Run("handles command without args", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogJSCommandIfDebugIsTrue("npm")
		})
	})

	t.Run("handles empty command", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotPanics(t, func() {
			executor.LogJSCommandIfDebugIsTrue("")
		})
	})
}

func TestNewDebugExecutor(t *testing.T) {
	t.Run("creates debug executor with false flag", func(t *testing.T) {
		executor := newDebugExecutor(false)

		assert.NotNil(t, executor)
		
		// Verify the debug flag is set correctly by testing behavior
		counter := 0
		executor.ExecuteIfDebugIsTrue(func() { counter++ })
		assert.Equal(t, 0, counter, "Debug executor with false flag should not execute callbacks")
	})

	t.Run("creates debug executor with true flag", func(t *testing.T) {
		executor := newDebugExecutor(true)

		assert.NotNil(t, executor)
		
		// Verify the debug flag is set correctly by testing behavior
		counter := 0
		executor.ExecuteIfDebugIsTrue(func() { counter++ })
		assert.Equal(t, 1, counter, "Debug executor with true flag should execute callbacks")
	})
}
