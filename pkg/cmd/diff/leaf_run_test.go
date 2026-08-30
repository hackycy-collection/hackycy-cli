package diff

import (
	"bytes"
	"context"
	"io"
	"strings"
	"sync"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/terminal"
)

func TestRunDiffClosesTheForegroundServerAfterPresentationCancellation(t *testing.T) {
	baseline := t.TempDir()
	target := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var output bytes.Buffer
	experience := terminal.NewExperience(terminal.ExperienceOptions{
		Capabilities: terminal.Capabilities{Interaction: terminal.Automation},
		Output:       &cancelAfterWrite{Writer: &output, Cancel: cancel},
	})

	err := runDiff(&Options{
		Context: ctx,
		Input: Input{
			BaselineDirectory: baseline,
			TargetDirectory:   target,
			Port:              0,
		},
		Terminal: experience,
		NetworkInterfaces: func() ([]NetworkInterface, error) {
			return nil, nil
		},
	})
	if err != nil || !strings.Contains(output.String(), "Directory diff: http://127.0.0.1:") {
		t.Fatalf("runDiff() = (%v, %q)", err, output.String())
	}
}

type cancelAfterWrite struct {
	io.Writer
	Cancel context.CancelFunc
	once   sync.Once
}

func (writer *cancelAfterWrite) Write(contents []byte) (int, error) {
	written, err := writer.Writer.Write(contents)
	writer.once.Do(writer.Cancel)
	return written, err
}
