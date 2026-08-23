package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
	"github.com/hackycy/hackycy-cli/internal/cliapp"
	cmcommand "github.com/hackycy/hackycy-cli/internal/commands/git/cm"
)

func newGitCMHandler(input io.Reader, output io.Writer) cliapp.GitCMHandler {
	return func(ctx context.Context, request cmcommand.Input) (cmcommand.Result, error) {
		store, err := appconfig.New(appconfig.Dependencies{})
		if err != nil {
			return cmcommand.Result{}, err
		}
		prompter := newTerminalGitCMPrompter(input, output)
		module, err := cmcommand.New(cmcommand.Dependencies{
			Git:       newOSCMGitRunner(),
			Files:     osCMSnapshotFileSystem{},
			Prompter:  prompter,
			Committer: prompter,
			Resolver:  store,
			Transport: http.DefaultClient,
		})
		if err != nil {
			return cmcommand.Result{}, err
		}
		result, err := module.Run(ctx, request)
		presenter := terminalGitCMPresenter{output: output}
		if err != nil {
			if result.Generated == nil && result.Profile != (cmcommand.ProfileDiagnostic{}) {
				presenter.Failure(result)
			}
			return result, err
		}
		if result.Generated != nil && !result.PromptedCommit {
			presenter.Generated(result)
		}
		presenter.Outcome(result)
		return result, nil
	}
}

type terminalGitCMPrompter struct {
	input  *bufio.Reader
	output io.Writer
}

func newTerminalGitCMPrompter(input io.Reader, output io.Writer) *terminalGitCMPrompter {
	return &terminalGitCMPrompter{input: bufio.NewReader(input), output: output}
}

func (prompter *terminalGitCMPrompter) SelectFiles(prompt cmcommand.StagePrompt) ([]string, bool) {
	_, _ = fmt.Fprintln(prompter.output, prompt.Message)
	for index, option := range prompt.Options {
		_, _ = fmt.Fprintf(prompter.output, "%d) %s\n", index+1, option.Label)
	}
	for {
		_, _ = fmt.Fprint(prompter.output, "> ")
		value, eof := prompter.readLine()
		if isGitCMCancellation(value) || (value == "" && eof) {
			return nil, true
		}
		if value == "" || strings.EqualFold(value, "all") {
			return append([]string(nil), prompt.InitialValues...), false
		}
		if strings.EqualFold(value, "none") {
			return []string{}, false
		}
		selected, valid := selectGitCMOptions(value, prompt.Options)
		if valid {
			return selected, false
		}
		if eof {
			return nil, true
		}
		_, _ = fmt.Fprintln(prompter.output, "Invalid selection")
	}
}

func (prompter *terminalGitCMPrompter) ConfirmCommit(prompt cmcommand.CommitPrompt) (bool, bool) {
	writeGitCMGeneratedMessage(prompter.output, prompt.Generated, prompt.Profile)
	for {
		_, _ = fmt.Fprintf(prompter.output, "%s [Y/n]: ", prompt.Message)
		value, eof := prompter.readLine()
		if eof && value == "" {
			return false, true
		}
		switch strings.ToLower(value) {
		case "", "y", "yes":
			return true, false
		case "n", "no":
			return false, false
		case "q", "quit", "cancel":
			return false, true
		default:
			_, _ = fmt.Fprintln(prompter.output, "Invalid confirmation")
			if eof {
				return false, true
			}
		}
	}
}

func (prompter *terminalGitCMPrompter) readLine() (string, bool) {
	line, err := prompter.input.ReadString('\n')
	return strings.TrimSpace(line), err != nil
}

func isGitCMCancellation(value string) bool {
	return strings.EqualFold(value, "q") || strings.EqualFold(value, "quit") || strings.EqualFold(value, "cancel")
}

func selectGitCMOptions(value string, options []cmcommand.StageOption) ([]string, bool) {
	parts := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t'
	})
	if len(parts) == 0 {
		return nil, false
	}
	selected := make([]string, 0, len(parts))
	seen := make(map[int]bool, len(parts))
	for _, part := range parts {
		index, err := strconv.Atoi(part)
		if err != nil || index < 1 || index > len(options) || seen[index] {
			return nil, false
		}
		seen[index] = true
		selected = append(selected, options[index-1].Value)
	}
	return selected, true
}

