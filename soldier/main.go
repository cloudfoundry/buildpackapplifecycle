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

if [ -d "$1/.profile.d" ]; then
  for env_file in "$1"/.profile.d/*; do
    source $env_file
  done
fi

shift

"$@"
`

func main() {
	argv := []string{
		"bash",
		"-c",
		soldier,
	}

	env := []string{
		"HOME=" + os.Args[1],
		"TMPDIR=" + os.Args[1] + "/tmp",
	}

	syscall.Exec("/bin/bash", append(argv, os.Args...), env)
}
