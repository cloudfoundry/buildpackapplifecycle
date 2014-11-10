package buildpackrunner

import (
	"fmt"
	"net/url"
	"os/exec"
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

	args := []string{
		"clone",
		"-depth",
		"1",
	}

	if branch != "" {
		args = append(args, "-b", branch)
	}

	args = append(args, "--recursive", gitUrl, destination)
	cmd := exec.Command(gitPath, args...)

	err = cmd.Run()

	if err != nil {
		cmd = exec.Command(gitPath, "clone", "--recursive", gitUrl, destination)
		err = cmd.Run()
		if err != nil {
			gitArgs := strings.Join(cmd.Args, " ")
			return fmt.Errorf("git clone failed: cmd: '%s' err: %s", gitArgs, err.Error())
		}

		if branch != "" {
			cmd = exec.Command(gitPath, "--git-dir="+destination+"/.git", "--work-tree="+destination, "checkout", branch)
			err = cmd.Run()
			if err != nil {
				gitArgs := strings.Join(cmd.Args, " ")
				return fmt.Errorf("git checkout failed: cmd: '%s' err: %s", gitArgs, err.Error())
			}
		}
	}

	return nil
}
