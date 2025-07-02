package main

import (
	"fmt"
	"os"
	"path/filepath" // Import filepath for joining paths in mocks
	"time"          // Added for MockFileInfo

	"github.com/louiss0/javascript-package-delegator/detect" // Correct import path for your package
	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"
)

// MockPathLookup is a mock implementation of PathLookup for testing.
type MockPathLookup struct {
	ExpectedLookPathResults map[string]struct {
		Path  string
		Error error
	}
}

func NewMockPathLookup() *MockPathLookup {
	return &MockPathLookup{
		ExpectedLookPathResults: make(map[string]struct {
			Path  string
			Error error
		}),
	}
}

func (m *MockPathLookup) LookPath(file string) (string, error) {
	if res, ok := m.ExpectedLookPathResults[file]; ok {
		return res.Path, res.Error
	}
	return "", fmt.Errorf("mock LookPath: no expectation set for '%s'", file) // Fallback for unconfigured mocks
}

// MockFileSystem is a mock implementation of FileSystem for testing.
type MockFileSystem struct {
	StatFn  func(name string) (os.FileInfo, error)
	GetwdFn func() (string, error)
}

// NewMockFileSystem creates a new MockFileSystem with default behaviors.
func NewMockFileSystem() *MockFileSystem {
	return &MockFileSystem{
		StatFn: func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist // Default: file does not exist
		},
		GetwdFn: func() (string, error) {
			return "/mock/current/dir", nil // Default: a mock current working directory
		},
	}
}

// Stat implements FileSystem using the mock StatFn.
func (m *MockFileSystem) Stat(name string) (os.FileInfo, error) {
	return m.StatFn(name)
}

// Getwd implements FileSystem using the mock GetwdFn.
func (m *MockFileSystem) Getwd() (string, error) {
	return m.GetwdFn()
}

// MockFileInfo is a mock implementation of os.FileInfo for testing.
type MockFileInfo struct {
	NameVal    string
	SizeVal    int64
	ModeVal    os.FileMode
	ModTimeVal time.Time
	IsDirVal   bool
	SysVal     interface{}
}

func (m *MockFileInfo) Name() string       { return m.NameVal }
func (m *MockFileInfo) Size() int64        { return m.SizeVal }
func (m *MockFileInfo) Mode() os.FileMode  { return m.ModeVal }
func (m *MockFileInfo) ModTime() time.Time { return m.ModTimeVal }
func (m *MockFileInfo) IsDir() bool        { return m.IsDirVal }
func (m *MockFileInfo) Sys() interface{}   { return m.SysVal }

