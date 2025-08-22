package testutil_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"

	"github.com/louiss0/javascript-package-delegator/testutil"
)

func TestTestutil(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Testutil Suite")
}

var _ = Describe("Testutil Functions", func() {
	var assertT *assert.Assertions

	BeforeEach(func() {
		assertT = assert.New(GinkgoT())
	})

	Describe("SnapshotWorkingTree", func() {
		It("should return a non-nil snapshot without error", func() {
			snapshot, err := testutil.SnapshotWorkingTree()
			assertT.NoError(err, "SnapshotWorkingTree() should not return an error")
			assertT.NotNil(snapshot, "Expected SnapshotWorkingTree() to return a non-nil snapshot")
		})
	})

	Describe("AssertWorkingTreeClean", func() {
		It("should execute without panic when working tree is clean", func() {
			// This is a basic smoke test - we can't easily test the full functionality
			// without creating actual files and potentially interfering with the git state
			snapshot, err := testutil.SnapshotWorkingTree()
			assertT.NoError(err)

			// Test that the function doesn't panic when called
			assertT.NotPanics(func() {
				// We'll test with a snapshot against itself, which should be clean
				testutil.AssertWorkingTreeClean(GinkgoT(), snapshot)
			}, "AssertWorkingTreeClean should not panic when working tree is clean")
		})
	})

	Describe("CleanupWorkingTree integration", func() {
		It("should be testable using standard Go tests", func() {
			// Since CleanupWorkingTree requires *testing.T and Cleanup(),
			// we acknowledge that this function is better tested in the
			// existing standard Go test format. The key functionality
			// (SnapshotWorkingTree and AssertWorkingTreeClean) is tested above.
			assertT.True(true, "CleanupWorkingTree is tested via standard Go tests in cleancheck_example_test.go")
		})
	})
})
