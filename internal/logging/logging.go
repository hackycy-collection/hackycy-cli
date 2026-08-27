package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Level controls which structured messages a Runtime writes.
type Level uint8

const (
	Debug Level = iota
	Info
	Warn
	Error
)

// ParseLevel accepts the global CLI spelling.
func ParseLevel(value string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return Debug, nil
	case "info", "":
		return Info, nil
	case "warn":
		return Warn, nil
	case "error":
		return Error, nil
	default:
		return Info, fmt.Errorf("invalid log level %q (expected debug, info, warn, or error)", value)
	}
}

func (level Level) String() string {
	switch level {
	case Debug:
		return "DEBUG"
	case Info:
		return "INFO"
	case Warn:
		return "WARN"
	case Error:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Options makes logging dependencies explicit at the composition root.
type Options struct {
	Writer io.Writer
	Now    func() time.Time
	Color  bool
	Format RecordFormat
}

// Runtime owns filtering, formatting, redaction, and output for scoped loggers.
type Runtime struct {
	mu     sync.RWMutex
	level  Level
	writer io.Writer
	now    func() time.Time
	color  bool
	format RecordFormat
}

// NewRuntime creates a runtime at the conventional info level.
func NewRuntime(options Options) *Runtime {
	if options.Writer == nil {
		options.Writer = os.Stderr
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Runtime{
		level:  Info,
		writer: options.Writer,
		now:    options.Now,
		color:  options.Color,
		format: normalizeRecordFormat(options.Format),
	}
}

func (runtime *Runtime) SetLevel(level Level) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.level = level
}

func (runtime *Runtime) Level() Level {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.level
}

// SetFormat changes the schema used for subsequent Diagnostic Records.
func (runtime *Runtime) SetFormat(format RecordFormat) {
	runtime.mu.Lock()
	defer runtime.mu.Unlock()
	runtime.format = normalizeRecordFormat(format)
}

// Format returns the current Diagnostic Record schema.
func (runtime *Runtime) Format() RecordFormat {
	runtime.mu.RLock()
	defer runtime.mu.RUnlock()
	return runtime.format
}

// Logger creates a logger whose scope is part of every message.
func (runtime *Runtime) Logger(scope string) Logger {
	return Logger{runtime: runtime, scope: scope}
}

// Logger carries only a scope and redacted structured context.
type Logger struct {
	runtime *Runtime
	scope   string
	context map[string]any
}

// With returns a child logger without mutating the parent context.
func (logger Logger) With(fields map[string]any) Logger {
	context := make(map[string]any, len(logger.context)+len(fields))
	for key, value := range logger.context {
		context[key] = value
	}
	for key, value := range fields {
		context[key] = value
	}
	logger.context = context
	return logger
}

func (logger Logger) Debug(message string, fields map[string]any) { logger.Log(Debug, message, fields) }
func (logger Logger) Info(message string, fields map[string]any)  { logger.Log(Info, message, fields) }
func (logger Logger) Warn(message string, fields map[string]any)  { logger.Log(Warn, message, fields) }
func (logger Logger) Error(message string, fields map[string]any) { logger.Log(Error, message, fields) }

// Log formats one redacted structured line on stderr-selected output.
func (logger Logger) Log(level Level, message string, fields map[string]any) {
	if logger.runtime == nil {
		return
	}

	logger.runtime.mu.Lock()
	defer logger.runtime.mu.Unlock()
	if level < logger.runtime.level {
		return
	}

	context := make(map[string]any, len(logger.context)+len(fields))
	for key, value := range logger.context {
		context[key] = value
	}
	for key, value := range fields {
		context[key] = value
	}

	record := diagnosticRecord{
		timestamp: diagnosticTimestamp(logger.runtime.now()),
		level:     level,
		scope:     logger.scope,
		message:   Redact(message),
		context:   redactContext(context),
	}
	_, _ = io.WriteString(logger.runtime.writer, renderRecord(record, logger.runtime.format, logger.runtime.color))
}

type diagnosticRecord struct {
	timestamp string
	level     Level
	scope     string
	message   string
	context   map[string]any
}

func diagnosticTimestamp(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05.000Z")
}

func renderRecord(record diagnosticRecord, format RecordFormat, color bool) string {
	if format == JSONFormat {
		return renderJSONRecord(record)
	}
	return renderTextRecord(record, color)
}

func renderTextRecord(record diagnosticRecord, color bool) string {
	timestamp := record.timestamp
	label := fmt.Sprintf("%-5s", record.level.String())
	scope := ""
	if record.scope != "" {
		scope = "[" + record.scope + "]"
	}
	if color {
		timestamp = styleTimestamp(timestamp)
		label = styleLevel(record.level, label)
		if scope != "" {
			scope = styleScope(scope)
		}
	}
	parts := []string{timestamp, label}
	if scope != "" {
		parts = append(parts, scope)
	}
	parts = append(parts, record.message)
	if len(record.context) > 0 {
		parts = append(parts, marshalContext(record.context))
	}
	return strings.Join(parts, " ") + "\n"
}

func renderJSONRecord(record diagnosticRecord) string {
	value := struct {
		Timestamp string          `json:"timestamp"`
		Level     string          `json:"level"`
		Scope     string          `json:"scope,omitempty"`
		Message   string          `json:"message"`
		Context   json.RawMessage `json:"context,omitempty"`
	}{
		Timestamp: record.timestamp,
		Level:     strings.ToLower(record.level.String()),
		Scope:     record.scope,
		Message:   record.message,
	}
	if len(record.context) > 0 {
		value.Context = json.RawMessage(marshalContext(record.context))
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return `{"timestamp":"` + record.timestamp + `","level":"error","message":"diagnostic encoding failed"}` + "\n"
	}
	return string(encoded) + "\n"
}

func marshalContext(context map[string]any) string {
	keys := make([]string, 0, len(context))
	for key := range context {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]any, len(context))
	for _, key := range keys {
		ordered[key] = context[key]
	}
	encoded, err := json.Marshal(ordered)
	if err != nil {
		return `{"logging":"context encoding failed"}`
	}
	return string(encoded)
}

func styleTimestamp(value string) string {
	return "\x1b[2;90m" + value + "\x1b[0m"
}

func styleLevel(level Level, value string) string {
	code := "32"
	switch level {
	case Debug:
		code = "2;90"
	case Warn:
		code = "33"
	case Error:
		code = "1;31"
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}

func styleScope(value string) string {
	return "\x1b[36m" + value + "\x1b[0m"
}

func normalizeRecordFormat(format RecordFormat) RecordFormat {
	if format == JSONFormat {
		return JSONFormat
	}
	return TextFormat
}
