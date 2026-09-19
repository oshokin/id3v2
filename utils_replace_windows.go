//go:build windows

package id3v2

import (
	"errors"
	"os"
)

// replaceFile replaces dst with src.
// Windows rename cannot overwrite an existing file, so dst is removed first.
func replaceFile(src, dst string) error {
	if err := os.Remove(dst); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return os.Rename(src, dst)
}
