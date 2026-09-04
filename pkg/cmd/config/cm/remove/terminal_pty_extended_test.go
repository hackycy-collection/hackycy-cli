package remove

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	terminalexperience "github.com/hackycy/hackycy-cli/internal/terminal"
	"github.com/hackycy/hackycy-cli/internal/terminaltest"
)

func TestRunCMRemoveRichPTYFourWayBJourney(t *testing.T) {
	const helperEnvironment = "YCY_CONFIG_CM_REMOVE_RICH_EXTENDED_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		runCMRemoveExtendedPTYHelper(t)
		return
	}

	for _, testCase := range []struct {
		name          string
		width, height uint16
		color         bool
	}{
		{name: "wide color", width: 120, height: 40, color: true},
		{name: "wide no color", width: 120, height: 40, color: false},
		{name: "compact color", width: 40, height: 15, color: true},
		{name: "compact no color", width: 40, height: 15, color: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			releasePath := filepath.Join(t.TempDir(), "release")
			command := exec.Command(os.Args[0], "-test.run=^TestRunCMRemoveRichPTYFourWayBJourney$")
			command.Env = cmRemoveExtendedPTYEnvironment(map[string]string{
				"NO_COLOR":                       map[bool]string{true: "", false: "1"}[testCase.color],
				"TERM":                           "xterm-256color",
				helperEnvironment:                "1",
				"YCY_CONFIG_CM_REMOVE_PTY_START": "1",
				"YCY_CONFIG_CM_REMOVE_RELEASE":   releasePath,
			})
			output := runCMRemoveExtendedPTYProcess(t, command, testCase.width, testCase.height, releasePath)
			assertCMRemoveExtendedPTYOutput(t, output, testCase.color, testCase.width >= 70)
		})
	}
}

func runCMRemoveExtendedPTYHelper(t *testing.T) {
	t.Helper()
	if os.Getenv("YCY_CONFIG_CM_REMOVE_PTY_START") == "1" {
		var start [2]byte
		if _, err := io.ReadFull(os.Stdin, start[:]); err != nil {
			t.Fatalf("wait for PTY sizing: %v", err)
		}
	}
	releasePath := os.Getenv("YCY_CONFIG_CM_REMOVE_RELEASE")
	if releasePath == "" {
		t.Fatal("missing CM remove release path")
	}
	color := os.Getenv("NO_COLOR") == ""
	experience := terminalexperience.NewExperience(terminalexperience.ExperienceOptions{
		Capabilities: terminalexperience.Capabilities{
			Interaction: terminalexperience.RichInteractive,
			Stdin:       terminalexperience.StreamCapability{Terminal: true},
			Stdout:      terminalexperience.StreamCapability{Terminal: true, Color: color},
			Stderr:      terminalexperience.StreamCapability{Terminal: true, Color: color},
		},
		Input:       os.Stdin,
		Output:      os.Stdout,
		Diagnostics: os.Stderr,
	})
	err := runRemove(&Options{
		Context:  context.Background(),
		Profile:  "work",
		Terminal: experience,
		Store: func() (Reader, RemoveWriter, error) {
			return cmRemoveReaderFunc(func() (appconfig.CMProfileList, error) {
					return appconfig.CMProfileList{DefaultProfile: "work", Profiles: []appconfig.CMProfile{
						{Name: "work", BaseURL: "https://work.example/v1", Model: "must-not-escape-model"},
						{Name: "keep", BaseURL: "https://keep.example/v1", Model: "also-secret-model"},
					}}, nil
				}), cmRemoveWriterFunc(func(name string) (bool, error) {
					if name != "work" {
						return false, fmt.Errorf("unexpected profile %q", name)
					}
					deadline := time.Now().Add(5 * time.Second)
					for {
						if _, statErr := os.Stat(releasePath); statErr == nil {
							_, _ = fmt.Fprintln(os.Stderr, "CM_REMOVE_WRITE_OK")
							return true, nil
						}
						if time.Now().After(deadline) {
							return false, errors.New("CM remove PTY release timed out")
						}
						time.Sleep(5 * time.Millisecond)
					}
				}), nil
		},
	})
	if err != nil {
		t.Fatalf("runRemove() error = %v", err)
	}
}

