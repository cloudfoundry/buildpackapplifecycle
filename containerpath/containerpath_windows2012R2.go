// +build windows,windows2012R2

package containerpath

import (
	"os"
	"path/filepath"
)

func For(path string) string {
	return filepath.Join(filepath.Clean(os.Getenv("USERPROFILE")), path)
}