var _ = Describe("Detect", func() {
	assert := assert.New(GinkgoT()) // Initialize assert for each spec

	Context("DetectLockfile", func() {
		var mockFs *MockFileSystem

		BeforeEach(func() {
			mockFs = NewMockFileSystem()
			// Default GetwdFn for mockFs
			mockFs.GetwdFn = func() (string, error) {
				return "/mock/test/dir", nil
			}
		})

		It("should detect deno from deno.lock", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.DENO_LOCK) {
					return &MockFileInfo{NameVal: detect.DENO_LOCK, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.DENO_LOCK, lockfile)
		})

		It("should detect deno from deno.json", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.DENO_JSON) {
					return &MockFileInfo{NameVal: detect.DENO_JSON, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.DENO_JSON, lockfile)
		})

		It("should detect deno from deno.jsonc", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.DENO_JSONC) {
					return &MockFileInfo{NameVal: detect.DENO_JSONC, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.DENO_JSONC, lockfile)
		})

		It("should detect bun from bun.lockb", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.BUN_LOCKB) {
					return &MockFileInfo{NameVal: detect.BUN_LOCKB, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.BUN_LOCKB, lockfile)
		})

		It("should detect pnpm from pnpm-lock.yaml", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.PNPM_LOCK_YAML) {
					return &MockFileInfo{NameVal: detect.PNPM_LOCK_YAML, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.PNPM_LOCK_YAML, lockfile)
		})

		It("should detect yarn from yarn.lock", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.YARN_LOCK) {
					return &MockFileInfo{NameVal: detect.YARN_LOCK, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.YARN_LOCK, lockfile)
		})

		It("should detect npm from package-lock.json", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.PACKAGE_LOCK_JSON) {
					return &MockFileInfo{NameVal: detect.PACKAGE_LOCK_JSON, IsDirVal: false}, nil
				}
				return nil, os.ErrNotExist
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.PACKAGE_LOCK_JSON, lockfile)
		})

		It("should return an error when no lock files found", func() {
			// Default mockFs.StatFn (returns os.ErrNotExist) covers this
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.Error(err)
			assert.Equal("", lockfile)
			assert.Contains(err.Error(), "No lock file found") // Check for specific error message
		})

		It("should prioritize deno over other package managers", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				mockDir := "/mock/test/dir"
				switch name {
				case filepath.Join(mockDir, detect.DENO_JSON):
					return &MockFileInfo{NameVal: detect.DENO_JSON, IsDirVal: false}, nil
				case filepath.Join(mockDir, detect.PACKAGE_LOCK_JSON):
					return &MockFileInfo{NameVal: detect.PACKAGE_LOCK_JSON, IsDirVal: false}, nil
				case filepath.Join(mockDir, detect.YARN_LOCK):
					return &MockFileInfo{NameVal: detect.YARN_LOCK, IsDirVal: false}, nil
				default:
					return nil, os.ErrNotExist
				}
			}
			lockfile, err := detect.DetectLockfile(mockFs)
			assert.NoError(err)
			assert.Equal(detect.DENO_JSON, lockfile) // Deno should be prioritized
		})

		It("should return an error if Getwd fails", func() {
			mockFs.GetwdFn = func() (string, error) {
				return "", fmt.Errorf("permission denied to get working directory")
			}
			_, err := detect.DetectLockfile(mockFs)
			assert.Error(err)
			assert.Contains(err.Error(), "permission denied")
		})

	})

	Context("DetectJSPacakgeManagerBasedOnLockFile", func() {
		var mockPath *MockPathLookup

		BeforeEach(func() {
			mockPath = NewMockPathLookup()
			// Default: assume all package managers are found in PATH
			for _, pm := range detect.SupportedJSPackageManagers {
				mockPath.ExpectedLookPathResults[pm] = struct {
					Path  string
					Error error
				}{Path: "/mock/bin/" + pm, Error: nil}
			}
		})

		It("should return the correct package manager if found in PATH", func() {
			pm, err := detect.DetectJSPacakgeManagerBasedOnLockFile(detect.PACKAGE_LOCK_JSON, mockPath)
			assert.NoError(err)
			assert.Equal(detect.NPM, pm)
		})

		It("should return ErrNoPackageManager if the detected manager is not in PATH", func() {
			mockPath.ExpectedLookPathResults[detect.BUN] = struct {
				Path  string
				Error error
			}{Path: "", Error: os.ErrNotExist} // Bun is NOT found

			pm, err := detect.DetectJSPacakgeManagerBasedOnLockFile(detect.BUN_LOCKB, mockPath)
			assert.Error(err)
			assert.Equal(detect.ErrNoPackageManager, err)
			assert.Equal("", pm)
		})

		It("should return an error for an unsupported lockfile", func() {
			pm, err := detect.DetectJSPacakgeManagerBasedOnLockFile("unsupported.lock", mockPath)
			assert.Error(err)
			assert.Equal("", pm)
			assert.Contains(err.Error(), "unsupported lockfile")
		})

		It("should return other LookPath errors directly", func() {
			mockPath.ExpectedLookPathResults[detect.YARN] = struct {
				Path  string
				Error error
			}{Path: "", Error: fmt.Errorf("permission denied to access PATH")}

			pm, err := detect.DetectJSPacakgeManagerBasedOnLockFile(detect.YARN_LOCK, mockPath)
			assert.Error(err)
			assert.Contains(err.Error(), "permission denied")
			assert.Equal("", pm)
		})
	})

	Context("DetectVolta", func() {
		var mockPath *MockPathLookup

		BeforeEach(func() {
			mockPath = NewMockPathLookup()
		})

		It("should return true if Volta is found in PATH", func() {
			mockPath.ExpectedLookPathResults[detect.VOLTA] = struct {
				Path  string
				Error error
			}{Path: "/usr/local/bin/volta", Error: nil}

			found := detect.DetectVolta(mockPath)
			assert.True(found)
		})

		It("should return false if Volta is not found in PATH", func() {
			mockPath.ExpectedLookPathResults[detect.VOLTA] = struct {
				Path  string
				Error error
			}{Path: "", Error: os.ErrNotExist}

			found := detect.DetectVolta(mockPath)
			assert.False(found)
		})

		It("should return false if LookPath for Volta returns any other error", func() {
			mockPath.ExpectedLookPathResults[detect.VOLTA] = struct {
				Path  string
				Error error
			}{Path: "", Error: fmt.Errorf("some other error")}

			found := detect.DetectVolta(mockPath)
			assert.False(found)
		})
	})
})
