package droplet_test

import (
	"errors"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	. "github.com/cloudfoundry-incubator/linux-smelter/droplet"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileSystem", func() {
	var (
		fs        *FileSystem
		runner    *fake_command_runner.FakeCommandRunner
		appDir    string
		outputDir string
	)

	BeforeEach(func() {
		runner = fake_command_runner.New()

		var err error

		appDir = "/path/to/app/dir"

		outputDir, err = ioutil.TempDir(os.TempDir(), "smelting-droplet")
		Ω(err).ShouldNot(HaveOccurred())

		fs = NewFileSystem(runner)
	})

	AfterEach(func() {
		os.RemoveAll(outputDir)
	})

	Describe("GenerateFiles", func() {
		It("copies the built app to app/ in the droplet dir", func() {
			err := fs.GenerateFiles(appDir, outputDir)
			Ω(err).ShouldNot(HaveOccurred())

			Ω(runner).Should(HaveExecutedSerially(
				fake_command_runner.CommandSpec{
					Path: "cp",
					Args: []string{"-a", appDir, path.Join(outputDir, "app")},
				},
			))
		})

		It("creates app/, tmp/, and logs/ in the droplet dir", func() {
			err := fs.GenerateFiles(appDir, outputDir)
			Ω(err).ShouldNot(HaveOccurred())

			fileInfo, err := os.Stat(path.Join(outputDir, "tmp"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeTrue())

			fileInfo, err = os.Stat(path.Join(outputDir, "logs"))
			Ω(err).ShouldNot(HaveOccurred())
			Ω(fileInfo.IsDir()).Should(BeTrue())
		})

		Context("when copying the app fails", func() {
			disaster := errors.New("fresh outta disk space")

			BeforeEach(func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "cp",
				}, func(*exec.Cmd) error {
					return disaster
				})
			})

			It("returns the error", func() {
				err := fs.GenerateFiles(appDir, outputDir)
				Ω(err).Should(Equal(disaster))
			})
		})
	})
})
