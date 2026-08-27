package terminal

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

var (
	// ErrAutomationInteraction reports an interaction request in an Automation Session.
	ErrAutomationInteraction = errors.New("interaction requires an interactive terminal")
	// ErrInteractionCancelled reports that an interactive answer was not supplied.
	ErrInteractionCancelled = errors.New("terminal interaction cancelled")
)

// InteractionOptions supplies explicit I/O and the immutable session to one handler.
type InteractionOptions struct {
	Session     Session
	Input       io.Reader
	Diagnostics io.Writer
}

// InteractionHandler translates terminal-owned semantic requests into one session's UI.
type InteractionHandler struct {
	session     Session
	input       io.Reader
	lines       *bufio.Reader
	diagnostics io.Writer
}

// NewInteractionHandler creates an interaction handler without consulting a process terminal.
func NewInteractionHandler(options InteractionOptions) *InteractionHandler {
	input := options.Input
	if input == nil {
		input = strings.NewReader("")
	}
	diagnostics := options.Diagnostics
	if diagnostics == nil {
		diagnostics = io.Discard
	}
	return &InteractionHandler{
		session:     options.Session,
		input:       input,
		lines:       bufio.NewReader(input),
		diagnostics: diagnostics,
	}
}

func (handler *InteractionHandler) withDiagnostics(diagnostics io.Writer) *InteractionHandler {
	if diagnostics == nil {
		diagnostics = io.Discard
	}
	return &InteractionHandler{
		session:     handler.session,
		input:       handler.input,
		lines:       handler.lines,
		diagnostics: diagnostics,
	}
}

// Ask obtains one answer according to the session selected before interaction begins.
func (handler *InteractionHandler) Ask(ctx context.Context, request InteractionRequest) (InteractionAnswer, error) {
	if err := ctx.Err(); err != nil {
		return InteractionAnswer{}, err
	}
	if handler.session.Kind == Automation {
		return InteractionAnswer{}, ErrAutomationInteraction
	}
	if handler.session.Kind == RichInteractive {
		return handler.askHuh(ctx, request)
	}
	return handler.askPlain(ctx, request)
}

func (handler *InteractionHandler) askHuh(ctx context.Context, request InteractionRequest) (InteractionAnswer, error) {
	form, answer, err := handler.huhForm(request)
	if err != nil {
		return InteractionAnswer{}, err
	}
	if err := form.RunWithContext(ctx); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return InteractionAnswer{}, contextErr
		}
		if errors.Is(err, huh.ErrUserAborted) {
			return InteractionAnswer{}, ErrInteractionCancelled
		}
		return InteractionAnswer{}, err
	}
	return answer(), nil
}

func (handler *InteractionHandler) huhForm(request InteractionRequest) (*huh.Form, func() InteractionAnswer, error) {
	newForm := func(field huh.Field) *huh.Form {
		return huh.NewForm(huh.NewGroup(field)).WithInput(handler.input).WithOutput(handler.diagnostics)
	}

	switch request.Kind {
	case InteractionText, InteractionSecret:
		value := ""
		if request.HasDefault {
			value = request.Default.Value
		}
		field := huh.NewInput().
			Title(request.Message).
			Description(request.Description).
			Placeholder(request.Placeholder).
			Value(&value).
			Validate(func(value string) error {
				return validateAnswer(request, InteractionAnswer{Value: value})
			})
		if request.Kind == InteractionSecret {
			field.EchoMode(huh.EchoModePassword)
		}
		return newForm(field), func() InteractionAnswer { return InteractionAnswer{Value: value} }, nil
	case InteractionSelect:
		value := ""
		if request.HasDefault {
			value = request.Default.Value
		}
		field := huh.NewSelect[string]().
			Title(request.Message).
			Description(request.Description).
			Options(huhOptions(request.Options)...).
			Value(&value).
			Validate(func(value string) error {
				return validateAnswer(request, InteractionAnswer{Value: value})
			})
		return newForm(field), func() InteractionAnswer { return InteractionAnswer{Value: value} }, nil
	case InteractionMultiSelect:
		values := []string(nil)
		if request.HasDefault {
			values = append(values, request.Default.Values...)
		}
		field := huh.NewMultiSelect[string]().
			Title(request.Message).
			Description(request.Description).
			Options(huhOptions(request.Options)...).
			Value(&values).
			Validate(func(values []string) error {
				return validateAnswer(request, InteractionAnswer{Values: values})
			})
		return newForm(field), func() InteractionAnswer { return InteractionAnswer{Values: append([]string(nil), values...)} }, nil
	case InteractionConfirm:
		confirmed := false
		if request.HasDefault {
			confirmed = request.Default.Confirmed
		}
		field := huh.NewConfirm().
			Title(request.Message).
			Description(request.Description).
			Value(&confirmed).
			Validate(func(confirmed bool) error {
				return validateAnswer(request, InteractionAnswer{Confirmed: confirmed})
			})
		return newForm(field), func() InteractionAnswer { return InteractionAnswer{Confirmed: confirmed} }, nil
	default:
		return nil, nil, fmt.Errorf("unsupported interaction kind %d", request.Kind)
	}
}

func huhOptions(options []InteractionOption) []huh.Option[string] {
	result := make([]huh.Option[string], 0, len(options))
	for _, option := range options {
		result = append(result, huh.NewOption(option.Label, option.Value))
	}
	return result
}

