package main

import (
	"encoding/json"
	"os"
	"strconv"
	"syscall"
)

const soldier = `
if [ -z "$1" ]; then
  echo "usage: $0 <app dir> <command to run>" >&2
  exit 1
fi

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
	argv := []string{
		"bash",
		"-c",
		soldier,
	}

	os.Setenv("HOME", os.Args[1])
	os.Setenv("TMPDIR", os.Args[1]+"/tmp")

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

	syscall.Exec("/bin/bash", append(argv, os.Args...), os.Environ())
}