type terminalGitCMPresenter struct {
	output io.Writer
}

func (presenter terminalGitCMPresenter) Generated(result cmcommand.Result) {
	if result.Generated == nil {
		return
	}
	writeGitCMGeneratedMessage(presenter.output, *result.Generated, result.Profile)
}

func (presenter terminalGitCMPresenter) Failure(result cmcommand.Result) {
	_, _ = fmt.Fprintf(presenter.output, "Provider: %s\nBase URL: %s\nModel: %s\n", result.Profile.Name, result.Profile.BaseURL, result.Profile.Model)
}

func (presenter terminalGitCMPresenter) Outcome(result cmcommand.Result) {
	if result.NoChanges {
		if result.NoChangeScope == cmcommand.ScopeStaged {
			_, _ = fmt.Fprintln(presenter.output, "No staged changes.")
			return
		}
		_, _ = fmt.Fprintln(presenter.output, "No uncommitted changes.")
		return
	}
	if result.NothingSelected {
		_, _ = fmt.Fprintln(presenter.output, "Nothing selected.")
		return
	}
	if result.Cancelled {
		_, _ = fmt.Fprintln(presenter.output, "Cancelled")
		return
	}
	if result.Pushed {
		_, _ = fmt.Fprintln(presenter.output, "Commit created and pushed")
		return
	}
	if result.Committed {
		_, _ = fmt.Fprintln(presenter.output, "Commit created")
	}
}

func writeGitCMGeneratedMessage(output io.Writer, generated cmcommand.GeneratedMessage, profile cmcommand.ProfileDiagnostic) {
	_, _ = fmt.Fprintln(output, generated.Message)
	_, _ = fmt.Fprintln(output)
	_, _ = fmt.Fprintf(output, "Profile: %s (%s)\n", profile.Name, profile.Model)
	_, _ = fmt.Fprintln(output, formatGitCMTokenUsage(generated.Usage))
	coverage := generated.Evidence
	_, _ = fmt.Fprintf(output, "Local evidence estimate: ~%s serialized prompt tokens / %d of %d clusters / %d of %d facts\n", formatGitCMCount(float64(coverage.EstimatedLocalPromptTokens)), coverage.RepresentedClusters, coverage.TotalClusters, coverage.IncludedFacts, coverage.IncludedFacts+coverage.OmittedFacts)
	if coverage.ContentCompacted {
		suffix := "s"
		if coverage.TotalClusters == 1 {
			suffix = ""
		}
		_, _ = fmt.Fprintf(output, "Commit scope: %d cluster%s represented with compacted semantic evidence. This does not affect which files are committed.\n", coverage.TotalClusters, suffix)
	}
}

func formatGitCMTokenUsage(usage *cmcommand.TokenUsage) string {
	if usage == nil {
		return "Provider tokens: unavailable"
	}
	return "Provider tokens: " + formatGitCMTokenValue(usage.PromptTokens) + " prompt / " + formatGitCMTokenValue(usage.CompletionTokens) + " completion / " + formatGitCMTokenValue(usage.TotalTokens) + " total"
}

func formatGitCMTokenValue(value *float64) string {
	if value == nil {
		return "unknown"
	}
	return formatGitCMCount(*value)
}

func formatGitCMCount(value float64) string {
	formatted := strconv.FormatFloat(value, 'f', -1, 64)
	decimal := strings.IndexByte(formatted, '.')
	integer := formatted
	fraction := ""
	if decimal >= 0 {
		integer = formatted[:decimal]
		fraction = formatted[decimal:]
	}
	start := 0
	if strings.HasPrefix(integer, "-") {
		start = 1
	}
	for index := len(integer) - 3; index > start; index -= 3 {
		integer = integer[:index] + "," + integer[index:]
	}
	return integer + fraction
}
