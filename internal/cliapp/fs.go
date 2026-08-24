package cliapp

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	fscommand "github.com/hackycy/hackycy-cli/internal/commands/fs"
	"github.com/spf13/cobra"
)

// FSHandler is the fixed typed handler for fs.
type FSHandler func(context.Context, fscommand.Input) (fscommand.Result, error)

func (app *App) registerFS(root *cobra.Command, configureLogging func() error) {
	port := "1204"
	address := "0.0.0.0"
	accounts := []string{}
	sessionDirectory := ""
	sessionIdleDays := ""
	chunkedUploads := false
	uploadChunkSize := ""
	var managementEnabled bool
	var safeHTML bool
	command := &cobra.Command{
		Use:   "fs [directory]",
		Short: "Browse a directory in a browser",
		Args:  cobra.MaximumNArgs(1),
		PreRunE: func(*cobra.Command, []string) error {
			return configureLogging()
		},
		RunE: func(command *cobra.Command, arguments []string) error {
			parsedPort, err := parseFSPort(port)
			if err != nil {
				return err
			}
			idleDays, err := parseFSIdleDays(fsValue(command, "session-idle-days", sessionIdleDays, app.environment("YCY_FS_SESSION_IDLE_DAYS"), "7"))
			if err != nil {
				return err
			}
			chunked, err := parseFSBool(fsValue(command, "chunked-upload", strconv.FormatBool(chunkedUploads), app.environment("YCY_FS_CHUNKED_UPLOAD"), "false"))
			if err != nil {
				return err
			}
			chunkSize, err := parseFSChunkSize(fsValue(command, "upload-chunk-size", uploadChunkSize, app.environment("YCY_FS_UPLOAD_CHUNK_SIZE_MIB"), "8"))
			if err != nil {
				return err
			}
			directory := "."
			if len(arguments) == 1 {
				directory = arguments[0]
			}
			sessions := fsValue(command, "session-dir", sessionDirectory, app.environment("YCY_FS_SESSION_DIR"), "")
			if len(accounts) > 0 && sessions == "" {
				sessions, err = fscommand.DefaultSessionDirectory(directory)
				if err != nil {
					return err
				}
			}
			_, err = app.fs(command.Context(), fscommand.Input{
				Directory:          directory,
				Port:               parsedPort,
				Address:            address,
				ManagementEnabled:  managementEnabled,
				SafeHTML:           safeHTML,
				Accounts:           append([]string(nil), accounts...),
				SessionDirectory:   sessions,
				SessionIdleTimeout: time.Duration(idleDays) * 24 * time.Hour,
				ChunkedUploads:     chunked,
				UploadChunkSize:    int64(chunkSize) * 1024 * 1024,
			})
			return err
		},
	}
	command.Flags().StringVarP(&port, "port", "p", port, "Port to serve on")
	command.Flags().StringVarP(&address, "address", "a", address, "Address to bind")
	command.Flags().BoolVarP(&managementEnabled, "manage", "m", false, "Enable filesystem management")
	command.Flags().BoolVar(&safeHTML, "safe-html", false, "Serve HTML and XHTML originals as sandboxed downloads")
	command.Flags().StringArrayVar(&accounts, "account", accounts, "Account in username:password form (repeatable)")
	command.Flags().StringVar(&sessionDirectory, "session-dir", sessionDirectory, "Directory for Go-owned login sessions")
	command.Flags().StringVar(&sessionIdleDays, "session-idle-days", sessionIdleDays, "Session idle lifetime in days")
	command.Flags().BoolVar(&chunkedUploads, "chunked-upload", false, "Enable retryable large-file uploads")
	command.Flags().StringVar(&uploadChunkSize, "upload-chunk-size", uploadChunkSize, "Chunk size in MiB (4-16)")
	root.AddCommand(command)
}

func fsValue(command *cobra.Command, flag, value, environment, fallback string) string {
	if command.Flags().Changed(flag) {
		return value
	}
	if environment != "" {
		return environment
	}
	return fallback
}

func parseFSPort(value string) (int, error) {
	if value == "" {
		return 0, fmt.Errorf("'%s' is not a valid port", value)
	}
	port := 0
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character < '0' || character > '9' {
			return 0, fmt.Errorf("'%s' is not a valid port", value)
		}
		port = port*10 + int(character-'0')
		if port > 65535 {
			return 0, errors.New("Port must be between 0 and 65535")
		}
	}
	return port, nil
}

func parseFSIdleDays(value string) (int64, error) {
	parsed, err := parseFSPositiveInteger(value)
	if err != nil || parsed > int64(time.Duration(1<<63-1)/(24*time.Hour)) {
		return 0, fmt.Errorf("'%s' is not a valid positive session idle day count", value)
	}
	return parsed, nil
}

func parseFSChunkSize(value string) (int64, error) {
	parsed, err := parseFSPositiveInteger(value)
	if err != nil || parsed < 4 || parsed > 16 {
		return 0, fmt.Errorf("'%s' is not a valid upload chunk size; use 4-16 MiB", value)
	}
	return parsed, nil
}

func parseFSPositiveInteger(value string) (int64, error) {
	if value == "" {
		return 0, errors.New("empty")
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, errors.New("not decimal")
		}
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, errors.New("not positive")
	}
	return parsed, nil
}

func parseFSBool(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "1", "true":
		return true, nil
	case "0", "false":
		return false, nil
	default:
		return false, fmt.Errorf("'%s' is not a valid chunked-upload value; use true, false, 1, or 0", value)
	}
}
