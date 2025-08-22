package custom_flags_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/build_info"
	"github.com/louiss0/javascript-package-delegator/custom_flags"
)

func TestCustomFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Flags Suite")
}

var _ = Describe("FilePathFlag", func() {
	var (
		flag    custom_flags.FilePathFlagInterface
		assertT *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		flagVal := custom_flags.NewFilePathFlag("testflag")
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have string type", func() {
			assertT.Equal("string", flag.Type())
		})

		It("should initialize with empty value", func() {
			assertT.Equal("", flag.String())
		})
	})

	Describe("Set method", func() {
		Context("when provided valid file paths", func() {
			It("should accept valid absolute path", func() {
				err := flag.Set("/path/to/file.txt")
				assertT.NoError(err)
				assertT.Equal("/path/to/file.txt", flag.String())
			})

			It("should accept valid relative path", func() {
				err := flag.Set("file.txt")
				assertT.NoError(err)
				assertT.Equal("file.txt", flag.String())
			})

			It("should accept path with dots", func() {
				err := flag.Set("../dir/file.log")
				assertT.NoError(err)
				assertT.Equal("../dir/file.log", flag.String())
			})

			It("should accept path with underscores", func() {
				err := flag.Set("my_file.txt")
				assertT.NoError(err)
				assertT.Equal("my_file.txt", flag.String())
			})

			It("should accept path with hyphens", func() {
				err := flag.Set("my-file.txt")
				assertT.NoError(err)
				assertT.Equal("my-file.txt", flag.String())
			})
		})

		Context("when provided invalid file paths", func() {
			It("should reject empty string", func() {
				err := flag.Set("")
				assertT.Error(err)
				assertT.Contains(err.Error(), "cannot be empty")
			})

			It("should reject whitespace only", func() {
				err := flag.Set("   ")
				assertT.Error(err)
				assertT.Contains(err.Error(), "cannot be empty")
			})

			It("should reject path with double slash", func() {
				err := flag.Set("path//file.txt")
				assertT.Error(err)
				assertT.Contains(err.Error(), "not a valid POSIX/UNIX file path")
			})

			It("should reject path with trailing slash", func() {
				err := flag.Set("path/")
				assertT.Error(err)
				assertT.Contains(err.Error(), "not a valid POSIX/UNIX file path")
			})
		})
	})
})

var _ = Describe("FolderPathFlag", func() {
	var (
		flag    custom_flags.FolderPathFlagInterface
		assertT *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		flagVal := custom_flags.NewFolderPathFlag("testflag")
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have string type", func() {
			assertT.Equal("string", flag.Type())
		})

		It("should initialize with empty value", func() {
			assertT.Equal("", flag.String())
		})
	})

	Describe("Set method", func() {
		Context("when provided valid folder paths", func() {
			It("should accept valid absolute path with slash", func() {
				err := flag.Set("/path/to/dir/")
				assertT.NoError(err)
				assertT.Equal("/path/to/dir/", flag.String())
			})

			It("should accept valid relative path with slash", func() {
				err := flag.Set("dir/")
				assertT.NoError(err)
				assertT.Equal("dir/", flag.String())
			})

			It("should accept root path", func() {
				err := flag.Set("/")
				assertT.NoError(err)
				assertT.Equal("/", flag.String())
			})

			Context("when in CI mode", func() {
				It("should accept path without trailing slash if InCI returns true", func() {
					if build_info.InCI() {
						err := flag.Set("/path/to/dir")
						assertT.NoError(err)
						assertT.Equal("/path/to/dir", flag.String())
					} else {
						Skip("Skipping CI-specific test when not in CI mode")
					}
				})
			})

			Context("when not in CI mode", func() {
				It("should reject path without trailing slash if InCI returns false", func() {
					if !build_info.InCI() {
						err := flag.Set("/path/to/dir")
						assertT.Error(err)
						assertT.Contains(err.Error(), "must end with '/'")
					} else {
						Skip("Skipping non-CI test when in CI mode")
					}
				})
			})
		})

		Context("when provided invalid folder paths", func() {
			It("should reject file-like paths", func() {
				err := flag.Set("/path/to/file.txt")
				assertT.Error(err)
				assertT.Contains(err.Error(), "not a valid POSIX/UNIX folder path")
			})

			It("should reject empty string", func() {
				err := flag.Set("")
				assertT.Error(err)
				assertT.Contains(err.Error(), "cannot be empty")
			})

			It("should reject whitespace only", func() {
				err := flag.Set("   ")
				assertT.Error(err)
				assertT.Contains(err.Error(), "cannot be empty")
			})
		})
	})
})

var _ = Describe("EmptyStringFlag", func() {
	var (
		flag    custom_flags.EmptyStringFlagInterface
		assertT *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		flagVal := custom_flags.NewEmptyStringFlag("testflag")
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have string type", func() {
			assertT.Equal("string", flag.Type())
		})

		It("should initialize with empty value", func() {
			assertT.Equal("", flag.String())
		})
	})

	Describe("Set method", func() {
		Context("when provided valid strings", func() {
			It("should accept valid non-empty string", func() {
				err := flag.Set("valid value")
				assertT.NoError(err)
				assertT.Equal("valid value", flag.String())
			})

			It("should accept empty string", func() {
				err := flag.Set("")
				assertT.NoError(err)
				assertT.Equal("", flag.String())
			})

			It("should accept string with content and spaces", func() {
				err := flag.Set("  content  ")
				assertT.NoError(err)
				assertT.Equal("  content  ", flag.String())
			})
		})

		Context("when provided invalid strings", func() {
			It("should reject whitespace-only strings", func() {
				err := flag.Set("   ")
				assertT.Error(err)
				assertT.Contains(err.Error(), "flag is empty")
			})
		})
	})
})