func runCMRemoveExtendedPTYProcess(t *testing.T, command *exec.Cmd, width, height uint16, releasePath string) string {
	t.Helper()
	process, err := terminaltest.StartPTY(command)
	if errors.Is(err, terminaltest.ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY helper: %v", err)
	}
	defer process.Close()
	if err := process.Resize(width, height); err != nil {
		t.Fatalf("resize PTY to %dx%d: %v", width, height, err)
	}
	var output lockedCMRemovePTYBuffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	if _, err := process.Terminal().Write([]byte("go\n")); err != nil {
		t.Fatalf("release PTY helper after sizing: %v", err)
	}
	if width >= 70 {
		waitForCMRemovePTYText(t, &output, `Remove CM profile "work"?`)
	} else {
		waitForCMRemovePTYText(t, &output, "confirmation")
	}
	if _, err := process.Terminal().Write([]byte("y\r")); err != nil {
		t.Fatalf("confirm CM remove: %v", err)
	}
	if err := os.WriteFile(releasePath, []byte("ok"), 0o600); err != nil {
		t.Fatalf("release CM remove writer: %v", err)
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("wait PTY helper: %v\n%s", err, output.String())
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close PTY helper: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out reading PTY output: %q", output.String())
	}
	return output.String()
}

func assertCMRemoveExtendedPTYOutput(t *testing.T, output string, color, wide bool) {
	t.Helper()
	visible := strings.ReplaceAll(output, "\r\n", "\n")
	enter := strings.Index(visible, "\x1b[?1049h")
	leave := strings.LastIndex(visible, "\x1b[?1049l")
	if strings.Count(visible, "\x1b[?1049h") != 1 || strings.Count(visible, "\x1b[?1049l") != 1 || enter < 0 || leave < enter || !strings.Contains(visible, "\x1b[?25h") {
		t.Fatalf("CM remove Rich PTY did not restore primary screen: %q", output)
	}
	live := cmRemoveExtendedPTYText(visible[enter:leave])
	for _, expected := range []string{"YCY / config cm remove", "Validate CM profile", "Remove CM profile", "work", "STATE", "PHASE", "DETAIL", "CM_REMOVE_WRITE_OK"} {
		if !strings.Contains(live, expected) {
			t.Fatalf("CM remove Rich PTY live Console missing %q: %q", expected, output)
		}
	}
	if wide {
		if !strings.Contains(live, "commit message profile removal") || !strings.Contains(live, "commit message configuration") {
			t.Fatalf("wide CM remove Rich PTY omitted descriptor context: %q", output)
		}
	} else if !strings.Contains(live, "commit") {
		t.Fatalf("compact CM remove Rich PTY omitted bounded target context: %q", output)
	}
	if wide {
		for _, expected := range []string{"Remove CM profile \"work\"?", "Removing the default selects the first remaining stored profile"} {
			if !strings.Contains(visible, expected) {
				t.Fatalf("CM remove Rich PTY missing %q: %q", expected, output)
			}
		}
	} else if !strings.Contains(live, "confirmation") {
		t.Fatalf("compact CM remove Rich PTY omitted confirmation context: %q", output)
	}
	if strings.Contains(output, "must-not-escape-model") || strings.Contains(output, "also-secret-model") {
		t.Fatalf("CM remove Rich PTY leaked profile credential: %q", output)
	}
	postLive := visible[leave:]
	resultStart := strings.LastIndex(postLive, "Profile work removed")
	if resultStart < 0 {
		t.Fatalf("CM remove Rich PTY result missing: %q", output)
	}
	transcript := cmRemoveExtendedPTYText(postLive[:resultStart])
	for _, expected := range []string{"Validate CM profile (completed)", "Role: Current default", "Remove CM profile \"work\": confirmed", "Remove CM profile (completed)", "succeeded"} {
		if !strings.Contains(transcript, expected) {
			t.Fatalf("CM remove Rich PTY Transcript missing %q: %q", expected, output)
		}
	}
	if !color {
		for _, prefix := range []string{"\x1b[38;", "\x1b[3m", "\x1b[9m"} {
			if strings.Contains(output, prefix) {
				t.Fatalf("NO_COLOR CM remove Rich PTY output contains %q: %q", prefix, output)
			}
		}
	}
}

func cmRemoveExtendedPTYText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	return strings.Join(strings.Fields(terminaltest.StripANSI(value)), " ")
}

func cmRemoveExtendedPTYEnvironment(overrides map[string]string) []string {
	environment := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		key, _, found := strings.Cut(entry, "=")
		if found {
			if _, replaced := overrides[key]; !replaced {
				environment = append(environment, entry)
			}
		}
	}
	for key, value := range overrides {
		environment = append(environment, key+"="+value)
	}
	return environment
}
