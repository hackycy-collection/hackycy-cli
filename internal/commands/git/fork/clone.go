package fork

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
)

// CloneOutput is the captured outcome of one external Git clone invocation.
type CloneOutput struct {
	Stderr   []byte
	ExitCode int
}

// CloneRunner is Git Fork's command-owned external-Git clone boundary.
type CloneRunner interface {
	Run(context.Context, []string) (CloneOutput, error)
}

// DirectoryRemover removes the clone metadata after a successful clone.
type DirectoryRemover interface {
	RemoveAll(string) error
}

// CloneFallback performs the legacy shallow-clone fallback and removes its Git metadata.
func CloneFallback(ctx context.Context, runner CloneRunner, remover DirectoryRemover, repository Repository, destination string) error {
	if runner == nil {
		return errors.New("git fork clone runner is required")
	}
	if remover == nil {
		return errors.New("git fork clone metadata remover is required")
	}
	arguments := []string{"clone", "--depth=1", "--single-branch"}
	if repository.Ref != "" {
		arguments = append(arguments, "--branch", repository.Ref)
	}
	arguments = append(arguments, CloneURL(repository), destination)
	output, err := runner.Run(ctx, arguments)
	if err != nil {
		return err
	}
	if output.ExitCode != 0 {
		message := strings.TrimSpace(string(output.Stderr))
		if message == "" {
			message = fmt.Sprintf("git clone failed with exit code %d", output.ExitCode)
		}
		return errors.New(message)
	}
	return remover.RemoveAll(filepath.Join(destination, ".git"))
}
