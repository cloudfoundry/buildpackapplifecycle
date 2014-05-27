package smelter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry-incubator/runtime-schema/models"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/cloudfoundry/gunk/command_runner"
)

type Smelter struct {
	config *models.LinuxSmeltingConfig

	runner command_runner.CommandRunner
}

type descriptiveError struct {
	message string
	err     error
}

func (e descriptiveError) Error() string {
	if e.err == nil {
		return e.message
	}
	return fmt.Sprintf("%s: %s", e.message, e.err.Error())
}

func newDescriptiveError(err error, message string, args ...interface{}) error {
	if len(args) == 0 {
		return descriptiveError{message: message, err: err}
	}
	return descriptiveError{message: fmt.Sprintf(message, args...), err: err}
}

type Release struct {
	DefaultProcessTypes struct {
		Web string `yaml:"web"`
	} `yaml:"default_process_types"`
}

func New(
	config *models.LinuxSmeltingConfig,
	runner command_runner.CommandRunner,
) *Smelter {
	return &Smelter{
		config: config,
		runner: runner,
	}
}

func (s *Smelter) Smelt() error {
	//set up the world
	err := s.makeDirectories()
	if err != nil {
		return newDescriptiveError(err, "failed to set up filesystem when generating droplet")
	}

	//detect, compile, release
	detectedBuildpack, detectedBuildpackDir, detectOutput, err := s.detect()
	if err != nil {
		return err
	}

	err = s.compile(detectedBuildpackDir)
	if err != nil {
		return newDescriptiveError(err, "failed to compile droplet")
	}

	releaseInfo, err := s.release(detectedBuildpackDir)
	if err != nil {
		return newDescriptiveError(err, "failed to build droplet release")
	}

	//generate staging_info.yml and result json file
	err = s.saveInfo(detectedBuildpack, detectOutput, releaseInfo)
	if err != nil {
		return newDescriptiveError(err, "failed to encode generated metadata")
	}

	//prepare the final droplet directory
	err = s.copyApp(s.config.AppDir(), path.Join(s.config.OutputDir(), "app"))
	if err != nil {
		return newDescriptiveError(err, "failed to copy compiled droplet")
	}

	err = os.MkdirAll(path.Join(s.config.OutputDir(), "tmp"), 0755)
	if err != nil {
		return newDescriptiveError(err, "failed to set up droplet filesystem")
	}

	err = os.MkdirAll(path.Join(s.config.OutputDir(), "logs"), 0755)
	if err != nil {
		return newDescriptiveError(err, "failed to set up droplet filesystem")
	}

	return nil
}

func (s *Smelter) makeDirectories() error {
	if err := os.MkdirAll(s.config.OutputDir(), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(s.config.ResultJsonDir(), 0755); err != nil {
		return err
	}

	if err := os.MkdirAll(s.config.BuildArtifactsCacheDir(), 0755); err != nil {
		return err
	}

	return nil
}

func (s *Smelter) buildpackPath(buildpack string) (string, error) {
	buildpackPath := s.config.BuildpackPath(buildpack)

	if s.pathHasBinDirectory(buildpackPath) {
		return buildpackPath, nil
	}

	files, err := ioutil.ReadDir(buildpackPath)
	if err != nil {
		return "", newDescriptiveError(nil, "failed to read buildpack directory for buildpack: %s", buildpack)
	}

	if len(files) == 1 {
		nestedPath := path.Join(buildpackPath, files[0].Name())

		if s.pathHasBinDirectory(nestedPath) {
			return nestedPath, nil
		}
	}

	return "", newDescriptiveError(nil, "malformed buildpack does not contain a /bin dir: %s", buildpack)
}

func (s *Smelter) pathHasBinDirectory(pathToTest string) bool {
	_, err := os.Stat(path.Join(pathToTest, "bin"))
	return err == nil
}

func (s *Smelter) detect() (string, string, string, error) {
	for _, buildpack := range s.config.BuildpackOrder() {
		output := new(bytes.Buffer)

		buildpackPath, err := s.buildpackPath(buildpack)

		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		err = s.runner.Run(&exec.Cmd{
			Path:   path.Join(buildpackPath, "bin", "detect"),
			Args:   []string{s.config.AppDir()},
			Stdout: output,
			Stderr: os.Stderr,
		})

		if err == nil {
			return buildpack, buildpackPath, strings.TrimRight(output.String(), "\n"), nil
		}
	}

	return "", "", "", newDescriptiveError(nil, "no valid buildpacks detected")
}

func (s *Smelter) compile(buildpackDir string) error {
	return s.runner.Run(&exec.Cmd{
		Path:   path.Join(buildpackDir, "bin", "compile"),
		Args:   []string{s.config.AppDir(), s.config.BuildArtifactsCacheDir()},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

func (s *Smelter) release(buildpackDir string) (Release, error) {
	releaseOut := new(bytes.Buffer)

	release := &exec.Cmd{
		Path:   path.Join(buildpackDir, "bin", "release"),
		Args:   []string{s.config.AppDir()},
		Stderr: os.Stderr,
		Stdout: releaseOut,
	}

	err := s.runner.Run(release)
	if err != nil {
		return Release{}, err
	}

	decoder := candiedyaml.NewDecoder(releaseOut)

	var parsedRelease Release

	err = decoder.Decode(&parsedRelease)
	if err != nil {
		return Release{}, newDescriptiveError(err, "buildpack's release output invalid")
	}

	return parsedRelease, nil
}

func (s *Smelter) saveInfo(buildpack string, detectOutput string, releaseInfo Release) error {
	infoFile, err := os.Create(filepath.Join(s.config.OutputDir(), "staging_info.yml"))
	if err != nil {
		return err
	}

	defer infoFile.Close()

	resultFile, err := os.Create(s.config.ResultJsonPath())
	if err != nil {
		return err
	}

	defer resultFile.Close()

	info := models.StagingInfo{
		BuildpackKey:         buildpack,
		DetectedBuildpack:    detectOutput,
		DetectedStartCommand: releaseInfo.DefaultProcessTypes.Web,
	}

	err = candiedyaml.NewEncoder(infoFile).Encode(info)
	if err != nil {
		return err
	}

	err = json.NewEncoder(resultFile).Encode(info)
	if err != nil {
		return err
	}

	return nil
}

func (s *Smelter) copyApp(appDir, stageDir string) error {
	return s.runner.Run(&exec.Cmd{
		Path:   "cp",
		Args:   []string{"-a", appDir, stageDir},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