var _ = Describe("BoolFlag", func() {
	var (
		flag    custom_flags.BoolFlagInterface
		assertT *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		flagVal := custom_flags.NewBoolFlag("testflag")
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have bool type", func() {
			assertT.Equal("bool", flag.Type())
		})

		It("should initialize with empty value", func() {
			assertT.Equal("", flag.String())
		})
	})

	Describe("Set method", func() {
		Context("when provided valid boolean values", func() {
			It("should accept 'true'", func() {
				err := flag.Set("true")
				assertT.NoError(err)
				assertT.Equal("true", flag.String())
				assertT.True(flag.Value())
			})

			It("should accept 'false'", func() {
				err := flag.Set("false")
				assertT.NoError(err)
				assertT.Equal("false", flag.String())
				assertT.False(flag.Value())
			})

			It("should accept empty string", func() {
				err := flag.Set("")
				assertT.NoError(err)
				assertT.Equal("", flag.String())
			})
		})

		Context("when provided invalid boolean values", func() {
			It("should reject 'yes'", func() {
				err := flag.Set("yes")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be one of")
			})

			It("should reject '1'", func() {
				err := flag.Set("1")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be one of")
			})
		})
	})

	Describe("Value method", func() {
		It("should return correct boolean for 'true'", func() {
			_ = flag.Set("true")
			assertT.True(flag.Value())
		})

		It("should return correct boolean for 'false'", func() {
			_ = flag.Set("false")
			assertT.False(flag.Value())
		})

		It("should return false for invalid value", func() {
			// We need to manually set an invalid value for this test
			// This simulates what happens when ParseBool fails
			_ = flag.Set("") // Empty string parses as false
			assertT.False(flag.Value())
		})
	})
})

var _ = Describe("UnionFlag", func() {
	var (
		flag        custom_flags.UnionFlagInterface
		allowedVals []string
		assertT     *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		allowedVals = []string{"option1", "option2", "option3"}
		flagVal := custom_flags.NewUnionFlag(allowedVals, "testflag")
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have string type", func() {
			assertT.Equal("string", flag.Type())
		})

		It("should store allowed values", func() {
			assertT.Equal(allowedVals, flag.AllowedValues())
		})

		It("should initialize with empty value", func() {
			assertT.Equal("", flag.String())
		})
	})

	Describe("Set method", func() {
		Context("when provided valid options", func() {
			It("should accept 'option1'", func() {
				err := flag.Set("option1")
				assertT.NoError(err)
				assertT.Equal("option1", flag.String())
			})

			It("should accept 'option2'", func() {
				err := flag.Set("option2")
				assertT.NoError(err)
				assertT.Equal("option2", flag.String())
			})

			It("should accept empty string", func() {
				err := flag.Set("")
				assertT.NoError(err)
				assertT.Equal("", flag.String())
			})
		})

		Context("when provided invalid options", func() {
			It("should reject 'option4'", func() {
				err := flag.Set("option4")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be one of")
			})
		})
	})
})

var _ = Describe("RangeFlag", func() {
	var (
		flag    custom_flags.RangeFlagInterface
		assertT *assert.Assertions
	)

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
		flagVal := custom_flags.NewRangeFlag("testflag", 1, 10)
		flag = &flagVal
	})

	Describe("initialization", func() {
		It("should set correct flag name", func() {
			assertT.Equal("testflag", flag.FlagName())
		})

		It("should have string type", func() {
			assertT.Equal("string", flag.Type())
		})

		It("should set correct min and max values", func() {
			assertT.Equal(1, flag.Min())
			assertT.Equal(10, flag.Max())
		})

		It("should initialize with zero value", func() {
			assertT.Equal(0, flag.Value())
		})
	})

	Describe("NewRangeFlag panics", func() {
		It("should panic when min > max", func() {
			assertT.Panics(func() {
				custom_flags.NewRangeFlag("test", 10, 5)
			}, "Should panic when min is greater than max")
		})

		It("should panic when min is negative", func() {
			assertT.Panics(func() {
				custom_flags.NewRangeFlag("test", -1, 10)
			}, "Should panic when min is negative")
		})

		It("should panic when max is negative", func() {
			assertT.Panics(func() {
				custom_flags.NewRangeFlag("test", 0, -1)
			}, "Should panic when max is negative")
		})
	})

	Describe("Set method", func() {
		Context("when provided valid values", func() {
			It("should accept value in range", func() {
				err := flag.Set("5")
				assertT.NoError(err)
				assertT.Equal(5, flag.Value())
				assertT.Equal("5", flag.String())
			})

			It("should accept minimum value", func() {
				err := flag.Set("1")
				assertT.NoError(err)
				assertT.Equal(1, flag.Value())
			})

			It("should accept maximum value", func() {
				err := flag.Set("10")
				assertT.NoError(err)
				assertT.Equal(10, flag.Value())
			})
		})

		Context("when provided invalid values", func() {
			It("should reject value below range", func() {
				err := flag.Set("0")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be between")
			})

			It("should reject value above range", func() {
				err := flag.Set("11")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be between")
			})

			It("should reject non-numeric value", func() {
				err := flag.Set("abc")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be an integer")
			})

			It("should reject empty string", func() {
				err := flag.Set("")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be an integer")
			})

			It("should reject negative number as string", func() {
				err := flag.Set("-1")
				assertT.Error(err)
				assertT.Contains(err.Error(), "must be an integer")
			})
		})
	})
})
