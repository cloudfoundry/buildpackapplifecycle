package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/buildpackapplifecycle/buildpackrunner"
	"code.cloudfoundry.org/buildpackapplifecycle/env"
	"code.cloudfoundry.org/goshims/osshim"
	yaml "gopkg.in/yaml.v2"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "%s: received only %d arguments\n", os.Args[0], len(os.Args)-1)
		fmt.Fprintf(os.Stderr, "Usage: %s <app-directory> <start-command> <metadata>", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]
	startCommand := os.Args[2]

	absDir, err := filepath.Abs(dir)
	if err == nil {
		dir = absDir
	}

	stagingInfo, err := unmarhsalStagingInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid staging info - %s", err)
		os.Exit(1)
	}

	var command string
	if startCommand != "" {
		command = startCommand
	} else {
		command = stagingInfo.StartCommand
	}

	if command == "" {
		fmt.Fprintf(os.Stderr, "%s: no start command specified or detected in droplet", os.Args[0])
		os.Exit(1)
	}

	if err := env.CalcEnv(&osshim.OsShim{}, dir); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(3)
	}

	runtime.GOMAXPROCS(1)
	runProcess(dir, command, stagingInfo.GetEntrypointPrefix())
}

func unmarhsalStagingInfo() (buildpackrunner.DeaStagingInfo, error) {
	stagingInfo := buildpackrunner.DeaStagingInfo{}
	stagingInfoData, err := ioutil.ReadFile(buildpackrunner.DeaStagingInfoFilename)
	if err != nil {
		if os.IsNotExist(err) {
			return stagingInfo, nil
		}
		return stagingInfo, err
	}

	err = yaml.Unmarshal(stagingInfoData, &stagingInfo)
	if err != nil {
		return stagingInfo, fmt.Errorf("failed to unmarshal %s: %w", buildpackrunner.DeaStagingInfoFilename, err)
	}

	return stagingInfo, nil
}
