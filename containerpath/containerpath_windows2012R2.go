// +build windows,windows2012R2

package containerpath

import (
	"path/filepath"
)

func New(root string) *cpath {
	return &cpath{
		root: filepath.Clean(root),
	}
}
