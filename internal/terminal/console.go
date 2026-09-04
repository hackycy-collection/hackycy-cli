package terminal

import (
	"errors"
	"fmt"
	"strings"
)

const (
	maxConsoleMetadata = 4
	maxConsoleField    = 160
)

// ErrInvalidConsoleDescriptor reports a descriptor that cannot safely be
// rendered as the bounded Ops Console context.
var ErrInvalidConsoleDescriptor = errors.New("terminal console descriptor is invalid")

func defaultConsoleDescriptor() ConsoleDescriptor {
	return ConsoleDescriptor{
		Command: "YCY",
		Target:  "terminal session",
		Status:  "READY",
		Metadata: []ConsoleMetadata{{
			Label: "mode",
			Value: "interactive",
		}},
	}
}

func normalizeConsoleDescriptor(descriptor ConsoleDescriptor) (ConsoleDescriptor, error) {
	command, err := normalizeConsoleField(descriptor.Command, "command", true)
	if err != nil {
		return ConsoleDescriptor{}, err
	}
	target, err := normalizeConsoleField(descriptor.Target, "target", false)
	if err != nil {
		return ConsoleDescriptor{}, err
	}
	status, err := normalizeConsoleField(descriptor.Status, "status", false)
	if err != nil {
		return ConsoleDescriptor{}, err
	}
	if status == "" {
		status = "READY"
	}
	if len(descriptor.Metadata) > maxConsoleMetadata {
		return ConsoleDescriptor{}, consoleDescriptorError("metadata has %d fields; at most %d are allowed", len(descriptor.Metadata), maxConsoleMetadata)
	}

	metadata := make([]ConsoleMetadata, 0, len(descriptor.Metadata))
	for _, field := range descriptor.Metadata {
		label, err := normalizeConsoleField(field.Label, "metadata label", true)
		if err != nil {
			return ConsoleDescriptor{}, err
		}
		value, err := normalizeConsoleField(field.Value, "metadata value", true)
		if err != nil {
			return ConsoleDescriptor{}, err
		}
		metadata = append(metadata, ConsoleMetadata{Label: label, Value: value})
	}

	return ConsoleDescriptor{
		Command:  command,
		Target:   target,
		Status:   status,
		Metadata: metadata,
	}, nil
}

func normalizeConsoleField(value, name string, required bool) (string, error) {
	value = strings.Join(strings.Fields(stripTerminalControl(strings.ToValidUTF8(value, "�"))), " ")
	if required && value == "" {
		return "", consoleDescriptorError("%s is required", name)
	}
	if len(value) > maxConsoleField {
		return "", consoleDescriptorError("%s exceeds %d bytes", name, maxConsoleField)
	}
	return value, nil
}

func consoleDescriptorError(format string, arguments ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrInvalidConsoleDescriptor}, arguments...)...)
}
