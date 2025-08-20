package detect_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestDetect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Detect Suite")
}
