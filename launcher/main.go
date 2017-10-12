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
	"strings"

	"code.cloudfoundry.org/buildpackapplifecycle/containerpath"
	"code.cloudfoundry.org/buildpackapplifecycle/databaseuri"

	"github.com/cloudfoundry-incubator/credhub-cli/credhub"

	yaml "gopkg.in/yaml.v2"
)

type PlatformOptions struct {
	CredhubURI string `json:"credhub_uri"`
}

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

	platformOptions, err := platformOptions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid platform options: %v", err)
		os.Exit(3)
	}
	if platformOptions != nil && platformOptions.CredhubURI != "" {
		ch, err := credhubClient(platformOptions.CredhubURI)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to set up credhub client: %v", err)
			os.Exit(4)
		}
		interpolatedServices, err := ch.InterpolateString(os.Getenv("VCAP_SERVICES"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to interpolate credhub references: %v", err)
			os.Exit(5)
		}
		os.Setenv("VCAP_SERVICES", interpolatedServices)
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

func credhubClient(credhubURI string) (*credhub.CredHub, error) {
	if os.Getenv("CF_INSTANCE_CERT") == "" || os.Getenv("CF_INSTANCE_KEY") == "" {
		return nil, fmt.Errorf("Missing CF_INSTANCE_CERT and/or CF_INSTANCE_KEY")
	}
	if os.Getenv("CF_SYSTEM_CERTS_PATH") == "" {
		return nil, fmt.Errorf("Missing CF_SYSTEM_CERTS_PATH")
	}

	systemCertsPath := containerpath.For(os.Getenv("CF_SYSTEM_CERTS_PATH"))
	caCerts := []string{}
	files, err := ioutil.ReadDir(systemCertsPath)
	if err != nil {
		return nil, fmt.Errorf("Can't read contents of system cert path: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".crt") {
			contents, err := ioutil.ReadFile(filepath.Join(systemCertsPath, file.Name()))
			if err != nil {
				return nil, fmt.Errorf("Can't read contents of cert in system cert path: %v", err)
			}
			caCerts = append(caCerts, string(contents))
		}
	}

	return credhub.New(
		credhubURI,
		credhub.ClientCert(containerpath.For(os.Getenv("CF_INSTANCE_CERT")), containerpath.For(os.Getenv("CF_INSTANCE_KEY"))),
		credhub.CaCerts(caCerts...),
	)
}

var cachedPlatformOptions *PlatformOptions

func platformOptions() (*PlatformOptions, error) {
	if cachedPlatformOptions == nil {
		jsonPlatformOptions := os.Getenv("VCAP_PLATFORM_OPTIONS")
		if jsonPlatformOptions != "" {
			platformOptions := PlatformOptions{}
			err := json.Unmarshal([]byte(jsonPlatformOptions), &platformOptions)
			if err != nil {
				return nil, err
			}
			cachedPlatformOptions = &platformOptions
		}
		os.Unsetenv("VCAP_PLATFORM_OPTIONS")
	}
	return cachedPlatformOptions, nil
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
