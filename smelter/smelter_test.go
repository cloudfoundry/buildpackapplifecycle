package smelter_test

import (
	"encoding/json"
	"errors"
	"github.com/fraenkel/candiedyaml"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	. "github.com/cloudfoundry-incubator/linux-smelter/smelter"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type ExpectedStagingResult struct {
	DetectedBuildpack string `yaml:"detected_buildpack" json:"detected_buildpack"`
	StartCommand      string `yaml:"start_command" json:"-"`
}

var _ = Describe("Smelter", func() {
	var smelter *Smelter
	var runner *fake_command_runner.FakeCommandRunner

	var (
		appDir    string
		outputDir string
		cacheDir  string
	)

	BeforeEach(func() {
		runner = fake_command_runner.New()

		var err error

		appDir, err = ioutil.TempDir(os.TempDir(), "smelting-app")
		Ω(err).ShouldNot(HaveOccurred())

		outputDir, err = ioutil.TempDir(os.TempDir(), "smelting-droplet")
		Ω(err).ShouldNot(HaveOccurred())

		cacheDir, err = ioutil.TempDir(os.TempDir(), "smelting-cache")
		Ω(err).ShouldNot(HaveOccurred())

		smelter = New(
			appDir,
			outputDir,
			[]string{"/buildpacks/a", "/buildpacks/b", "/buildpacks/c"},
			cacheDir,
			runner,
		)
	})

	AfterEach(func() {
		os.RemoveAll(appDir)
		os.RemoveAll(outputDir)
	})

	Describe("smelting", func() {
		Context("when a buildpack successfully detects", func() {
			BeforeEach(func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/buildpacks/a/bin/detect",
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/buildpacks/b/bin/detect",
				}, func(cmd *exec.Cmd) error {
					// detected!
					cmd.Stdout.Write([]byte("Always Matching\n"))
					return nil
				})
			})

			setupSuccessfulRelease := func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: "/buildpacks/b/bin/release",
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte("--- {}\n"))
					return nil
				})
			}

			It("stops trying to detect other buildpacks", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(runner).Should(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/a/bin/detect",
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/detect",
						Args: []string{appDir},
					},
				))

				Ω(runner).ShouldNot(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/c/bin/detect",
						Args: []string{appDir},
					},
				))
			})

			It("runs bin/compile on the first matching buildpack", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(runner).Should(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/a/bin/detect",
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/detect",
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/compile",
						Args: []string{appDir, cacheDir},
					},
				))
			})

			It("copies the built app to app/ in the droplet dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(runner).Should(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: "cp",
						Args: []string{"-a", appDir, path.Join(outputDir, "app")},
					},
				))
			})

			It("creates app/, tmp/, and logs/ in the droplet dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				fileInfo, err := os.Stat(path.Join(outputDir, "tmp"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(fileInfo.IsDir()).Should(BeTrue())

				fileInfo, err = os.Stat(path.Join(outputDir, "logs"))
				Ω(err).ShouldNot(HaveOccurred())
				Ω(fileInfo.IsDir()).Should(BeTrue())
			})

			It("writes the detected buildpack to staging_info.yml in the droplet dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				var output ExpectedStagingResult

				file, err := os.Open(path.Join(outputDir, "staging_info.yml"))
				Ω(err).ShouldNot(HaveOccurred())

				err = candiedyaml.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedBuildpack).Should(Equal("Always Matching"))
			})

			It("writes the detected buildpack to result.json in the droplet dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(outputDir, "result.json"))
				Ω(err).ShouldNot(HaveOccurred())

				var output ExpectedStagingResult

				err = json.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedBuildpack).Should(Equal("Always Matching"))
			})

			Context("when bin/release has a start command", func() {
				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/release",
					}, func(cmd *exec.Cmd) error {
						cmd.Stdout.Write([]byte("---\n"))
						cmd.Stdout.Write([]byte("default_process_types:\n"))
						cmd.Stdout.Write([]byte("  web: some-command\n"))
						return nil
					})
				})

				It("writes it to staging_info.yml as start_command", func() {
					err := smelter.Smelt()
					Ω(err).ShouldNot(HaveOccurred())

					file, err := os.Open(path.Join(outputDir, "staging_info.yml"))
					Ω(err).ShouldNot(HaveOccurred())

					var output ExpectedStagingResult

					err = candiedyaml.NewDecoder(file).Decode(&output)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(output.StartCommand).Should(Equal("some-command"))
				})
			})

			Context("when bin/compile fails", func() {
				disaster := errors.New("buildpack blew up")

				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/compile",
					}, func(*exec.Cmd) error {
						return disaster
					})
				})

				It("returns the error", func() {
					err := smelter.Smelt()
					Ω(err).Should(Equal(disaster))
				})
			})

			Context("when bin/release fails", func() {
				disaster := errors.New("buildpack blew up")

				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/release",
					}, func(*exec.Cmd) error {
						return disaster
					})
				})

				It("returns the error", func() {
					err := smelter.Smelt()
					Ω(err).Should(Equal(disaster))
				})
			})

			Context("when bin/release outputs malformed YAML", func() {
				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "/buildpacks/b/bin/release",
					}, func(cmd *exec.Cmd) error {
						cmd.Stdout.Write([]byte("["))
						return nil
					})
				})

				It("returns an MalformedReleaseYAMLError", func() {
					err := smelter.Smelt()

					var expectedError MalformedReleaseYAMLError
					Ω(err).Should(BeAssignableToTypeOf(expectedError))
				})
			})

			Context("when copying the app fails", func() {
				disaster := errors.New("fresh outta disk space")

				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "cp",
					}, func(*exec.Cmd) error {
						return disaster
					})

					setupSuccessfulRelease()
				})

				It("returns the error", func() {
					err := smelter.Smelt()
					Ω(err).Should(Equal(disaster))
				})
			})
		})

		Context("when no buildpacks match the app", func() {
			It("returns a NoneDetectedError", func() {
				for _, name := range []string{"a", "b", "c"} {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: "/buildpacks/" + name + "/bin/detect",
					}, func(*exec.Cmd) error {
						// detection failed
						return errors.New("exit status 1")
					})
				}

				err := smelter.Smelt()
				Ω(err).Should(Equal(NoneDetectedError{AppDir: appDir}))
			})
		})
	})
})
