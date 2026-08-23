package cm

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestStageSelectedChangesLeavesTheIndexUntouchedForCancelledOrEmptySelection(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		selected  []string
		cancelled bool
	}{
		{name: "cancelled", cancelled: true},
		{name: "empty selection"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			runner := newStageRunner(root, " M changed.go\x00")
			prompter := stagePrompter{selected: testCase.selected, cancelled: testCase.cancelled}

			result, err := StageSelectedChanges(context.Background(), runner, diskSnapshotFileSystem{}, &prompter)
			if err != nil {
				t.Fatalf("StageSelectedChanges() error = %v", err)
			}
			want := StageResult{RepositoryRoot: root, Cancelled: testCase.cancelled, NothingSelected: !testCase.cancelled}
			if result != want {
				t.Fatalf("result = %#v, want %#v", result, want)
			}
			runner.requireCalls(t,
				[]string{"rev-parse", "--show-toplevel"},
				[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
				[]string{"-C", root, "ls-files", "--stage", "-z"},
			)
		})
	}
}

func TestStageSelectedChangesSelectsEveryDisplayedPathByDefault(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "changed.go")
	writeStageFile(t, root, "new.go")
	runner := newStageRunner(root, " M changed.go\x00?? new.go\x00")
	prompter := stagePrompter{useInitialValues: true}

	result, err := StageSelectedChanges(context.Background(), runner, diskSnapshotFileSystem{}, &prompter)
	if err != nil {
		t.Fatalf("StageSelectedChanges() error = %v", err)
	}
	if result != (StageResult{RepositoryRoot: root}) {
		t.Fatalf("result = %#v", result)
	}
	wantPrompt := StagePrompt{
		Message: "Select files to stage",
		Options: []StageOption{
			{Value: "changed.go", Label: "M changed.go"},
			{Value: "new.go", Label: "A new.go"},
		},
		InitialValues: []string{"changed.go", "new.go"},
	}
	if !reflect.DeepEqual(prompter.prompt, wantPrompt) {
		t.Fatalf("prompt = %#v, want %#v", prompter.prompt, wantPrompt)
	}
	runner.requireCalls(t,
		[]string{"rev-parse", "--show-toplevel"},
		[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
		[]string{"-C", root, "ls-files", "--stage", "-z"},
		[]string{"-C", root, "add", "-A", "--", "changed.go", "new.go"},
	)
}

func TestStageSelectedChangesRestoresUnselectedTrackedPaths(t *testing.T) {
	root := t.TempDir()
	writeStageFile(t, root, "staged.go")
	writeStageFile(t, root, "worktree.go")
	runner := newStageRunner(root, "M  staged.go\x00 M worktree.go\x00")
	prompter := stagePrompter{selected: []string{"worktree.go"}}

	_, err := StageSelectedChanges(context.Background(), runner, diskSnapshotFileSystem{}, &prompter)
	if err != nil {
		t.Fatalf("StageSelectedChanges() error = %v", err)
	}
	runner.requireCalls(t,
		[]string{"rev-parse", "--show-toplevel"},
		[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
		[]string{"-C", root, "ls-files", "--stage", "-z"},
		[]string{"-C", root, "restore", "--staged", "--", "staged.go"},
		[]string{"-C", root, "add", "-A", "--", "worktree.go"},
	)
}

func TestStageSelectedChangesSplitsExistingAndMissingPathsWithoutChangingPathspecs(t *testing.T) {
	root := t.TempDir()
	magicPath := ":(glob)*.go"
	writeStageFile(t, root, magicPath)
	runner := newStageRunner(root, "?? "+magicPath+"\x00 D deleted.go\x00")
	prompter := stagePrompter{selected: []string{magicPath, "deleted.go"}}

	_, err := StageSelectedChanges(context.Background(), runner, diskSnapshotFileSystem{}, &prompter)
	if err != nil {
		t.Fatalf("StageSelectedChanges() error = %v", err)
	}
	runner.requireCalls(t,
		[]string{"rev-parse", "--show-toplevel"},
		[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
		[]string{"-C", root, "ls-files", "--stage", "-z"},
		[]string{"-C", root, "add", "-A", "--", magicPath},
		[]string{"-C", root, "update-index", "--remove", "--", "deleted.go"},
	)
}

func TestStageSelectedChangesReportsNoChangesWithoutPromptOrMutation(t *testing.T) {
	root := t.TempDir()
	runner := newStageRunner(root, "")
	prompter := stagePrompter{}

	result, err := StageSelectedChanges(context.Background(), runner, diskSnapshotFileSystem{}, &prompter)
	if err != nil {
		t.Fatalf("StageSelectedChanges() error = %v", err)
	}
	if result != (StageResult{RepositoryRoot: root, NoChanges: true}) {
		t.Fatalf("result = %#v", result)
	}
	if prompter.called {
		t.Fatal("prompt called with no changes")
	}
	runner.requireCalls(t,
		[]string{"rev-parse", "--show-toplevel"},
		[]string{"-C", root, "status", "--porcelain=v1", "-z", "--untracked-files=all"},
		[]string{"-C", root, "ls-files", "--stage", "-z"},
	)
}

func TestStageAllChangesUsesGitAddAll(t *testing.T) {
	root := t.TempDir()
	runner := &scriptedGitRunner{responses: map[string]GitOutput{
		"rev-parse --show-toplevel": {Stdout: []byte(root + "\n")},
	}}

	gotRoot, err := StageAllChanges(context.Background(), runner)
	if err != nil {
		t.Fatalf("StageAllChanges() error = %v", err)
	}
	if gotRoot != root {
		t.Fatalf("root = %q, want %q", gotRoot, root)
	}
	runner.requireCalls(t,
		[]string{"rev-parse", "--show-toplevel"},
		[]string{"-C", root, "add", "-A"},
	)
}

type stagePrompter struct {
	prompt           StagePrompt
	selected         []string
	cancelled        bool
	useInitialValues bool
	called           bool
}

func (prompter *stagePrompter) SelectFiles(prompt StagePrompt) ([]string, bool) {
	prompter.called = true
	prompter.prompt = prompt
	if prompter.useInitialValues {
		return append([]string(nil), prompt.InitialValues...), prompter.cancelled
	}
	return append([]string(nil), prompter.selected...), prompter.cancelled
}

func newStageRunner(root, status string) *scriptedGitRunner {
	return &scriptedGitRunner{responses: map[string]GitOutput{
		"rev-parse --show-toplevel":                                      {Stdout: []byte(root + "\n")},
		"-C " + root + " status --porcelain=v1 -z --untracked-files=all": {Stdout: []byte(status)},
		"-C " + root + " ls-files --stage -z":                            {},
	}}
}

func writeStageFile(t *testing.T, root, relativePath string) {
	t.Helper()
	filePath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(filePath), 0o700); err != nil {
		t.Fatalf("create parent directory: %v", err)
	}
	if err := os.WriteFile(filePath, []byte("content\n"), 0o600); err != nil {
		t.Fatalf("write stage fixture: %v", err)
	}
}
