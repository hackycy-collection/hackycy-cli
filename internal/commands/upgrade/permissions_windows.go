//go:build windows

package upgrade

import (
	"os"

	"github.com/hackycy/hackycy-cli/internal/windowsacl"
)

func protectUpgradePath(path string, _ os.FileMode, _ func(string, os.FileMode) error) error {
	return windowsacl.RestrictPrivatePath(path)
}
