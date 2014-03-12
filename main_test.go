package main_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/cloudfoundry/gunk/runner_support"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/vito/cmdtest"
	. "github.com/vito/cmdtest/matchers"
)

var _ = Describe("Smelting", func() {
	buildpackFixtures := "fixtures/buildpacks"
	appFixtures := "fixtures/apps"

	var (
		smelterCmd    *exec.Cmd
		appDir        string
		buildpacksDir string
		outputDir     string
		cacheDir      string
		resultDir     string
	)

	smelt := func() *cmdtest.Session {
		session, err := cmdtest.StartWrapped(
			smelterCmd,
			runner_support.TeeToGinkgoWriter,
			runner_support.TeeToGinkgoWriter,
		)
		Ω(err).ShouldNot(HaveOccurred())

		return session
	}

	BeforeEach(func() {
		smelterCmd = exec.Command(smelterPath)

		var err error

		appDir, err = ioutil.TempDir(os.TempDir(), "smelting-app")
		Ω(err).ShouldNot(HaveOccurred())

		buildpacksDir, err = ioutil.TempDir(os.TempDir(), "smelting-buildpacks")
		Ω(err).ShouldNot(HaveOccurred())

		outputDir, err = ioutil.TempDir(os.TempDir(), "smelting-droplet")
		Ω(err).ShouldNot(HaveOccurred())

		cacheDir, err = ioutil.TempDir(os.TempDir(), "smelting-cache")
		Ω(err).ShouldNot(HaveOccurred())

		resultDir, err = ioutil.TempDir(os.TempDir(), "smelting-result")
		Ω(err).ShouldNot(HaveOccurred())

		smelterCmd.Env = append(
			os.Environ(),
			"APP_DIR="+appDir,
			"BUILDPACKS_DIR="+buildpacksDir,
			"OUTPUT_DIR="+outputDir,
			"CACHE_DIR="+cacheDir,
			"BUILDPACK_ORDER=always-detects",
			"RESULT_DIR="+resultDir,
		)

		cp(path.Join(buildpackFixtures, "always-detects"), buildpacksDir)
		cp(path.Join(appFixtures, "bash-app", "app.sh"), appDir)

		Ω(smelt()).Should(ExitWith(0))
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(buildpacksDir)
		os.RemoveAll(outputDir)
	})

	Describe("the contents of the output dir", func() {
		var tarBallLocation string

		BeforeEach(func() {
			tarBallLocation = path.Join(outputDir, "droplet.tgz")
		})

		It("should produce a gzipped droplet tarball", func() {
			_, err := os.Stat(tarBallLocation)
			Ω(err).ShouldNot(HaveOccurred())

			assertIsGzip(tarBallLocation)
		})

		Describe("the contents of the droplet tarball", func() {
			var dropletUntarDir string

			BeforeEach(func() {
				var err error
				dropletUntarDir, err = ioutil.TempDir(os.TempDir(), "droplet-untar")
				Ω(err).ShouldNot(HaveOccurred())

				untar(tarBallLocation, dropletUntarDir)
			})

			AfterEach(func() {
				os.RemoveAll(dropletUntarDir)
			})

			It("should contain an /app dir with the contents of the compilation", func() {
				appDirLocation := path.Join(dropletUntarDir, "app")
				contents, err := ioutil.ReadDir(appDirLocation)
				Ω(contents, err).Should(HaveLen(2))

				names := []string{contents[0].Name(), contents[1].Name()}
				Ω(names).Should(ContainElement("app.sh"))
				Ω(names).Should(ContainElement("compiled"))
			})

			It("should contain a droplet containing an empty /tmp directory", func() {
				tmpDirLocation := path.Join(dropletUntarDir, "tmp")
				Ω(ioutil.ReadDir(tmpDirLocation)).Should(BeEmpty())
			})

			It("should contain a droplet containing an empty /logs directory", func() {
				logsDirLocation := path.Join(dropletUntarDir, "logs")
				Ω(ioutil.ReadDir(logsDirLocation)).Should(BeEmpty())
			})

			It("should contain a staging_info.yml with the detected buildpack and start command", func() {
				stagingInfoLocation := path.Join(dropletUntarDir, "staging_info.yml")
				stagingInfo, err := ioutil.ReadFile(stagingInfoLocation)
				Ω(err).ShouldNot(HaveOccurred())

				expectedYAML := "\"detected_buildpack\": \"Always Matching\"\n\"start_command\": \"the start command\"\n"
				Ω(string(stagingInfo)).Should(Equal(expectedYAML))
			})
		})
	})

	Describe("the result.json, which is used to communicate back to the stager", func() {
		It("exists, and contains the detected buildpack", func() {
			resultLocation := path.Join(resultDir, "result.json")
			resultInfo, err := ioutil.ReadFile(resultLocation)
			Ω(err).ShouldNot(HaveOccurred())
			expectedJSON := `{
			"detected_buildpack": "Always Matching"
		}`

			Ω(resultInfo).Should(MatchJSON(expectedJSON))
		})
	})
})

func cp(src string, dst string) {
	session, err := cmdtest.StartWrapped(
		exec.Command("cp", "-a", src, dst),
		runner_support.TeeToGinkgoWriter,
		runner_support.TeeToGinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(session).Should(ExitWith(0))
}

func assertIsGzip(src string) {
	session, err := cmdtest.StartWrapped(
		exec.Command("gunzip", src),
		runner_support.TeeToGinkgoWriter,
		runner_support.TeeToGinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(session).Should(ExitWith(0), "Expected a gzipped file, got something else")
}

func untar(src string, dst string) {
	session, err := cmdtest.StartWrapped(
		exec.Command("tar", "-xvf", src, "-C", dst),
		runner_support.TeeToGinkgoWriter,
		runner_support.TeeToGinkgoWriter,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(session).Should(ExitWith(0))
}
