package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/cloudfoundry-incubator/linux-circus/protocol"
)

const soldier = `
cd "$1"

if [ -d .profile.d ]; then
  for env_file in .profile.d/*; do
    source $env_file
  done
fi

shift

eval "$@"
`

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: %s <app directory> <start command> <metadata>", os.Args[0])
		os.Exit(1)
	}

	dir := os.Args[1]
	startCommand := os.Args[2]
	metadata := os.Args[3]

	os.Setenv("HOME", dir)
	os.Setenv("TMPDIR", filepath.Join(dir, "tmp"))

	vcapAppEnv := map[string]interface{}{}
	err := json.Unmarshal([]byte(os.Getenv("VCAP_APPLICATION")), &vcapAppEnv)
	if err == nil {
		vcapAppEnv["host"] = "0.0.0.0"

		vcapAppEnv["instance_id"] = os.Getenv("CF_INSTANCE_GUID")

		port, err := strconv.Atoi(os.Getenv("PORT"))
		if err == nil {
			vcapAppEnv["port"] = port
		}

		index, err := strconv.Atoi(os.Getenv("CF_INSTANCE_INDEX"))
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
		var executionMetadata protocol.ExecutionMetadata
		err := json.Unmarshal([]byte(metadata), &executionMetadata)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid metadata - %s", err)
			os.Exit(1)
		} else {
			command = executionMetadata.StartCommand
		}
	}

	syscall.Exec("/bin/bash", []string{
		"bash",
		"-c",
		soldier,
		os.Args[0],
		dir,
		command,
	}, os.Environ())
}
