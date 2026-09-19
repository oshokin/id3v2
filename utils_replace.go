//go:build !windows

package id3v2

import "os"

// replaceFile replaces dst with src. Unix rename overwrites dst atomically.
func replaceFile(src, dst string) error {
	return os.Rename(src, dst)
}
