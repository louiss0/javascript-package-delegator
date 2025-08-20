package services_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
)

func TestServices(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Services Suite")
}
