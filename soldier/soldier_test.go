package main_test

import (
	"bytes"
	"encoding/json"
	"io"
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
	var appDir string

	BeforeEach(func() {
		var err error

		appDir, err = ioutil.TempDir("", "app-dir")
		Ω(err).ShouldNot(HaveOccurred())

		soldier, err = gexec.Build("github.com/cloudfoundry-incubator/linux-circus/soldier")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
	})

	It("executes it with $HOME as the given dir", func() {
		session, err := gexec.Start(
			exec.Command(soldier, appDir, "echo HOME set to $HOME"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("HOME set to " + appDir))
	})

	It("executes it with $TMPDIR as the given dir + /tmp", func() {
		session, err := gexec.Start(
			exec.Command(soldier, appDir, "echo TMPDIR set to $TMPDIR"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("TMPDIR set to " + appDir + "/tmp"))
	})

	It("executes with the environment of the caller", func() {
		os.Setenv("CALLERENV", "some-value")

		session, err := gexec.Start(
			exec.Command(soldier, appDir, "echo CALLERENV set to $CALLERENV"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("CALLERENV set to some-value"))
	})

	It("changes to the app directory when running", func() {
		session, err := gexec.Start(
			exec.Command(soldier, appDir, "echo PWD is $(pwd)"),
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gbytes.Say("PWD is " + appDir))
	})

	It("munges VCAP_APPLICATION appropriately", func() {
		outBuf := new(bytes.Buffer)

		cmd := exec.Command(soldier, appDir, "echo $VCAP_APPLICATION")
		cmd.Env = append(
			os.Environ(),
			"PORT=8080",
			"CF_INSTANCE_GUID=some-instance-guid",
			"CF_INSTANCE_INDEX=123",
			`VCAP_APPLICATION={"foo":1}`,
		)

		session, err := gexec.Start(
			cmd,
			io.MultiWriter(GinkgoWriter, outBuf),
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		Eventually(session).Should(gexec.Exit(0))

		vcapApplication := map[string]interface{}{}
		err = json.Unmarshal(outBuf.Bytes(), &vcapApplication)
		Ω(err).ShouldNot(HaveOccurred())

		Ω(vcapApplication["host"]).Should(Equal("0.0.0.0"))
		Ω(vcapApplication["port"]).Should(Equal(float64(8080)))
		Ω(vcapApplication["instance_index"]).Should(Equal(float64(123)))
		Ω(vcapApplication["instance_id"]).Should(Equal("some-instance-guid"))
		Ω(vcapApplication["foo"]).Should(Equal(float64(1)))
	})

	Context("when the given dir has .profile.d with scripts in it", func() {
		BeforeEach(func() {
			var err error

			profileDir := path.Join(appDir, ".profile.d")

			err = os.MkdirAll(profileDir, 0755)
			Ω(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(path.Join(profileDir, "a.sh"), []byte("echo sourcing a\nexport A=1\n"), 0644)
			Ω(err).ShouldNot(HaveOccurred())

			err = ioutil.WriteFile(path.Join(profileDir, "b.sh"), []byte("echo sourcing b\nexport B=1\n"), 0644)
			Ω(err).ShouldNot(HaveOccurred())
		})

		It("sources them before executing", func() {
			session, err := gexec.Start(
				exec.Command(soldier, appDir, "env; echo running app"),
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
