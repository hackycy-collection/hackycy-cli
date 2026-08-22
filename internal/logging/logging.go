package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
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

var (
	assignmentSecret = regexp.MustCompile(`(?i)\b(authorization|cookie|password|secret|token)\s*[:=]\s*[^\s,;]+`)
	bearerSecret     = regexp.MustCompile(`(?i)\bbearer\s+[^\s,;]+`)
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
}

// Runtime owns filtering, formatting, redaction, and output for scoped loggers.
type Runtime struct {
	mu     sync.RWMutex
	level  Level
	writer io.Writer
	now    func() time.Time
	color  bool
}

// NewRuntime creates a runtime at the conventional info level.
func NewRuntime(options Options) *Runtime {
	if options.Writer == nil {
		options.Writer = os.Stderr
	}
	if options.Now == nil {
		options.Now = time.Now
	}
	return &Runtime{level: Info, writer: options.Writer, now: options.Now, color: options.Color}
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

	logger.runtime.mu.RLock()
	defer logger.runtime.mu.RUnlock()
	if level < logger.runtime.level {
		return
	}

	context := make(map[string]any, len(logger.context)+len(fields))
	for key, value := range logger.context {
		context[key] = redactValue(key, value)
	}
	for key, value := range fields {
		context[key] = redactValue(key, value)
	}

	label := fmt.Sprintf("%-5s", level.String())
	if logger.runtime.color {
		label = colorize(level, label)
	}
	line := fmt.Sprintf("%s %s", logger.runtime.now().UTC().Format(time.RFC3339), label)
	if logger.scope != "" {
		line += " " + logger.scope
	}
	line += " " + Redact(message)
	if len(context) > 0 {
		line += " " + marshalContext(context)
	}
	_, _ = fmt.Fprintln(logger.runtime.writer, line)
}

// Redact removes common credential-shaped values from diagnostics and messages.
func Redact(value string) string {
	value = assignmentSecret.ReplaceAllStringFunc(value, func(match string) string {
		separator := "="
		if strings.Contains(match, ":") && !strings.Contains(match, "=") {
			separator = ":"
		}
		key := strings.TrimSpace(strings.SplitN(match, separator, 2)[0])
		return key + separator + "[REDACTED]"
	})
	value = bearerSecret.ReplaceAllString(value, "Bearer [REDACTED]")
	return strings.ReplaceAll(value, "\n", `\n`)
}

func redactValue(key string, value any) any {
	lowerKey := strings.ToLower(key)
	if strings.Contains(lowerKey, "authorization") || strings.Contains(lowerKey, "cookie") || strings.Contains(lowerKey, "password") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "token") {
		return "[REDACTED]"
	}
	if text, ok := value.(string); ok {
		return Redact(text)
	}
	return value
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

func colorize(level Level, value string) string {
	code := "32"
	switch level {
	case Debug:
		code = "36"
	case Warn:
		code = "33"
	case Error:
		code = "31"
	}
	return "\x1b[" + code + "m" + value + "\x1b[0m"
}
