package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Soldier", func() {
	var soldier string

	BeforeEach(func() {
		var err error

		soldier, err = gexec.Build("github.com/cloudfoundry-incubator/linux-circus/soldier")
		Ω(err).ShouldNot(HaveOccurred())
	})

	It("executes it with $HOME as the given dir", func() {
		session, err := gexec.Start(
			exec.Command(soldier, "/some-app-dir", "bash", "-c", "echo HOME set to $HOME"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("HOME set to /some-app-dir"))
	})

	It("executes it with $TMPDIR as the given dir + /tmp", func() {
		session, err := gexec.Start(
			exec.Command(soldier, "/some-app-dir", "bash", "-c", "echo TMPDIR set to $TMPDIR"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("TMPDIR set to /some-app-dir/tmp"))
	})

	Context("when the given dir has .profile.d with scripts in it", func() {
		var appDir string

		BeforeEach(func() {
			var err error

			appDir, err = ioutil.TempDir("", "app-dir")
			Ω(err).ShouldNot(HaveOccurred())

			profileDir := path.Join(appDir, ".profile.d")

			err = os.MkdirAll(profileDir, 0755)
			Ω(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(path.Join(profileDir, "a.sh"), []byte("echo sourcing a\nexport A=1\n"), 0644)
			Ω(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(path.Join(profileDir, "b.sh"), []byte("echo sourcing b\nexport B=1\n"), 0644)
			Ω(err).ShouldNot(HaveOccurred())
		})

		AfterEach(func() {
			os.RemoveAll(appDir)
		})

		It("sources them before executing", func() {
			session, err := gexec.Start(
				exec.Command(soldier, appDir, "bash", "-c", "env; echo running app"),
				GinkgoWriter,
				GinkgoWriter,
			)
			Ω(err).ShouldNot(HaveOccurred())

			Eventually(session).Should(gbytes.Say("sourcing a"))
			Eventually(session).Should(gbytes.Say("sourcing b"))
			Eventually(session).Should(gbytes.Say("A=1"))
			Eventually(session).Should(gbytes.Say("B=1"))
			Eventually(session).Should(gbytes.Say("running app"))
		})
	})
})
