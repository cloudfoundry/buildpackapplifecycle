package buildpackrunner

import "path/filepath"

func hasFinalize(buildpackPath string) (bool, error) {
	return windowsExecutableExists(filepath.Join(buildpackPath, "bin", "finalize"))
}

func hasSupply(buildpackPath string) (bool, error) {
	return windowsExecutableExists(filepath.Join(buildpackPath, "bin", "supply"))
}

func windowsExecutableExists(file string) (bool, error) {
	extensions := []string{".bat", ".exe", ".cmd"}

	for _, exe := range extensions {
		exists, err := fileExists(file + exe)
		if err != nil {
			return false, err
		} else if exists {
			return true, nil
		}
	}

	return false, nil
}
