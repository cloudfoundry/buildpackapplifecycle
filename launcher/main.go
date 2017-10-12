package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"code.cloudfoundry.org/buildpackapplifecycle/credhub"
	"code.cloudfoundry.org/buildpackapplifecycle/databaseuri"
	"code.cloudfoundry.org/buildpackapplifecycle/platformoptions"

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
	os.Setenv("HOME", dir)

	tmpDir, err := filepath.Abs(filepath.Join(dir, "..", "tmp"))
	if err == nil {
		os.Setenv("TMPDIR", tmpDir)
	}

	depsDir, err := filepath.Abs(filepath.Join(dir, "..", "deps"))
	if err == nil {
		os.Setenv("DEPS_DIR", depsDir)
	}

	vcapAppEnv := map[string]interface{}{}
	err = json.Unmarshal([]byte(os.Getenv("VCAP_APPLICATION")), &vcapAppEnv)
	if err == nil {
		vcapAppEnv["host"] = "0.0.0.0"

		vcapAppEnv["instance_id"] = os.Getenv("INSTANCE_GUID")

		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			vcapAppEnv["port"] = port
		}

		index, err := strconv.Atoi(os.Getenv("INSTANCE_INDEX"))
		if err == nil {
			vcapAppEnv["instance_index"] = index
		}

		mungedAppEnv, err := json.Marshal(vcapAppEnv)
		if err == nil {
			os.Setenv("VCAP_APPLICATION", string(mungedAppEnv))
		}
	}

	var command string
	if startCommand != "" {
		command = startCommand
	} else {
		command, err = startCommandFromStagingInfo("staging_info.yml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid staging info - %s", err)
			os.Exit(1)
		}
	}

	if command == "" {
		fmt.Fprintf(os.Stderr, "%s: no start command specified or detected in droplet", os.Args[0])
		os.Exit(1)
	}

	if platformOptions, err := platformoptions.Get(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid platform options: %v", err)
		os.Exit(3)
	} else if platformOptions != nil && platformOptions.CredhubURI != "" {
		err := credhub.InterpolateServiceRefs(platformOptions.CredhubURI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to interpolate credhub refs: %v", err)
			os.Exit(4)
		}
	}

	if os.Getenv("DATABASE_URL") == "" {
		dbUri := databaseuri.New()
		if creds, err := dbUri.Credentials([]byte(os.Getenv("VCAP_SERVICES"))); err == nil {
			os.Setenv("DATABASE_URL", dbUri.Uri(creds))
		}
	}

	runtime.GOMAXPROCS(1)
	runProcess(dir, command)
}

type stagingInfo struct {
	StartCommand string `yaml:"start_command"`
}

func startCommandFromStagingInfo(stagingInfoPath string) (string, error) {
	stagingInfoData, err := ioutil.ReadFile(stagingInfoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	info := stagingInfo{}

	err = yaml.Unmarshal(stagingInfoData, &info)
	if err != nil {
		return "", errors.New("invalid YAML")
	}

	return info.StartCommand, nil
}
