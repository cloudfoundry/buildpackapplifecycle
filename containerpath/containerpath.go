// +build !windows2012R2

package containerpath

import (
	"path/filepath"
)

func For(path string) string {
	return filepath.Clean(path)
}
