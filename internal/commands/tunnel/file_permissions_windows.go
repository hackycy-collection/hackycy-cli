//go:build windows

package tunnel

import (
	"os"

	"github.com/hackycy/hackycy-cli/internal/windowsacl"
)

func protectTunnelFile(path string, _ os.FileMode) error {
	return windowsacl.RestrictPrivatePath(path)
}
