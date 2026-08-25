package tunnel

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

func defaultManagedFRPRuntimeDirectory() (string, error) {
	return managedFRPRuntimeDirectory(os.Getenv, os.UserHomeDir, runtime.GOOS)
}

func managedFRPRuntimeDirectory(environment func(string) string, userHomeDirectory func() (string, error), platform string) (string, error) {
	if environment == nil {
		environment = os.Getenv
	}
	if environment("YCY_TUNNEL_DOCKER") == "1" {
		return filepath.Join("/opt", "ycy", "frp", FRPVersion), nil
	}
	stateRoot, err := tunnelStateRoot(environment, userHomeDirectory, platform)
	if err != nil {
		return "", err
	}
	return filepath.Join(stateRoot, "ycy", "frp", FRPVersion), nil
}

func tunnelStateRoot(environment func(string) string, userHomeDirectory func() (string, error), platform string) (string, error) {
	if environment == nil {
		environment = os.Getenv
	}
	if platform == "windows" {
		if root := environment("LOCALAPPDATA"); root != "" {
			return root, nil
		}
	}
	if platform == "linux" {
		if root := environment("XDG_STATE_HOME"); root != "" {
			return root, nil
		}
	}
	if userHomeDirectory == nil {
		userHomeDirectory = os.UserHomeDir
	}
	home, err := userHomeDirectory()
	if err != nil {
		return "", fmt.Errorf("resolve Tunnel state root: %w", err)
	}
	switch platform {
	case "windows":
		return filepath.Join(home, "AppData", "Local"), nil
	case "darwin":
		return filepath.Join(home, "Library", "Application Support"), nil
	default:
		return filepath.Join(home, ".local", "state"), nil
	}
}
