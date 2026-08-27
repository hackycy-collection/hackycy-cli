package terminaltest

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"

	"golang.org/x/term"
)

func TestFactsAreFullyInjected(t *testing.T) {
	facts := Facts{
		Stdin:  StreamFacts{Terminal: true, Size: Size{Width: 120, Height: 40}},
		Stdout: StreamFacts{Terminal: true},
		Stderr: StreamFacts{Terminal: false},
		Environment: map[string]string{
			"CI":       "",
			"NO_COLOR": "1",
		},
	}

	if !facts.Stream(Stdin).Terminal || facts.Stream(Stdin).Size != (Size{Width: 120, Height: 40}) {
		t.Fatalf("stdin facts = %#v", facts.Stream(Stdin))
	}
	if facts.Stream(Stderr).Terminal || facts.Stream(Stream("unknown")).Terminal {
		t.Fatalf("stderr/unknown facts = %#v / %#v", facts.Stream(Stderr), facts.Stream(Stream("unknown")))
	}
	if value, ok := facts.LookupEnv("CI"); !ok || value != "" {
		t.Fatalf("CI lookup = (%q, %t), want set empty value", value, ok)
	}
	if _, ok := facts.LookupEnv("TERM"); ok {
		t.Fatal("TERM unexpectedly present")
	}
}

func TestRecordingRunCapturesSemanticOperationOrder(t *testing.T) {
	run := NewRecordingRun(Answer{Value: "selected"}, Answer{Cancelled: true})
	if answer := run.Ask("choose project"); answer.Value != "selected" || answer.Cancelled || answer.Err != nil {
		t.Fatalf("first answer = %#v", answer)
	}
	run.Present("project report")
	run.Track("scan")
	if answer := run.Ask("confirm"); !answer.Cancelled {
		t.Fatalf("second answer = %#v", answer)
	}
	run.Close()

	want := []Operation{
		{Kind: AskOperation, Value: "choose project"},
		{Kind: PresentOperation, Value: "project report"},
		{Kind: TrackOperation, Value: "scan"},
		{Kind: AskOperation, Value: "confirm"},
		{Kind: CloseOperation},
	}
	operations := run.Operations()
	if !reflect.DeepEqual(operations, want) {
		t.Fatalf("operations = %#v, want %#v", operations, want)
	}
	operations[0].Value = "mutated assertion"
	if run.Operations()[0].Value != "choose project" {
		t.Fatal("Operations returned mutable recorder state")
	}
}

func TestRedirectedStreamsSeparateOutputAndRejectTerminalControl(t *testing.T) {
	streams := NewRedirectedStreams("answer\n")
	input, err := bufio.NewReader(streams.Stdin).ReadString('\n')
	if err != nil || input != "answer\n" {
		t.Fatalf("redirected input = (%q, %v)", input, err)
	}
	_, _ = streams.Stdout.WriteString("result\n")
	_, _ = streams.Stderr.WriteString("diagnostic\n")
	if streams.Stdout.String() != "result\n" || streams.Stderr.String() != "diagnostic\n" {
		t.Fatalf("redirected output = %q / %q", streams.Stdout.String(), streams.Stderr.String())
	}
	if ContainsTerminalControl(streams.Stdout.Bytes()) || !ContainsTerminalControl([]byte("\x1b[2K")) || !ContainsTerminalControl([]byte{0x9b, 'm'}) {
		t.Fatal("terminal-control detection did not preserve the Automation assertion boundary")
	}
}

func TestControlledPTYRunsATerminalSubprocess(t *testing.T) {
	const helperEnvironment = "YCY_TERMINALTEST_PTY_HELPER"
	if os.Getenv(helperEnvironment) == "1" {
		fmt.Printf("tty=%t/%t/%t\n", term.IsTerminal(int(os.Stdin.Fd())), term.IsTerminal(int(os.Stdout.Fd())), term.IsTerminal(int(os.Stderr.Fd())))
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			t.Fatalf("read PTY input: %v", err)
		}
		fmt.Printf("input=%s", line)
		return
	}

	command := exec.Command(os.Args[0], "-test.run=^TestControlledPTYRunsATerminalSubprocess$")
	command.Env = append(os.Environ(), helperEnvironment+"=1")
	process, err := StartPTY(command)
	if errors.Is(err, ErrPTYUnsupported) {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("start PTY: %v", err)
	}
	defer process.Close()

	var output bytes.Buffer
	readDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(&output, process.Terminal())
		close(readDone)
	}()
	if _, err := io.WriteString(process.Terminal(), "reply\n"); err != nil {
		t.Fatalf("write PTY input: %v", err)
	}
	if err := process.Wait(); err != nil {
		t.Fatalf("wait PTY helper: %v", err)
	}
	if err := process.Close(); err != nil {
		t.Fatalf("close PTY: %v", err)
	}
	select {
	case <-readDone:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out reading PTY output")
	}
	if got := output.String(); !bytes.Contains([]byte(got), []byte("tty=true/true/true")) || !bytes.Contains([]byte(got), []byte("input=reply")) {
		t.Fatalf("PTY output = %q", got)
	}
}
