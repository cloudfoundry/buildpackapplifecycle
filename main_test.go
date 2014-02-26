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
			runner_support.TeeIfVerbose,
			runner_support.TeeIfVerbose,
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
		cp(path.Join(appFixtures, "bash-app")+"/", appDir)

		Ω(smelt()).Should(ExitWith(0))
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(buildpacksDir)
		os.RemoveAll(outputDir)
	})

	It("should produce an /app dir with the contents of the compilation", func() {
		appDirLocation := path.Join(outputDir, "app")
		contents, err := ioutil.ReadDir(appDirLocation)
		Ω(contents, err).Should(HaveLen(2))

		names := []string{contents[0].Name(), contents[1].Name()}
		Ω(names).Should(ContainElement("app.sh"))
		Ω(names).Should(ContainElement("compiled"))
	})

	It("should produce a droplet containing an empty /tmp directory", func() {
		tmpDirLocation := path.Join(outputDir, "tmp")
		Ω(ioutil.ReadDir(tmpDirLocation)).Should(BeEmpty())
	})

	It("should produce a droplet containing an empty /logs directory", func() {
		logsDirLocation := path.Join(outputDir, "logs")
		Ω(ioutil.ReadDir(logsDirLocation)).Should(BeEmpty())
	})

	It("should produce a staging_info.yml with the correct information", func() {
		stagingInfoLocation := path.Join(outputDir, "staging_info.yml")
		stagingInfo, err := ioutil.ReadFile(stagingInfoLocation)
		Ω(err).ShouldNot(HaveOccurred())

		expectedYAML := "\"detected_buildpack\": \"Always Matching\"\n\"start_command\": \"the start command\"\n"
		Ω(string(stagingInfo)).Should(Equal(expectedYAML))
	})

	It("produces a result.json", func() {
		resultLocation := path.Join(resultDir, "result.json")
		resultInfo, err := ioutil.ReadFile(resultLocation)
		Ω(err).ShouldNot(HaveOccurred())
		expectedJSON := `{
			"detected_buildpack": "Always Matching"
		}`

		Ω(resultInfo).Should(MatchJSON(expectedJSON))
	})
})

func cp(src string, dst string) {
	session, err := cmdtest.StartWrapped(
		exec.Command("cp", "-a", src, dst),
		runner_support.TeeIfVerbose,
		runner_support.TeeIfVerbose,
	)
	Ω(err).ShouldNot(HaveOccurred())
	Ω(session).Should(ExitWith(0))
}
