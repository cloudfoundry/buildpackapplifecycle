package linux_circus_test

import (
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega"
	"testing"
)

func TestLinuxCircus(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LinuxCircus Suite")
}
