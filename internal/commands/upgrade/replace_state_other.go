//go:build !windows

package upgrade

import "os"

func replaceStateFile(candidate, target string) error {
	return os.Rename(candidate, target)
}
