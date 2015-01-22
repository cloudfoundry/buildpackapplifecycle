package main_test

import (
	"testing"

	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

var tailorPath string

func TestLinuxCircusTailor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux-Circus-Tailor Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	tailor, err := gexec.Build("github.com/cloudfoundry-incubator/linux-circus/tailor")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(tailor)
}, func(tailor []byte) {
	tailorPath = string(tailor)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