func (handler *InteractionHandler) askPlain(ctx context.Context, request InteractionRequest) (InteractionAnswer, error) {
	for {
		if err := ctx.Err(); err != nil {
			return InteractionAnswer{}, err
		}
		answer, err := handler.readPlainAnswer(ctx, request)
		if err != nil {
			return InteractionAnswer{}, err
		}
		if err := validateAnswer(request, answer); err != nil {
			_, _ = fmt.Fprintln(handler.diagnostics, err)
			continue
		}
		return answer, nil
	}
}

func (handler *InteractionHandler) readPlainAnswer(ctx context.Context, request InteractionRequest) (InteractionAnswer, error) {
	if request.Kind == InteractionSecret {
		return handler.readPlainSecret(ctx, request)
	}
	for {
		handler.writePlainPrompt(request)
		line, err := handler.readPlainLine(ctx)
		if err != nil {
			return InteractionAnswer{}, err
		}
		if isCancelValue(line, request.CancelValues) {
			return InteractionAnswer{}, ErrInteractionCancelled
		}
		if line == "" && request.HasDefault {
			return request.Default, nil
		}
		if request.ParsePlain != nil {
			answer, err := request.ParsePlain(line)
			if err == nil {
				return answer, nil
			}
			_, _ = fmt.Fprintln(handler.diagnostics, err)
			continue
		}

		switch request.Kind {
		case InteractionText:
			return InteractionAnswer{Value: line}, nil
		case InteractionSelect:
			value, err := selectValue(line, request.Options)
			if err == nil {
				return InteractionAnswer{Value: value}, nil
			}
			_, _ = fmt.Fprintln(handler.diagnostics, err)
		case InteractionMultiSelect:
			values, err := selectValues(line, request.Options)
			if err == nil {
				return InteractionAnswer{Values: values}, nil
			}
			_, _ = fmt.Fprintln(handler.diagnostics, err)
		case InteractionConfirm:
			confirmed, err := parseConfirmation(line)
			if err == nil {
				return InteractionAnswer{Confirmed: confirmed}, nil
			}
			_, _ = fmt.Fprintln(handler.diagnostics, err)
		default:
			return InteractionAnswer{}, fmt.Errorf("unsupported interaction kind %d", request.Kind)
		}
	}
}

func isCancelValue(value string, cancelValues []string) bool {
	value = strings.TrimSpace(value)
	for _, cancelValue := range cancelValues {
		if strings.EqualFold(value, strings.TrimSpace(cancelValue)) {
			return true
		}
	}
	return false
}

func (handler *InteractionHandler) writePlainPrompt(request InteractionRequest) {
	if request.Description != "" {
		_, _ = fmt.Fprintln(handler.diagnostics, request.Description)
	}
	if request.PlainLead != "" {
		_, _ = fmt.Fprintln(handler.diagnostics, request.PlainLead)
	}
	if request.Kind == InteractionSelect || request.Kind == InteractionMultiSelect {
		for index, option := range request.Options {
			label := option.Label
			if option.Description != "" {
				label += " - " + option.Description
			}
			_, _ = fmt.Fprintf(handler.diagnostics, "%d) %s\n", index+1, label)
		}
	}
	prompt := request.Message
	if request.Kind == InteractionConfirm {
		prompt += " [y/n]"
	}
	if request.Placeholder != "" && request.Kind == InteractionText {
		prompt += " (" + request.Placeholder + ")"
	}
	if request.PlainPrompt != "" {
		_, _ = fmt.Fprint(handler.diagnostics, request.PlainPrompt)
		return
	}
	_, _ = fmt.Fprint(handler.diagnostics, prompt+": ")
}

func (handler *InteractionHandler) readPlainSecret(ctx context.Context, request InteractionRequest) (InteractionAnswer, error) {
	file, ok := handler.input.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return InteractionAnswer{}, ErrAutomationInteraction
	}
	handler.writePlainPrompt(request)
	value, err := term.ReadPassword(int(file.Fd()))
	_, _ = fmt.Fprintln(handler.diagnostics)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return InteractionAnswer{}, contextErr
		}
		return InteractionAnswer{}, ErrInteractionCancelled
	}
	return InteractionAnswer{Value: string(value)}, nil
}

func (handler *InteractionHandler) readPlainLine(ctx context.Context) (string, error) {
	line, err := handler.lines.ReadString('\n')
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if err != nil && line == "" {
		if contextErr := ctx.Err(); contextErr != nil {
			return "", contextErr
		}
		return "", ErrInteractionCancelled
	}
	return line, nil
}

func validateAnswer(request InteractionRequest, answer InteractionAnswer) error {
	if request.Validate == nil {
		return nil
	}
	return request.Validate(answer)
}

func selectValue(value string, options []InteractionOption) (string, error) {
	value = strings.TrimSpace(value)
	for index, option := range options {
		if value == option.Value || value == option.Label || value == strconv.Itoa(index+1) {
			return option.Value, nil
		}
	}
	return "", fmt.Errorf("invalid selection")
}

func selectValues(value string, options []InteractionOption) ([]string, error) {
	parts := strings.FieldsFunc(value, func(character rune) bool {
		return character == ',' || character == ' ' || character == '\t'
	})
	if len(parts) == 0 {
		return nil, fmt.Errorf("at least one selection is required")
	}
	values := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		selected, err := selectValue(part, options)
		if err != nil {
			return nil, err
		}
		if _, duplicate := seen[selected]; duplicate {
			continue
		}
		seen[selected] = struct{}{}
		values = append(values, selected)
	}
	return values, nil
}

func parseConfirmation(value string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("please answer yes or no")
	}
}
