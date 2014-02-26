package droplet

import (
	"os"
	"os/exec"
	"path"

	"github.com/cloudfoundry/gunk/command_runner"
)

type FileSystem struct {
	runner command_runner.CommandRunner
}

func NewFileSystem(runner command_runner.CommandRunner) *FileSystem {
	return &FileSystem{
		runner: runner,
	}
}

func (fs *FileSystem) GenerateFiles(appDir, outputDir string) error {
	err := fs.copyApp(appDir, outputDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(outputDir, "tmp"), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(outputDir, "logs"), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) copyApp(appDir, outputDir string) error {
	return fs.runner.Run(&exec.Cmd{
		Path:   "cp",
		Args:   []string{"-a", appDir, path.Join(outputDir, "app")},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
