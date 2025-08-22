package detect_test

import (
	"fmt"
	"os"
	"path/filepath" // Import filepath for joining paths in mocks
	"time"          // Added for MockFileInfo

	. "github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/detect"
	"github.com/louiss0/javascript-package-delegator/mock"
)

var _ = Describe("Detect", Label("fast", "unit"), func() {
	assert := assert.New(GinkgoT()) // Initialize assert for each spec

	Context("DetectJSPackageManager", func() {
		var mockPath *mock.MockPathLookup

		BeforeEach(func() {
			mockPath = mock.NewMockPathLookup()
		})

		type TestCase struct {
			MockLookPath map[string]struct {
				Path  string
				Error error
			}
			ExpectedManager string
		}

		DescribeTable("should detect the correct package manager based on priority",
			func(
				tc TestCase,
			) {
				// The mockPath is reset by the outer BeforeEach, ensuring a clean state for each entry.
				// Populate the mockPath with specific expectations for this test case.
				for pm, result := range tc.MockLookPath {
					mockPath.ExpectedLookPathResults[pm] = result
				}

				manager, err := detect.DetectJSPackageManager(mockPath)

				assert.NoError(err)
				assert.Equal(tc.ExpectedManager, manager, "expected manager to be %s", tc.ExpectedManager)
			},
			Entry("should detect Deno if it is available in PATH",
				TestCase{
					MockLookPath: map[string]struct {
						Path  string
						Error error
					}{
						detect.DENO: {Path: "/usr/local/bin/deno", Error: nil},
						// Other package managers could also be present, but Deno should be prioritized as it's first in the list
						detect.NPM: {Path: "/usr/bin/npm", Error: nil},
					},
					ExpectedManager: detect.DENO,
				},
			),
			Entry("should detect Bun if Deno is not available but Bun is",
				TestCase{
					MockLookPath: map[string]struct {
						Path  string
						Error error
					}{
						detect.DENO: {Path: "", Error: os.ErrNotExist},
						detect.BUN:  {Path: "/usr/local/bin/bun", Error: nil},
						// Other package managers could also be present, but Bun should be prioritized over them
						detect.YARN: {Path: "/usr/bin/yarn", Error: nil},
					},
					ExpectedManager: detect.BUN,
				},
			),
			Entry("should detect PNPM if Deno and Bun are not available but PNPM is",
				TestCase{
					MockLookPath: map[string]struct {
						Path  string
						Error error
					}{
						detect.DENO: {Path: "", Error: os.ErrNotExist},
						detect.BUN:  {Path: "", Error: os.ErrNotExist},
						detect.PNPM: {Path: "/opt/pnpm/bin/pnpm", Error: nil},
					},
					ExpectedManager: detect.PNPM,
				},
			),
			Entry("should detect Yarn if Deno, Bun, and PNPM are not available but Yarn is",
				TestCase{
					MockLookPath: map[string]struct {
						Path  string
						Error error
					}{
						detect.DENO: {Path: "", Error: os.ErrNotExist},
						detect.BUN:  {Path: "", Error: os.ErrNotExist},
						detect.PNPM: {Path: "", Error: os.ErrNotExist},
						detect.YARN: {Path: "/usr/bin/yarn", Error: nil},
					},
					ExpectedManager: detect.YARN,
				},
			),
			Entry("should detect NPM if it's the only package manager available",
				TestCase{
					MockLookPath: map[string]struct {
						Path  string
						Error error
					}{
						// Explicitly set all preceding managers as not found
						detect.DENO: {Path: "", Error: os.ErrNotExist},
						detect.BUN:  {Path: "", Error: os.ErrNotExist},
						detect.PNPM: {Path: "", Error: os.ErrNotExist},
						detect.YARN: {Path: "", Error: os.ErrNotExist},
						detect.NPM:  {Path: "/usr/bin/npm", Error: nil},
					},
					ExpectedManager: detect.NPM,
				},
			),
		)

		It("should return ErrNoPackageManager if no supported package managers are found in PATH", func() {
			// Set all supported package managers to not be found
			for _, manager := range detect.SupportedJSPackageManagers {
				mockPath.ExpectedLookPathResults[manager] = struct {
					Path  string
					Error error
				}{Path: "", Error: os.ErrNotExist}
			}

			manager, err := detect.DetectJSPackageManager(mockPath)
			assert.Error(err)
			assert.Equal(detect.ErrNoPackageManager, err)
			assert.Empty(manager) // Manager should be an empty string
		})
	})

	Context("DetectLockfile", func() {
		var mockFs *mock.MockFileSystem

		BeforeEach(func() {
			mockFs = mock.NewMockFileSystem()
			// Default GetwdFn for mockFs
			mockFs.GetwdFn = func() (string, error) {
				return "/mock/test/dir", nil
			}
		})

		It("should detect deno from deno.lock", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				if name == filepath.Join("/mock/test/dir", detect.DENO_LOCK) {
					return mock.NewMockFileInfo(detect.DENO_LOCK, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.DENO_JSON, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.DENO_JSONC, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.BUN_LOCKB, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.PNPM_LOCK_YAML, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.YARN_LOCK, 0, 0, time.Time{}, false), nil
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
					return mock.NewMockFileInfo(detect.PACKAGE_LOCK_JSON, 0, 0, time.Time{}, false), nil
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
			assert.Contains(err.Error(), "no lock file found") // Check for specific error message
		})

		It("should prioritize deno over other package managers", func() {
			mockFs.StatFn = func(name string) (os.FileInfo, error) {
				mockDir := "/mock/test/dir"
				switch name {
				case filepath.Join(mockDir, detect.DENO_JSON):
					return mock.NewMockFileInfo(detect.DENO_JSON, 0, 0, time.Time{}, false), nil
				case filepath.Join(mockDir, detect.PACKAGE_LOCK_JSON):
					return mock.NewMockFileInfo(detect.PACKAGE_LOCK_JSON, 0, 0, time.Time{}, false), nil
				case filepath.Join(mockDir, detect.YARN_LOCK):
					return mock.NewMockFileInfo(detect.YARN_LOCK, 0, 0, time.Time{}, false), nil
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

	Context("DetectJSPackageManagerBasedOnLockFile", func() {
		var mockPath *mock.MockPathLookup

		BeforeEach(func() {
			mockPath = mock.NewMockPathLookup()
			// Default: assume all package managers are found in PATH
			for _, pm := range detect.SupportedJSPackageManagers {
				mockPath.ExpectedLookPathResults[pm] = struct {
					Path  string
					Error error
				}{Path: "/mock/bin/" + pm, Error: nil}
			}
		})

		It("should return the correct package manager if found in PATH", func() {
			pm, err := detect.DetectJSPackageManagerBasedOnLockFile(detect.PACKAGE_LOCK_JSON, mockPath)
			assert.NoError(err)
			assert.Equal(detect.NPM, pm)
		})

		It("should return ErrNoPackageManager if the detected manager is not in PATH", func() {
			mockPath.ExpectedLookPathResults[detect.BUN] = struct {
				Path  string
				Error error
			}{Path: "", Error: os.ErrNotExist} // Bun is NOT found

			pm, err := detect.DetectJSPackageManagerBasedOnLockFile(detect.BUN_LOCKB, mockPath)
			assert.Error(err)
			assert.Equal(detect.ErrNoPackageManager, err)
			assert.Equal("", pm)
		})

		It("should return an error for an unsupported lockfile", func() {
			pm, err := detect.DetectJSPackageManagerBasedOnLockFile("unsupported.lock", mockPath)
			assert.Error(err)
			assert.Equal("", pm)
			assert.Contains(err.Error(), "unsupported lockfile")
		})

		It("should return other LookPath errors directly", func() {
			mockPath.ExpectedLookPathResults[detect.YARN] = struct {
				Path  string
				Error error
			}{Path: "", Error: fmt.Errorf("permission denied to access PATH")}

			pm, err := detect.DetectJSPackageManagerBasedOnLockFile(detect.YARN_LOCK, mockPath)
			assert.Error(err)
			assert.Contains(err.Error(), "permission denied")
			assert.Equal("", pm)
		})

		It("should propagate non-ErrNotExist errors for NPM as well", func() {
			// Test with a different error type to ensure all managers handle errors consistently
			mockPath.ExpectedLookPathResults[detect.NPM] = struct {
				Path  string
				Error error
			}{Path: "", Error: fmt.Errorf("network error accessing NPM registry")}

			pm, err := detect.DetectJSPackageManagerBasedOnLockFile(detect.PACKAGE_LOCK_JSON, mockPath)
			assert.Error(err)
			assert.Contains(err.Error(), "network error")
			assert.NotEqual(detect.ErrNoPackageManager, err) // Ensure it's not wrapped as ErrNoPackageManager
			assert.Equal("", pm)
		})

		type LockfileMappingCase struct {
			Lockfile   string
			ExpectedPM string
		}

		DescribeTable("maps every supported lockfile to its package manager when found in PATH",
			func(tc LockfileMappingCase) {
				pm, err := detect.DetectJSPackageManagerBasedOnLockFile(tc.Lockfile, mockPath)
				assert.NoError(err)
				assert.Equal(tc.ExpectedPM, pm)
			},
			Entry("deno.lock -> deno", LockfileMappingCase{Lockfile: detect.DENO_LOCK, ExpectedPM: detect.DENO}),
			Entry("deno.json -> deno", LockfileMappingCase{Lockfile: detect.DENO_JSON, ExpectedPM: detect.DENO}),
			Entry("deno.jsonc -> deno", LockfileMappingCase{Lockfile: detect.DENO_JSONC, ExpectedPM: detect.DENO}),
			Entry("package-lock.json -> npm", LockfileMappingCase{Lockfile: detect.PACKAGE_LOCK_JSON, ExpectedPM: detect.NPM}),
			Entry("pnpm-lock.yaml -> pnpm", LockfileMappingCase{Lockfile: detect.PNPM_LOCK_YAML, ExpectedPM: detect.PNPM}),
			Entry("bun.lockb -> bun", LockfileMappingCase{Lockfile: detect.BUN_LOCKB, ExpectedPM: detect.BUN}),
			// The following two are recently added; current implementation validates against 'lockFiles' slice,
			// so these may fail as 'unsupported lockfile' until validation is aligned.
			Entry("bun.lock.json -> bun", LockfileMappingCase{Lockfile: detect.BUN_LOCK_JSON, ExpectedPM: detect.BUN}),
			Entry("yarn.lock -> yarn", LockfileMappingCase{Lockfile: detect.YARN_LOCK, ExpectedPM: detect.YARN}),
			Entry("yarn.lock.json -> yarn", LockfileMappingCase{Lockfile: detect.YARN_LOCK_JSON, ExpectedPM: detect.YARN}),
		)

		DescribeTable("returns ErrNoPackageManager when the mapped manager is not in PATH",
			func(tc LockfileMappingCase) {
				// Simulate the mapped package manager missing from PATH
				mockPath.ExpectedLookPathResults[tc.ExpectedPM] = struct {
					Path  string
					Error error
				}{Path: "", Error: os.ErrNotExist}

				pm, err := detect.DetectJSPackageManagerBasedOnLockFile(tc.Lockfile, mockPath)
				assert.Error(err)
				assert.Equal(detect.ErrNoPackageManager, err)
				assert.Equal("", pm)
			},
			Entry("deno.lock -> deno missing", LockfileMappingCase{Lockfile: detect.DENO_LOCK, ExpectedPM: detect.DENO}),
			Entry("deno.json -> deno missing", LockfileMappingCase{Lockfile: detect.DENO_JSON, ExpectedPM: detect.DENO}),
			Entry("deno.jsonc -> deno missing", LockfileMappingCase{Lockfile: detect.DENO_JSONC, ExpectedPM: detect.DENO}),
			Entry("package-lock.json -> npm missing", LockfileMappingCase{Lockfile: detect.PACKAGE_LOCK_JSON, ExpectedPM: detect.NPM}),
			Entry("pnpm-lock.yaml -> pnpm missing", LockfileMappingCase{Lockfile: detect.PNPM_LOCK_YAML, ExpectedPM: detect.PNPM}),
			Entry("bun.lockb -> bun missing", LockfileMappingCase{Lockfile: detect.BUN_LOCKB, ExpectedPM: detect.BUN}),
			// These two may currently return "unsupported lockfile" before consulting PATH and thus fail this expectation.
			Entry("bun.lock.json -> bun missing", LockfileMappingCase{Lockfile: detect.BUN_LOCK_JSON, ExpectedPM: detect.BUN}),
			Entry("yarn.lock -> yarn missing", LockfileMappingCase{Lockfile: detect.YARN_LOCK, ExpectedPM: detect.YARN}),
			Entry("yarn.lock.json -> yarn missing", LockfileMappingCase{Lockfile: detect.YARN_LOCK_JSON, ExpectedPM: detect.YARN}),
		)
	})

	Context("DetectVolta", func() {
		var mockPath *mock.MockPathLookup

		BeforeEach(func() {
			mockPath = mock.NewMockPathLookup()
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
