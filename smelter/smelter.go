package smelter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/gunk/command_runner"
	"github.com/fraenkel/candiedyaml"
)

type Smelter struct {
	appDir        string
	outputDir     string
	buildpackDirs []string
	cacheDir      string

	runner command_runner.CommandRunner
}

type NoneDetectedError struct {
	AppDir string
}

func (e NoneDetectedError) Error() string {
	return fmt.Sprintf("no buildpack detected for %s", e.AppDir)
}

type MalformedReleaseYAMLError struct {
	ParseError error
}

func (e MalformedReleaseYAMLError) Error() string {
	return fmt.Sprintf(
		"buildpack's release output invalid: %s",
		e.ParseError,
	)
}

type Release struct {
	DefaultProcessTypes struct {
		Web string `yaml:"web"`
	} `yaml:"default_process_types"`
}

type StagingInfo struct {
	DetectedBuildpack string `yaml:"detected_buildpack"`
	StartCommand      string `yaml:"start_command"`
}

func New(
	appDir string,
	outputDir string,
	buildpackDirs []string,
	cacheDir string,
	runner command_runner.CommandRunner,
) *Smelter {
	return &Smelter{
		appDir:        appDir,
		outputDir:     outputDir,
		buildpackDirs: buildpackDirs,
		cacheDir:      cacheDir,

		runner: runner,
	}
}

func (s *Smelter) Smelt() error {
	detectedBuildpackDir, detectedName, err := s.detect()
	if err != nil {
		return err
	}

	err = s.compile(detectedBuildpackDir)
	if err != nil {
		return err
	}

	releaseInfo, err := s.release(detectedBuildpackDir)
	if err != nil {
		return err
	}

	err = s.saveInfo(detectedName, releaseInfo)
	if err != nil {
		return err
	}

	err = s.copyApp()
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(s.outputDir, "tmp"), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(s.outputDir, "logs"), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (s *Smelter) detect() (string, string, error) {
	for _, buildpackDir := range s.buildpackDirs {
		output := new(bytes.Buffer)

		err := s.runner.Run(&exec.Cmd{
			Path:   path.Join(buildpackDir, "bin", "detect"),
			Args:   []string{s.appDir},
			Stdout: output,
			Stderr: os.Stderr,
		})

		if err == nil {
			return buildpackDir, strings.TrimRight(output.String(), "\n"), nil
		}
	}

	return "", "", NoneDetectedError{AppDir: s.appDir}
}

func (s *Smelter) compile(buildpackDir string) error {
	return s.runner.Run(&exec.Cmd{
		Path:   path.Join(buildpackDir, "bin", "compile"),
		Args:   []string{s.appDir, s.cacheDir},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}

func (s *Smelter) release(buildpackDir string) (Release, error) {
	release := &exec.Cmd{
		Path:   path.Join(buildpackDir, "bin", "release"),
		Args:   []string{s.appDir},
		Stderr: os.Stderr,
	}

	out, err := release.StdoutPipe()
	if err != nil {
		return Release{}, err
	}

	err = s.runner.Start(release)
	if err != nil {
		return Release{}, err
	}

	defer s.runner.Wait(release)

	decoder := candiedyaml.NewDecoder(out)

	var parsedRelease Release

	err = decoder.Decode(&parsedRelease)
	if err != nil {
		return Release{}, MalformedReleaseYAMLError{err}
	}

	return parsedRelease, nil
}

func (s *Smelter) saveInfo(detectedName string, releaseInfo Release) error {
	file, err := os.Create(filepath.Join(s.outputDir, "staging_info.yml"))
	if err != nil {
		return err
	}

	info := StagingInfo{
		DetectedBuildpack: detectedName,
		StartCommand:      releaseInfo.DefaultProcessTypes.Web,
	}

	return candiedyaml.NewEncoder(file).Encode(info)
}

func (s *Smelter) copyApp() error {
	return s.runner.Run(&exec.Cmd{
		Path:   "cp",
		Args:   []string{"-a", s.appDir, path.Join(s.outputDir, "app")},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
