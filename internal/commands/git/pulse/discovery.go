// Package pulse owns workspace Git activity reporting.
package pulse

import (
	"context"
	"os"
	"path/filepath"
)

const scanYieldEvery = 100

var excludedDirectoryNames = map[string]struct{}{
	"node_modules":     {},
	"vendor":           {},
	"dist":             {},
	".cache":           {},
	"Library":          {},
	".Trash":           {},
	"bower_components": {},
	"__pycache__":      {},
	".venv":            {},
	"venv":             {},
}

// DirectoryReader is the command-owned filesystem boundary for repository scanning.
type DirectoryReader interface {
	ReadDir(string) ([]os.DirEntry, error)
}

// ScanRepositories finds directories that contain a direct .git directory.
// It intentionally keeps walking below a discovered repository so nested repositories remain visible.
func ScanRepositories(ctx context.Context, root string, reader DirectoryReader, onFound func(string), yield func()) ([]string, error) {
	repositories := make([]string, 0)
	stack := []string{root}
	scanned := 0

	for len(stack) > 0 {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		last := len(stack) - 1
		current := stack[last]
		stack = stack[:last]
		scanned++

		entries, err := reader.ReadDir(current)
		if err != nil {
			continue
		}

		hasGitDirectory := false
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			if entry.Name() == ".git" {
				hasGitDirectory = true
				continue
			}
			if _, excluded := excludedDirectoryNames[entry.Name()]; excluded {
				continue
			}
			stack = append(stack, filepath.Join(current, entry.Name()))
		}

		if hasGitDirectory {
			repositories = append(repositories, current)
			if onFound != nil {
				onFound(current)
			}
		}
		if scanned%scanYieldEvery == 0 && yield != nil {
			yield()
		}
	}

	return repositories, nil
}
