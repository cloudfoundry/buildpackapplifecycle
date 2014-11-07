package buildpackrunner

import (
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
)

func Clone(repo url.URL, destination string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return err
	}

	branch := repo.Fragment
	repo.Fragment = ""
	gitUrl := repo.String()

	baseName := filepath.Base(repo.Path)
	extIndex := strings.LastIndex(baseName, ".")
	if extIndex != -1 {
		baseName = baseName[:extIndex]
	}
	targetDir := filepath.Join(destination, baseName)

	args := []string{
		"clone",
		"-depth",
		"1",
	}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	args = append(args, "--recursive", gitUrl, targetDir)
	cmd := exec.Command(gitPath, args...)

	err = cmd.Run()

	if err != nil {
		cmd = exec.Command(gitPath, "clone", "--recursive", gitUrl, targetDir)
		err = cmd.Run()
		if err != nil {
			gitArgs := strings.Join(cmd.Args, " ")
			return fmt.Errorf("git clone failed: cmd: '%s' err: %s", gitArgs, err.Error())
		}

		if branch != "" {
			cmd = exec.Command(gitPath, "--git-dir="+targetDir+"/.git", "--work-tree="+targetDir, "checkout", branch)
			err = cmd.Run()
			if err != nil {
				gitArgs := strings.Join(cmd.Args, " ")
				return fmt.Errorf("git checkout failed: cmd: '%s' err: %s", gitArgs, err.Error())
			}
		}
	}

	return nil
}
