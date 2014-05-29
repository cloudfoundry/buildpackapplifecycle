package main_test

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Tailoring", func() {
	buildpackFixtures := "fixtures/buildpacks"
	appFixtures := "fixtures/apps"

	var (
		tailorCmd              *exec.Cmd
		appDir                 string
		buildpacksDir          string
		outputDropletDir       string
		buildArtifactsCacheDir string
		outputMetadataDir      string
		buildpacks             []string
	)

	tailor := func() *gexec.Session {
		session, err := gexec.Start(
			tailorCmd,
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		return session
	}

	cpBuildpack := func(buildpack string) {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(buildpack)))
		cp(path.Join(buildpackFixtures, buildpack), path.Join(buildpacksDir, hash))
	}

	BeforeEach(func() {
		var err error

		appDir, err = ioutil.TempDir(os.TempDir(), "tailoring-app")
		Ω(err).ShouldNot(HaveOccurred())

		buildpacksDir, err = ioutil.TempDir(os.TempDir(), "tailoring-buildpacks")
		Ω(err).ShouldNot(HaveOccurred())

		outputDropletDir, err = ioutil.TempDir(os.TempDir(), "tailoring-droplet")
		Ω(err).ShouldNot(HaveOccurred())

		buildArtifactsCacheDir, err = ioutil.TempDir(os.TempDir(), "tailoring-cache")
		Ω(err).ShouldNot(HaveOccurred())

		outputMetadataDir, err = ioutil.TempDir(os.TempDir(), "tailoring-metadata")
		Ω(err).ShouldNot(HaveOccurred())

		buildpacks = nil
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(buildpacksDir)
		os.RemoveAll(outputDropletDir)
	})

	JustBeforeEach(func() {
		flags := []string{
			"-appDir", appDir,
			"-buildpacksDir", buildpacksDir,
			"-outputDropletDir", outputDropletDir,
			"-buildArtifactsCacheDir", buildArtifactsCacheDir,
			"-outputMetadataDir", outputMetadataDir,
		}

		for _, name := range buildpacks {
			flags = append(flags, "-buildpack", name)
		}

		tailorCmd = exec.Command(tailorPath, flags...)

		tailorCmd.Env = os.Environ()
	})

	Context("with a normal buildpack", func() {
		BeforeEach(func() {
			buildpacks = []string{"always-detects", "also-always-detects"}

			cpBuildpack("always-detects")
			cpBuildpack("also-always-detects")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		JustBeforeEach(func() {
			Eventually(tailor()).Should(gexec.Exit(0))
		})

		Describe("the contents of the output dir", func() {
			It("should contain an /app dir with the contents of the compilation", func() {
				appDirLocation := path.Join(outputDropletDir, "app")
				contents, err := ioutil.ReadDir(appDirLocation)
				Ω(contents, err).Should(HaveLen(2))

				names := []string{contents[0].Name(), contents[1].Name()}
				Ω(names).Should(ContainElement("app.sh"))
				Ω(names).Should(ContainElement("compiled"))
			})

			It("should contain a droplet containing an empty /tmp directory", func() {
				tmpDirLocation := path.Join(outputDropletDir, "tmp")
				Ω(ioutil.ReadDir(tmpDirLocation)).Should(BeEmpty())
			})

			It("should contain a droplet containing an empty /logs directory", func() {
				logsDirLocation := path.Join(outputDropletDir, "logs")
				Ω(ioutil.ReadDir(logsDirLocation)).Should(BeEmpty())
			})

			It("should stop after detecting, and contain a staging_info.yml with the detected buildpack", func() {
				stagingInfoLocation := path.Join(outputDropletDir, "staging_info.yml")
				stagingInfo, err := ioutil.ReadFile(stagingInfoLocation)
				Ω(err).ShouldNot(HaveOccurred())

				expectedYAML := `detected_buildpack: Always Matching
start_command: the start command
`
				Ω(string(stagingInfo)).Should(Equal(expectedYAML))
			})
		})

		Describe("the result.json, which is used to communicate back to the stager", func() {
			It("exists, and contains the detected buildpack", func() {
				resultLocation := path.Join(outputMetadataDir, "result.json")
				resultInfo, err := ioutil.ReadFile(resultLocation)
				Ω(err).ShouldNot(HaveOccurred())
				expectedJSON := `{
          "detected_buildpack": "Always Matching",
          "detected_start_command": "the start command",
          "buildpack_key": "always-detects"
        }`

				Ω(resultInfo).Should(MatchJSON(expectedJSON))
			})
		})
	})

	Context("when no buildpacks match", func() {
		BeforeEach(func() {
			buildpacks = []string{"always-fails"}

			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
			cpBuildpack("always-fails")
		})

		It("should exit with an error", func() {
			session := tailor()
			Eventually(session.Err).Should(gbytes.Say("no valid buildpacks detected"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack fails in compile", func() {
		BeforeEach(func() {
			buildpacks = []string{"fails-to-compile"}

			cpBuildpack("fails-to-compile")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		It("should exit with an error", func() {
			session := tailor()
			Eventually(session.Err).Should(gbytes.Say("failed to compile droplet: exit status 1"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack release generates invalid yaml", func() {
		BeforeEach(func() {
			buildpacks = []string{"release-generates-bad-yaml"}

			cpBuildpack("release-generates-bad-yaml")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		It("should exit with an error", func() {
			session := tailor()
			Eventually(session.Err).Should(gbytes.Say("buildpack's release output invalid"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("when the buildpack fails to release", func() {
		BeforeEach(func() {
			buildpacks = []string{"fails-to-release"}

			cpBuildpack("fails-to-release")
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		It("should exit with an error", func() {
			session := tailor()
			Eventually(session.Err).Should(gbytes.Say("failed to build droplet release: exit status 1"))
			Eventually(session).Should(gexec.Exit(1))
		})
	})

	Context("with a nested buildpack", func() {
		BeforeEach(func() {
			nestedBuildpack := "nested-buildpack"
			buildpacks = []string{nestedBuildpack}

			nestedBuildpackHash := "70d137ae4ee01fbe39058ccdebf48460"

			nestedBuildpackDir := path.Join(buildpacksDir, nestedBuildpackHash)
			err := os.MkdirAll(nestedBuildpackDir, 0777)
			Ω(err).ShouldNot(HaveOccurred())

			cp(path.Join(buildpackFixtures, "always-detects"), nestedBuildpackDir)
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		It("should detect the nested buildpack", func() {
			Eventually(tailor()).Should(gexec.Exit(0))
		})
	})
})

func cp(src string, dst string) {
	session, err := gexec.Start(
		exec.Command("cp", "-a", src, dst),
		GinkgoWriter,
		GinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
}
