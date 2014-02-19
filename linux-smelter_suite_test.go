package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/cmdtest"
)

var smelterPath string

func TestLinuxSmelter(t *testing.T) {
	var err error

	smelterPath, err = cmdtest.Build("github.com/cloudfoundry-incubator/linux-smelter")
	Ω(err).ShouldNot(HaveOccurred())

	RegisterFailHandler(Fail)
	RunSpecs(t, "Linux-Smelter Suite")
}
