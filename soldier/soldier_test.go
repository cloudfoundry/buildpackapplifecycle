package main_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Soldier", func() {
	var appDir string
	var soldierCmd *exec.Cmd
	var session *gexec.Session

	BeforeEach(func() {
		os.Setenv("CALLERENV", "some-value")

		var err error
		appDir, err = ioutil.TempDir("", "app-dir")
		Ω(err).ShouldNot(HaveOccurred())

		soldierCmd = &exec.Cmd{
			Path: soldier,
			Env: append(
				os.Environ(),
				"PORT=8080",
				"CF_INSTANCE_GUID=some-instance-guid",
				"CF_INSTANCE_INDEX=123",
				`VCAP_APPLICATION={"foo":1}`,
			),
		}
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
	})

	JustBeforeEach(func() {
		var err error
		session, err = gexec.Start(soldierCmd, GinkgoWriter, GinkgoWriter)
		Ω(err).ShouldNot(HaveOccurred())
	})

	var ItExecutesTheCommandWithTheRightEnvironment = func() {
		It("executes the start command with $HOME as the given dir", func() {
			Eventually(session).Should(gbytes.Say("HOME=" + appDir))
		})

		It("executes the start command with $TMPDIR as the given dir + /tmp", func() {
			Eventually(session).Should(gbytes.Say("TMPDIR=" + appDir + "/tmp"))
		})

		It("executes with the environment of the caller", func() {
			Eventually(session).Should(gbytes.Say("CALLERENV=some-value"))
		})

		It("changes to the app directory when running", func() {
			Eventually(session).Should(gbytes.Say("PWD=" + appDir))
		})

		It("munges VCAP_APPLICATION appropriately", func() {
			Eventually(session).Should(gexec.Exit(0))

			vcapAppPattern := regexp.MustCompile("VCAP_APPLICATION=(.*)")
			vcapApplicationBytes := vcapAppPattern.FindSubmatch(session.Out.Contents())[1]

			vcapApplication := map[string]interface{}{}
			err := json.Unmarshal(vcapApplicationBytes, &vcapApplication)
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
				Eventually(session).Should(gbytes.Say("sourcing a"))
				Eventually(session).Should(gbytes.Say("sourcing b"))
				Eventually(session).Should(gbytes.Say("A=1"))
				Eventually(session).Should(gbytes.Say("B=1"))
				Eventually(session).Should(gbytes.Say("running app"))
			})
		})
	}

	Context("when a start command is given", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
				appDir,
				"env; echo running app",
				`{ "start_command": "echo should not run this" }`,
			}
		})

		ItExecutesTheCommandWithTheRightEnvironment()
	})

	Context("when no start command is given", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
				appDir,
				"",
				`{ "start_command": "env; echo running app" }`,
			}
		})

		ItExecutesTheCommandWithTheRightEnvironment()
	})

	ItPrintsUsageInformation := func() {
		It("prints usage information", func() {
			Eventually(session.Err).Should(gbytes.Say("Usage: soldier <app directory> <start command> <metadata>"))
			Eventually(session).Should(gexec.Exit(1))
		})
	}

	Context("when no arguments are given", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
			}
		})

		ItPrintsUsageInformation()
	})

	Context("when the start command and metadata are missing", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
				appDir,
			}
		})

		ItPrintsUsageInformation()
	})

	Context("when the metadata is missing", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
				appDir,
				"env",
			}
		})

		ItPrintsUsageInformation()
	})

	Context("when the given execution metadata is not valid JSON", func() {
		BeforeEach(func() {
			soldierCmd.Args = []string{
				"soldier",
				appDir,
				"",
				"{ not-valid-json }",
			}
		})

		It("prints an error message", func() {
			Eventually(session.Err).Should(gbytes.Say("Invalid metadata"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})
})
