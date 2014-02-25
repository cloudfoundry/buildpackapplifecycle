package candiedyaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCandiedyaml(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Candiedyaml Suite")
}
