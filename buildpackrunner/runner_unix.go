// +build !windows

package buildpackrunner

import "path/filepath"

func hasFinalize(buildpackPath string) (bool, error) {
	return fileExists(filepath.Join(buildpackPath, "bin", "finalize"))
}

func hasSupply(buildpackPath string) (bool, error) {
	return fileExists(filepath.Join(buildpackPath, "bin", "supply"))
}
