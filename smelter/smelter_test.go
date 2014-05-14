package smelter_test

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/cloudfoundry-incubator/runtime-schema/models"
	"github.com/cloudfoundry/gunk/command_runner/fake_command_runner"
	. "github.com/cloudfoundry/gunk/command_runner/fake_command_runner/matchers"

	. "github.com/cloudfoundry-incubator/linux-smelter/smelter"
)

var _ = Describe("Smelter", func() {
	var smelter *Smelter
	var runner *fake_command_runner.FakeCommandRunner

	var (
		smeltingDir            string
		appDir                 string
		outputDir              string
		resultDir              string
		buildpacksDir          string
		buildArtifactsCacheDir string
		config                 models.LinuxSmeltingConfig
	)

	BeforeEach(func() {
		runner = fake_command_runner.New()

		var err error

		smeltingDir, err = ioutil.TempDir(os.TempDir(), "smelting")
		Ω(err).ShouldNot(HaveOccurred())

		appDir = path.Join(smeltingDir, "app")
		outputDir = path.Join(smeltingDir, "output")
		resultDir = path.Join(smeltingDir, "result")
		buildpacksDir = path.Join(smeltingDir, "buildpacks")
		buildArtifactsCacheDir = path.Join(smeltingDir, "cache")

		config = models.NewLinuxSmeltingConfig([]string{"a", "b", "c"})
		config.Set(models.LinuxSmeltingAppDirFlag, appDir)
		config.Set(models.LinuxSmeltingOutputDirFlag, outputDir)
		config.Set(models.LinuxSmeltingResultDirFlag, resultDir)
		config.Set(models.LinuxSmeltingBuildArtifactsCacheDirFlag, buildArtifactsCacheDir)
		config.Set(models.LinuxSmeltingBuildpacksDirFlag, buildpacksDir)

		os.MkdirAll(path.Join(config.BuildpackPath("a"), "bin"), 0777)
		os.MkdirAll(path.Join(config.BuildpackPath("b"), "bin"), 0777)
		os.MkdirAll(path.Join(config.BuildpackPath("c"), "inner", "bin"), 0777)

		smelter = New(&config, runner)
	})

	AfterEach(func() {
		os.RemoveAll(smeltingDir)
	})

	Describe("smelting", func() {
		Context("when a buildpack successfully detects", func() {
			BeforeEach(func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("a"), "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("b"), "bin", "detect"),
				}, func(cmd *exec.Cmd) error {
					// detected!
					cmd.Stdout.Write([]byte("Always Matching\n"))
					return nil
				})
			})

			setupSuccessfulRelease := func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("b"), "bin", "release"),
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte(`---
default_process_types:
  web: ./some-start-command
`))
					return nil
				})
			}

			It("stops trying to detect other buildpacks", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()

				Ω(runner).Should(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("a"), "bin", "detect"),
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "detect"),
						Args: []string{appDir},
					},
				))

				Ω(runner).ShouldNot(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("c"), "bin", "detect"),
						Args: []string{appDir},
					},
				))

				Ω(err).ShouldNot(HaveOccurred())
			})

			It("runs bin/compile on the first matching buildpack", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				Ω(runner).Should(HaveExecutedSerially(
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("a"), "bin", "detect"),
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "detect"),
						Args: []string{appDir},
					},
					fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "compile"),
						Args: []string{appDir, buildArtifactsCacheDir},
					},
				))
			})

			It("copies the built app to app/ in the output dir", func() {
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

			It("creates app/, tmp/, and logs/ in the output dir", func() {
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

			It("writes the detected buildpack to staging_info.yml in the output dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				file, err := os.Open(path.Join(outputDir, "staging_info.yml"))
				Ω(err).ShouldNot(HaveOccurred())

				err = candiedyaml.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedBuildpack).Should(Equal("Always Matching"))
			})

			It("writes the detected buildpack to result.json in the result dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(resultDir, "result.json"))
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				err = json.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedBuildpack).Should(Equal("Always Matching"))
			})

			It("writes the detected start command to staging_info.yml in the output dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				file, err := os.Open(path.Join(outputDir, "staging_info.yml"))
				Ω(err).ShouldNot(HaveOccurred())

				err = candiedyaml.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedStartCommand).Should(Equal("./some-start-command"))
			})

			It("writes the detected start command to result.json in the result dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(resultDir, "result.json"))
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				err = json.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedStartCommand).Should(Equal("./some-start-command"))
			})

			It("writes the buildpack key to result.json in the result dir", func() {
				setupSuccessfulRelease()

				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(resultDir, "result.json"))
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				err = json.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.BuildpackKey).Should(Equal("b"))
			})

			Context("when bin/release has a start command", func() {
				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "release"),
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

					var output models.StagingInfo

					err = candiedyaml.NewDecoder(file).Decode(&output)
					Ω(err).ShouldNot(HaveOccurred())

					Ω(output.DetectedStartCommand).Should(Equal("some-command"))
				})
			})

			Context("when bin/compile fails", func() {
				disaster := errors.New("buildpack blew up")

				BeforeEach(func() {
					runner.WhenRunning(fake_command_runner.CommandSpec{
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "compile"),
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
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "release"),
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
						Path: filepath.Join(config.BuildpackPath("b"), "bin", "release"),
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

		Context("when the buildpack is nested under a directory (can happen with zip buildpacks served by github)", func() {
			BeforeEach(func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("a"), "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("b"), "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("c"), "inner", "bin", "detect"),
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte("C Buildpack\n"))
					return nil
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("c"), "inner", "bin", "release"),
				}, func(cmd *exec.Cmd) error {
					cmd.Stdout.Write([]byte("--- {}\n"))
					return nil
				})
			})

			It("should dive into the nested directory", func() {
				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(outputDir, "staging_info.yml"))
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				err = candiedyaml.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.DetectedBuildpack).Should(Equal("C Buildpack"))
			})

			It("writes the correct buildpack key to result.json", func() {
				err := smelter.Smelt()
				Ω(err).ShouldNot(HaveOccurred())

				file, err := os.Open(path.Join(resultDir, "result.json"))
				Ω(err).ShouldNot(HaveOccurred())

				var output models.StagingInfo

				err = json.NewDecoder(file).Decode(&output)
				Ω(err).ShouldNot(HaveOccurred())

				Ω(output.BuildpackKey).Should(Equal("c"))
			})
		})

		Context("when no buildpacks match the app", func() {
			It("returns a NoneDetectedError", func() {
				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("a"), "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("b"), "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				runner.WhenRunning(fake_command_runner.CommandSpec{
					Path: filepath.Join(config.BuildpackPath("c"), "inner", "bin", "detect"),
				}, func(*exec.Cmd) error {
					// detection failed
					return errors.New("exit status 1")
				})

				err := smelter.Smelt()
				Ω(err).Should(Equal(NoneDetectedError{AppDir: appDir}))
			})
		})
	})
})
