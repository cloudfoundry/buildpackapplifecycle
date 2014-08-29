package main_test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Spy", func() {
	var (
		server     *ghttp.Server
		serverAddr string
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		serverAddr = server.HTTPTestServer.Listener.Addr().String()
	})

	runSpy := func() *gexec.Session {
		session, err := gexec.Start(exec.Command(spy, "-addr", serverAddr), GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
		return session
	}

	Context("when the address is listening", func() {
		It("exits 0", func() {
			session := runSpy()
			Eventually(session).Should(gexec.Exit(0))
		})

		It("logs that the healthcheck passed", func() {
			session := runSpy()
			Eventually(session.Out).Should(gbytes.Say("healthcheck passed"))
		})
	})

	Context("when the address is not listening", func() {
		BeforeEach(func() {
			server.Close()
		})

		It("exits 1", func() {
			session := runSpy()
			Eventually(session).Should(gexec.Exit(1))
		})

		It("logs that the healthcheck failed", func() {
			session := runSpy()
			Eventually(session.Out).Should(gbytes.Say("healthcheck failed"))
		})
	})
})
