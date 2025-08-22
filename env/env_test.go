package env_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/env"
)

func TestEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Env Suite")
}

var _ = Describe("GoEnv", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Describe("NewGoEnv", func() {
		It("should initialize GoEnv with build_info.GO_MODE", func() {
			goEnv := env.NewGoEnv()
			expected := build_info.GO_MODE.String()
			assertT.Equal(expected, goEnv.Mode())
		})
	})

	Describe("Mode", func() {
		Context("when getting mode from initialized GoEnv", func() {
			It("should return the current build mode", func() {
				goEnv := env.NewGoEnv()
				result := goEnv.Mode()
				// The actual mode depends on build_info.GO_MODE
				expected := build_info.GO_MODE.String()
				assertT.Equal(expected, result)
			})
		})
	})

	Describe("IsDebugMode", func() {
		Context("when GoEnv is in debug mode", func() {
			It("should return true", func() {
				// We test the actual behavior based on build_info
				goEnv := env.NewGoEnv()
				isDebug := goEnv.IsDebugMode()
				// The result depends on the actual build_info.GO_MODE value
				expectedDebug := build_info.GO_MODE.String() == "debug"
				assertT.Equal(expectedDebug, isDebug)
			})
		})
	})

	Describe("IsDevelopmentMode", func() {
		Context("when evaluating development mode", func() {
			It("should return correct development mode status", func() {
				goEnv := env.NewGoEnv()
				isDevelopment := goEnv.IsDevelopmentMode()
				// Development mode returns true for "development" and empty mode (default)
				currentMode := build_info.GO_MODE.String()
				expectedDevelopment := currentMode == "development" || currentMode == ""
				assertT.Equal(expectedDevelopment, isDevelopment)
			})
		})
	})

	Describe("IsProductionMode", func() {
		Context("when evaluating production mode", func() {
			It("should return correct production mode status", func() {
				goEnv := env.NewGoEnv()
				isProduction := goEnv.IsProductionMode()
				// Production mode returns true only for "production"
				expectedProduction := build_info.GO_MODE.String() == "production"
				assertT.Equal(expectedProduction, isProduction)
			})
		})
	})

	Describe("ExecuteIfModeIsProduction", func() {
		Context("when executing conditional callback", func() {
			It("should execute callback only in production mode", func() {
				goEnv := env.NewGoEnv()
				executed := false
				callback := func() {
					executed = true
				}

				goEnv.ExecuteIfModeIsProduction(callback)

				// Should execute only if current mode is production
				expectedExecution := build_info.GO_MODE.String() == "production"
				assertT.Equal(expectedExecution, executed)
			})
		})

		Context("when callback panics in production mode", func() {
			It("should propagate the panic", func() {
				// Only test panic propagation if we're actually in production mode
				if build_info.GO_MODE.String() == "production" {
					goEnv := env.NewGoEnv()
					assertT.Panics(func() {
						goEnv.ExecuteIfModeIsProduction(func() {
							panic("test panic")
						})
					}, "Should panic when callback panics in production mode")
				} else {
					// If not in production mode, callback won't execute, so no panic
					goEnv := env.NewGoEnv()
					assertT.NotPanics(func() {
						goEnv.ExecuteIfModeIsProduction(func() {
							panic("test panic")
						})
					}, "Should not panic when not in production mode")
				}
			})
		})
	})
})
