package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Smelting", func() {
	buildpackFixtures := "fixtures/buildpacks"
	appFixtures := "fixtures/apps"
	alwaysDetectsHash := "4a32704add4e2bd294c35ce4ed262f62" //md5 hash of "always-detects"

	var (
		smelterCmd             *exec.Cmd
		appDir                 string
		buildpacksDir          string
		outputDir              string
		buildArtifactsCacheDir string
		resultDir              string
	)

	smelt := func() *gexec.Session {
		session, err := gexec.Start(
			smelterCmd,
			GinkgoWriter,
			GinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		return session
	}

	BeforeEach(func() {
		var err error

		appDir, err = ioutil.TempDir(os.TempDir(), "smelting-app")
		Ω(err).ShouldNot(HaveOccurred())

		buildpacksDir, err = ioutil.TempDir(os.TempDir(), "smelting-buildpacks")
		Ω(err).ShouldNot(HaveOccurred())

		outputDir, err = ioutil.TempDir(os.TempDir(), "smelting-droplet")
		Ω(err).ShouldNot(HaveOccurred())

		buildArtifactsCacheDir, err = ioutil.TempDir(os.TempDir(), "smelting-cache")
		Ω(err).ShouldNot(HaveOccurred())

		resultDir, err = ioutil.TempDir(os.TempDir(), "smelting-result")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(buildpacksDir)
		os.RemoveAll(outputDir)
	})

	Context("with a normal buildpack", func() {
		BeforeEach(func() {
			smelterCmd = exec.Command(smelterPath,
				"-appDir", appDir,
				"-buildpacksDir", buildpacksDir,
				"-outputDir", outputDir,
				"-buildArtifactsCacheDir", buildArtifactsCacheDir,
				"-buildpackOrder", "always-detects",
				"-resultDir", resultDir)
			smelterCmd.Env = os.Environ()

			cp(path.Join(buildpackFixtures, "always-detects"), path.Join(buildpacksDir, alwaysDetectsHash))
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)

			Eventually(smelt()).Should(gexec.Exit(0))
		})

		Describe("the contents of the output dir", func() {
			It("should contain an /app dir with the contents of the compilation", func() {
				appDirLocation := path.Join(outputDir, "app")
				contents, err := ioutil.ReadDir(appDirLocation)
				Ω(contents, err).Should(HaveLen(2))

				names := []string{contents[0].Name(), contents[1].Name()}
				Ω(names).Should(ContainElement("app.sh"))
				Ω(names).Should(ContainElement("compiled"))
			})

			It("should contain a droplet containing an empty /tmp directory", func() {
				tmpDirLocation := path.Join(outputDir, "tmp")
				Ω(ioutil.ReadDir(tmpDirLocation)).Should(BeEmpty())
			})

			It("should contain a droplet containing an empty /logs directory", func() {
				logsDirLocation := path.Join(outputDir, "logs")
				Ω(ioutil.ReadDir(logsDirLocation)).Should(BeEmpty())
			})

			It("should contain a staging_info.yml with the detected buildpack and start command", func() {
				stagingInfoLocation := path.Join(outputDir, "staging_info.yml")
				stagingInfo, err := ioutil.ReadFile(stagingInfoLocation)
				Ω(err).ShouldNot(HaveOccurred())

				expectedYAML := "\"detected_buildpack\": \"Always Matching\"\n\"start_command\": \"the start command\"\n"
				Ω(string(stagingInfo)).Should(Equal(expectedYAML))
			})
		})

		Describe("the result.json, which is used to communicate back to the stager", func() {
			It("exists, and contains the detected buildpack", func() {
				resultLocation := path.Join(resultDir, "result.json")
				resultInfo, err := ioutil.ReadFile(resultLocation)
				Ω(err).ShouldNot(HaveOccurred())
				expectedJSON := `{
					"detected_buildpack": "Always Matching",
					"buildpack_key": "always-detects"
				}`

				Ω(resultInfo).Should(MatchJSON(expectedJSON))
			})
		})
	})

	Context("with a nested buildpack", func() {
		BeforeEach(func() {
			nestedBuildpack := "nested-buildpack"
			nestedBuildpackHash := "70d137ae4ee01fbe39058ccdebf48460"

			nestedBuildpackDir := path.Join(buildpacksDir, nestedBuildpackHash)
			err := os.MkdirAll(nestedBuildpackDir, 0777)
			Ω(err).ShouldNot(HaveOccurred())

			smelterCmd = exec.Command(smelterPath,
				"-appDir", appDir,
				"-buildpacksDir", buildpacksDir,
				"-outputDir", outputDir,
				"-buildArtifactsCacheDir", buildArtifactsCacheDir,
				"-buildpackOrder", nestedBuildpack,
				"-resultDir", resultDir)
			smelterCmd.Env = os.Environ()

			cp(path.Join(buildpackFixtures, "always-detects"), nestedBuildpackDir)
			cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)
		})

		It("should detect the nested buildpack", func() {
			Eventually(smelt()).Should(gexec.Exit(0))
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
