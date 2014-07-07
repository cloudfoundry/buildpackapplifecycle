package main

import (
	"os"
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

	syscall.Exec("/bin/bash", append(argv, os.Args...), os.Environ())
}
