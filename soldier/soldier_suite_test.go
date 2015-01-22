package main_test

import (
	"testing"

	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/ginkgo"
	. "github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega"
	"github.com/cloudfoundry-incubator/linux-circus/Godeps/_workspace/src/github.com/onsi/gomega/gexec"
)

var soldier string

func TestLinuxCircusSoldier(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux-Circus-Soldier Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	soldierPath, err := gexec.Build("github.com/cloudfoundry-incubator/linux-circus/soldier")
	Ω(err).ShouldNot(HaveOccurred())
	return []byte(soldierPath)
}, func(soldierPath []byte) {
	soldier = string(soldierPath)
})

var _ = SynchronizedAfterSuite(func() {
	//noop
}, func() {
	gexec.CleanupBuildArtifacts()
})
