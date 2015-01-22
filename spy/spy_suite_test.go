package main_test

import (
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
	"testing"
)

var spy string

func TestLinuxCircusSpy(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux-Circus-Spy Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	spyPath, err := gexec.Build("github.com/cloudfoundry-incubator/linux-circus/spy")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(spyPath)
}, func(spyPath []byte) {
	spy = string(spyPath)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
