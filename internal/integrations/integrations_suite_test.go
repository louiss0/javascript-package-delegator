package integrations

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestIntegrations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integrations Suite")
}
