package shell_alias

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestShellAlias(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Shell Alias Suite")
}
