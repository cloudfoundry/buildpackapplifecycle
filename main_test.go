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
	buildpackFixtures := "./fixtures/buildpacks"
	appFixtures := "./fixtures/apps"

	var smelterCmd *exec.Cmd

	var (
		appDir         string
		buildpacksDir  string
		outputDir      string
		cacheDir       string
		buildpackOrder string
		resultDir      string
	)

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
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(buildpacksDir)
		os.RemoveAll(outputDir)
	})

	JustBeforeEach(func() {
		smelterCmd.Env = append(
			os.Environ(),
			"APP_DIR="+appDir,
			"BUILDPACKS_DIR="+buildpacksDir,
			"OUTPUT_DIR="+outputDir,
			"CACHE_DIR="+cacheDir,
			"BUILDPACK_ORDER="+buildpackOrder,
			"RESULT_DIR="+resultDir,
		)
	})

	smelt := func() *cmdtest.Session {
		session, err := cmdtest.StartWrapped(
			smelterCmd,
			runner_support.TeeIfVerbose,
			runner_support.TeeIfVerbose,
		)
		Ω(err).ShouldNot(HaveOccurred())

		return session
	}

	cp := func(src string, dst string) {
		session, err := cmdtest.StartWrapped(
			exec.Command("cp", "-a", src, dst),
			runner_support.TeeIfVerbose,
			runner_support.TeeIfVerbose,
		)
		Ω(err).ShouldNot(HaveOccurred())
		Ω(session).Should(ExitWith(0))
	}

	Context("when a buildpack succeeds", func() {
		BeforeEach(func() {
			cp(path.Join(buildpackFixtures, "always-detects"), buildpacksDir)
			cp(path.Join(appFixtures, "bash-app/"), appDir)

			buildpackOrder = "always-detects"
		})

		It("produces a droplet", func() {
			Ω(smelt()).Should(ExitWith(0))

			fileInfo, err := os.Stat(path.Join(outputDir, "app"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeTrue())

			fileInfo, err = os.Stat(path.Join(outputDir, "tmp"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeTrue())

			fileInfo, err = os.Stat(path.Join(outputDir, "logs"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeTrue())

			fileInfo, err = os.Stat(path.Join(outputDir, "staging_info.yml"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeFalse())
		})
	})
})
