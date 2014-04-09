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

func (fs *FileSystem) GenerateFiles(appDir, stageDir string) error {
	err := fs.copyApp(appDir, stageDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(stageDir, "tmp"), 0755)
	if err != nil {
		return err
	}

	err = os.MkdirAll(path.Join(stageDir, "logs"), 0755)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FileSystem) copyApp(appDir, stageDir string) error {
	return fs.runner.Run(&exec.Cmd{
		Path:   "cp",
		Args:   []string{"-a", appDir, path.Join(stageDir, "app")},
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	})
}
