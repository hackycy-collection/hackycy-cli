package run

import (
	"reflect"
	"testing"
)

func TestSelectScriptPreservesTheLegacyPromptAndScriptHints(t *testing.T) {
	prompter := &recordingRunPrompter{script: "build"}
	scripts := []Script{
		{Name: "check", Command: "go test ./..."},
		{Name: "build", Command: "go build ./cmd/ycy"},
	}

	selected, cancelled := selectScript(prompter, scripts)

	if cancelled || selected != "build" {
		t.Fatalf("selectScript() = (%q, %t)", selected, cancelled)
	}
	want := ScriptPrompt{
		Message: "Select a script to run:",
		Options: []ScriptChoice{
			{Value: "check", Label: "check", Hint: "go test ./..."},
			{Value: "build", Label: "build", Hint: "go build ./cmd/ycy"},
		},
	}
	if !reflect.DeepEqual(prompter.scriptPrompt, want) {
		t.Fatalf("script prompt = %#v, want %#v", prompter.scriptPrompt, want)
	}
}

func TestSelectScriptReportsCancellation(t *testing.T) {
	prompter := &recordingRunPrompter{scriptCancelled: true}
	_, cancelled := selectScript(prompter, []Script{{Name: "check", Command: "go test ./..."}})
	if !cancelled {
		t.Fatal("selectScript() did not report cancellation")
	}
}

func TestSelectPackageManagerPreservesTheComputedOrder(t *testing.T) {
	prompter := &recordingRunPrompter{manager: PackageManagerExternal}
	order := []PackageManager{PackageManagerExternal, PackageManagerPNPM, PackageManagerNPM, PackageManagerYarn}

	selected, cancelled := selectPackageManager(prompter, order)

	if cancelled || selected != PackageManagerExternal {
		t.Fatalf("selectPackageManager() = (%q, %t)", selected, cancelled)
	}
	want := PackageManagerPrompt{
		Message: "Select a package manager:",
		Options: []PackageManagerChoice{
			{Value: PackageManagerExternal, Label: string(PackageManagerExternal)},
			{Value: PackageManagerPNPM, Label: "pnpm"},
			{Value: PackageManagerNPM, Label: "npm"},
			{Value: PackageManagerYarn, Label: "yarn"},
		},
	}
	if !reflect.DeepEqual(prompter.managerPrompt, want) {
		t.Fatalf("manager prompt = %#v, want %#v", prompter.managerPrompt, want)
	}
}

func TestSelectPackageManagerReportsCancellation(t *testing.T) {
	prompter := &recordingRunPrompter{managerCancelled: true}
	_, cancelled := selectPackageManager(prompter, []PackageManager{PackageManagerPNPM})
	if !cancelled {
		t.Fatal("selectPackageManager() did not report cancellation")
	}
}

type recordingRunPrompter struct {
	script           string
	scriptCancelled  bool
	manager          PackageManager
	managerCancelled bool
	scriptPrompt     ScriptPrompt
	managerPrompt    PackageManagerPrompt
}

func (prompter *recordingRunPrompter) SelectScript(prompt ScriptPrompt) (string, bool) {
	prompter.scriptPrompt = prompt
	return prompter.script, prompter.scriptCancelled
}

func (prompter *recordingRunPrompter) SelectPackageManager(prompt PackageManagerPrompt) (PackageManager, bool) {
	prompter.managerPrompt = prompt
	return prompter.manager, prompter.managerCancelled
}
