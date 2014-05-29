package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Spy", func() {
	var check string

	var server *ghttp.Server
	var serverAddr string

	BeforeEach(func() {
		var err error

		check, err = gexec.Build("github.com/cloudfoundry-incubator/linux-circus/spy")
		Ω(err).ShouldNot(HaveOccurred())

		server = ghttp.NewServer()

		serverAddr = server.HTTPTestServer.Listener.Addr().String()
	})

	Context("when the address is listening", func() {
		It("exits 0", func() {
			session, err := gexec.Start(
				exec.Command(check, "-addr", serverAddr),
				GinkgoWriter,
				GinkgoWriter,
			)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(0))
		})
	})

	Context("when the address is not listening", func() {
		BeforeEach(func() {
			server.Close()
		})

		It("exits 1", func() {
			session, err := gexec.Start(
				exec.Command(check, "-addr", serverAddr),
				GinkgoWriter,
				GinkgoWriter,
			)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
