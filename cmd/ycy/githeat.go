package main

import (
	"fmt"
	"io"
	"time"

	heatcommand "github.com/hackycy/hackycy-cli/internal/commands/git/heat"
)

func newGitHeatModule(output io.Writer, color bool) (*heatcommand.Module, error) {
	return heatcommand.New(heatcommand.Dependencies{
		Git:       newOSHeatGitRunner(),
		Presenter: terminalGitHeatPresenter{output: output, color: color},
		Now:       time.Now,
	})
}

type terminalGitHeatPresenter struct {
	output io.Writer
	color  bool
}

func (presenter terminalGitHeatPresenter) Present(report heatcommand.Report, now time.Time) error {
	if !report.IsEmpty() {
		if _, err := fmt.Fprintln(presenter.output, "HACKYCY CLI"); err != nil {
			return err
		}
		if _, err := fmt.Fprintln(presenter.output); err != nil {
			return err
		}
	}
	_, err := io.WriteString(presenter.output, heatcommand.RenderReport(report, now, presenter.color))
	return err
}
