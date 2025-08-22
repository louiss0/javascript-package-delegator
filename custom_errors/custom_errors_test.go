package custom_errors_test

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/custom_errors"
)

func TestCustomErrors(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Errors Suite")
}

var _ = Describe("FlagName", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Describe("Error method", func() {
		Context("when flag name is valid", func() {
			It("should return nil for alphanumeric flag names", func() {
				flag := custom_errors.FlagName("validflag123")
				err := flag.Error()
				assertT.NoError(err)
			})

			It("should return nil for all letters flag names", func() {
				flag := custom_errors.FlagName("anotherflag")
				err := flag.Error()
				assertT.NoError(err)
			})

			It("should return nil for all numbers flag names", func() {
				flag := custom_errors.FlagName("12345")
				err := flag.Error()
				assertT.NoError(err)
			})
		})

		Context("when flag name is invalid", func() {
			It("should return error for flag names with hyphen", func() {
				flag := custom_errors.FlagName("invalid-flag")
				err := flag.Error()
				assertT.Error(err)
				expectedErr := fmt.Errorf("%w: invalid-flag must be alphanumeric: invalid-flag", custom_errors.ErrInvalidFlag)
				assertT.Equal(expectedErr.Error(), err.Error())
			})

			It("should return error for flag names with underscore", func() {
				flag := custom_errors.FlagName("invalid_flag")
				err := flag.Error()
				assertT.Error(err)
				expectedErr := fmt.Errorf("%w: invalid_flag must be alphanumeric: invalid_flag", custom_errors.ErrInvalidFlag)
				assertT.Equal(expectedErr.Error(), err.Error())
			})

			It("should return error for flag names with special characters", func() {
				flag := custom_errors.FlagName("flag!@#")
				err := flag.Error()
				assertT.Error(err)
				expectedErr := fmt.Errorf("%w: flag!@# must be alphanumeric: flag!@#", custom_errors.ErrInvalidFlag)
				assertT.Equal(expectedErr.Error(), err.Error())
			})

			It("should return error for empty flag name", func() {
				flag := custom_errors.FlagName("")
				err := flag.Error()
				assertT.Error(err)
				expectedErr := fmt.Errorf("%w:  must be alphanumeric: ", custom_errors.ErrInvalidFlag)
				assertT.Equal(expectedErr.Error(), err.Error())
			})
		})
	})
})

var _ = Describe("CreateInvalidFlagErrorWithMessage", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Context("when creating custom flag error messages", func() {
		It("should create error with custom message for valid flag", func() {
			flagName := custom_errors.FlagName("validflag")
			message := "custom error message"
			err := custom_errors.CreateInvalidFlagErrorWithMessage(flagName, message)
			expectedErr := fmt.Errorf("%w: validflag custom error message", custom_errors.ErrInvalidFlag)
			assertT.Equal(expectedErr.Error(), err.Error())
		})

		It("should ignore custom message for invalid flag and use alphanumeric error", func() {
			flagName := custom_errors.FlagName("invalid-flag")
			message := "this should not appear"
			err := custom_errors.CreateInvalidFlagErrorWithMessage(flagName, message)
			expectedErr := fmt.Errorf("%w: invalid-flag must be alphanumeric: invalid-flag", custom_errors.ErrInvalidFlag)
			assertT.Equal(expectedErr.Error(), err.Error())
		})

		It("should handle empty message", func() {
			flagName := custom_errors.FlagName("anotherflag")
			message := ""
			err := custom_errors.CreateInvalidFlagErrorWithMessage(flagName, message)
			expectedErr := fmt.Errorf("%w: anotherflag ", custom_errors.ErrInvalidFlag)
			assertT.Equal(expectedErr.Error(), err.Error())
		})
	})
})

var _ = Describe("CreateInvalidArgumentErrorWithMessage", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Context("when creating argument error messages", func() {
		It("should create standard argument error", func() {
			message := "test message"
			err := custom_errors.CreateInvalidArgumentErrorWithMessage(message)
			expectedErr := fmt.Errorf("%w: test message", custom_errors.ErrInvalidArgument)
			assertT.Equal(expectedErr.Error(), err.Error())
		})

		It("should handle empty message", func() {
			message := ""
			err := custom_errors.CreateInvalidArgumentErrorWithMessage(message)
			expectedErr := fmt.Errorf("%w: ", custom_errors.ErrInvalidArgument)
			assertT.Equal(expectedErr.Error(), err.Error())
		})

		It("should handle message with special characters", func() {
			message := "message with !@#$"
			err := custom_errors.CreateInvalidArgumentErrorWithMessage(message)
			expectedErr := fmt.Errorf("%w: message with !@#$", custom_errors.ErrInvalidArgument)
			assertT.Equal(expectedErr.Error(), err.Error())
		})
	})
})

var _ = Describe("Sentinel Errors", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Context("when using errors.Is for error detection", func() {
		It("should detect ErrInvalidFlag with errors.Is", func() {
			err := custom_errors.CreateInvalidFlagErrorWithMessage("myflag", "some message")
			assertT.True(errors.Is(err, custom_errors.ErrInvalidFlag), "Expected error to be ErrInvalidFlag")
		})

		It("should detect ErrInvalidArgument with errors.Is", func() {
			err := custom_errors.CreateInvalidArgumentErrorWithMessage("another message")
			assertT.True(errors.Is(err, custom_errors.ErrInvalidArgument), "Expected error to be ErrInvalidArgument")
		})

		It("should not match custom error sentinels for generic errors", func() {
			genericErr := errors.New("some other error")
			assertT.False(errors.Is(genericErr, custom_errors.ErrInvalidFlag), "Generic error should not be ErrInvalidFlag")
			assertT.False(errors.Is(genericErr, custom_errors.ErrInvalidArgument), "Generic error should not be ErrInvalidArgument")
		})
	})
})
