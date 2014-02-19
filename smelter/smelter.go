package smelter

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cloudfoundry/gunk/command_runner"
	"github.com/kylelemons/go-gypsy/yaml"
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

type MalformedReleaseYAML struct{}

func (e MalformedReleaseYAML) Error() string {
	return fmt.Sprintf("buildpack's release script provided malformed YAML")
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
			return buildpackDir, output.String(), nil
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

func (s *Smelter) release(buildpackDir string) (yaml.Node, error) {
	release := &exec.Cmd{
		Path:   path.Join(buildpackDir, "bin", "release"),
		Args:   []string{s.appDir},
		Stderr: os.Stderr,
	}

	out, err := release.StdoutPipe()
	if err != nil {
		return nil, err
	}

	err = s.runner.Start(release)
	if err != nil {
		return nil, err
	}

	defer s.runner.Wait(release)

	outBuf := bufio.NewReader(out)

	// FIXME(GYPSY)
	// hack around Gypsy's lack of full YAML parsitude
	peeked, err := outBuf.Peek(4)
	if err != nil {
		return nil, err
	}

	if string(peeked) == "---\n" {
		_, err := outBuf.ReadBytes('\n')
		if err != nil {
			return nil, err
		}
	}

	parsed, err := yaml.Parse(outBuf)
	if err != nil || parsed == nil {
		return nil, MalformedReleaseYAML{}
	}

	return parsed, nil
}

func (s *Smelter) saveInfo(detectedName string, releaseInfo yaml.Node) error {
	info := map[string]yaml.Node{
		"detected_buildpack": yaml.Scalar(detectedName),
	}

	command, err := yaml.Child(releaseInfo, ".default_process_types.web")
	if err == nil {
		info["start_command"] = command
	}

	return ioutil.WriteFile(
		filepath.Join(s.outputDir, "staging_info.yml"),
		[]byte(yaml.Render(yaml.Map(info))),
		0644,
	)
}

func (s *Smelter) copyApp() error {
	return s.runner.Run(&exec.Cmd{
		Path:   "cp",
		Args:   []string{"-a", s.appDir, path.Join(s.outputDir, "app")},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
