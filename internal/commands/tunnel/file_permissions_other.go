//go:build !windows

package tunnel

import "os"

func protectTunnelFile(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}
