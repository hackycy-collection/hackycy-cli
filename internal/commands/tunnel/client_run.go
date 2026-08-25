package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hackycy/hackycy-cli/internal/logging"
)

const (
	clientDefaultYCYVersion              = "0.0.0-dev"
	frpcConfigurationVerificationTimeout = 10 * time.Second
)

var (
	ErrFRPCConfigurationVerificationTimeout = errors.New("FRPC configuration verification timed out")
	errClientControlRevoked                 = errors.New("Tunnel client was revoked")
	errClientControlFatal                   = errors.New("Tunnel client could not process a control message")
)

// ClientRunOptions supplies the private dependencies for an unregistered
// foreground client. The CLI binding and terminal selection remain later work.
type ClientRunOptions struct {
	InstanceIdentity ClientInstanceIdentity
	StateRoot        string
	Logger           logging.Logger
	LogLevel         string
	YCYVersion       string
	HTTPClient       *http.Client
	WebSocketDialer  *websocket.Dialer
	OnAuthenticated  func() error
	Backoff          []time.Duration

	newRuntime func(context.Context, logging.Logger) (ClientFRPRuntime, error)
}

type clientFRPRuntimeStateObserver interface {
	State() FRPSupervisorState
	Observe(func(FRPSupervisorState)) func()
}

type clientFRPRuntimeEnsurer func(context.Context, string, FRPArtifact) (FRPRuntimePaths, error)

type managedClientFRPRuntimeOptions struct {
	Logger              logging.Logger
	frpArtifact         *FRPArtifact
	frpRuntimeDirectory string
	ensureFRPRuntime    clientFRPRuntimeEnsurer
}

// managedClientFRPRuntime binds the reconciler's narrow runtime surface to
// one manifest-pinned, owner-local frpc supervisor.
type managedClientFRPRuntime struct {
	binaryPath string
	supervisor *FRPSupervisor
}

func newManagedClientFRPRuntime(ctx context.Context, options managedClientFRPRuntimeOptions) (*managedClientFRPRuntime, error) {
	artifact := FRPArtifact{}
	if options.frpArtifact != nil {
		artifact = *options.frpArtifact
	} else {
		resolved, err := CurrentFRPArtifact()
		if err != nil {
			return nil, err
		}
		artifact = resolved
	}
	directory := options.frpRuntimeDirectory
	if directory == "" {
		resolved, err := defaultManagedFRPRuntimeDirectory()
		if err != nil {
			return nil, err
		}
		directory = resolved
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return nil, fmt.Errorf("resolve managed FRP runtime directory: %w", err)
	}
	ensure := options.ensureFRPRuntime
	if ensure == nil {
		ensure = EnsureFRPRuntimeAt
	}
	paths, err := ensure(ctx, directory, artifact)
	if err != nil {
		return nil, err
	}
	expected := frpRuntimePaths(directory, artifact.Target)
	if paths != expected {
		return nil, errors.New("managed FRP preparation returned paths outside the pinned runtime directory")
	}
	supervisor, err := NewFRPSupervisor(FRPSupervisorOptions{
		BinaryPath: paths.FRPC,
		Role:       FRPRoleClient,
		Logger:     options.Logger,
	})
	if err != nil {
		return nil, err
	}
	return &managedClientFRPRuntime{binaryPath: paths.FRPC, supervisor: supervisor}, nil
}

func defaultClientFRPRuntime(ctx context.Context, logger logging.Logger) (ClientFRPRuntime, error) {
	return newManagedClientFRPRuntime(ctx, managedClientFRPRuntimeOptions{Logger: logger})
}

func (runtime *managedClientFRPRuntime) Verify(ctx context.Context, configurationPath string) error {
	if runtime == nil || strings.TrimSpace(runtime.binaryPath) == "" {
		return errors.New("Tunnel client FRP runtime is unavailable")
	}
	return verifyFRPCConfiguration(ctx, runtime.binaryPath, configurationPath, frpcConfigurationVerificationTimeout)
}

func (runtime *managedClientFRPRuntime) Start(configurationPath string) error {
	if runtime == nil || runtime.supervisor == nil {
		return errors.New("Tunnel client FRP runtime is unavailable")
	}
	return runtime.supervisor.Start(configurationPath)
}

func (runtime *managedClientFRPRuntime) Restart() error {
	if runtime == nil || runtime.supervisor == nil {
		return errors.New("Tunnel client FRP runtime is unavailable")
	}
	return runtime.supervisor.Restart()
}

func (runtime *managedClientFRPRuntime) Stop() error {
	if runtime == nil || runtime.supervisor == nil {
		return nil
	}
	return runtime.supervisor.Stop()
}

func (runtime *managedClientFRPRuntime) State() FRPSupervisorState {
	if runtime == nil || runtime.supervisor == nil {
		return FRPSupervisorState{State: FRPProcessStopped}
	}
	return runtime.supervisor.State()
}

func (runtime *managedClientFRPRuntime) Observe(listener func(FRPSupervisorState)) func() {
	if runtime == nil || runtime.supervisor == nil {
		return func() {}
	}
	return runtime.supervisor.Observe(listener)
}

func verifyFRPCConfiguration(ctx context.Context, binaryPath, configurationPath string, timeout time.Duration) error {
	if ctx == nil {
		ctx = context.Background()
	}
	verificationContext, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, err := exec.CommandContext(verificationContext, binaryPath, "verify", "-c", configurationPath).CombinedOutput()
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if errors.Is(verificationContext.Err(), context.DeadlineExceeded) {
		return fmt.Errorf("%w: frpc did not respond within %s", ErrFRPCConfigurationVerificationTimeout, timeout)
	}
	return fmt.Errorf("frpc rejected the generated configuration: %w", err)
}

// RunClient owns one resolved client instance through authentication, v3
// reconciliation, frpc supervision, reconnect, and final ordered shutdown.
func RunClient(ctx context.Context, config ClientConfig, options ClientRunOptions) (result error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if ctx.Err() != nil {
		return nil
	}
	backoff, err := clientReconnectBackoff(options.Backoff)
	if err != nil {
		return err
	}
	instance, err := AcquireClientInstance(config, options.InstanceIdentity, ClientInstanceOptions{
		StateRoot: options.StateRoot,
		OnCleanupError: func(cleanupErr error) {
			options.Logger.Warn("Could not clean expired Tunnel client state", map[string]any{"error": cleanupErr.Error()})
		},
	})
	if err != nil {
		options.Logger.Error("Could not start Tunnel client", map[string]any{"error": err.Error()})
		return err
	}

	connections := &clientControlConnectionHolder{}
	reporter := &clientProcessStateReporter{state: FRPSupervisorState{State: FRPProcessStopped}}
	var runtime ClientFRPRuntime
	var reconciler *ClientReconciler
	var stopObserving func()
	defer func() {
		_ = connections.Close()
		reporter.Clear()
		if stopObserving != nil {
			stopObserving()
		}
		if reconciler != nil {
			result = errors.Join(result, reconciler.Stop())
		} else if runtime != nil {
			result = errors.Join(result, runtime.Stop())
		}
		result = errors.Join(result, instance.Release())
		if result != nil && ctx.Err() == nil {
			options.Logger.Error("Tunnel client failed", map[string]any{"error": result.Error()})
		}
		options.Logger.Info("Tunnel client stopped", nil)
	}()

	shutdownDone := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = connections.Close()
		case <-shutdownDone:
		}
	}()
	defer close(shutdownDone)

	lastAppliedRevision := int64(0)
	if applied, found := ReadClientAppliedState(instance.StateDirectory); found {
		lastAppliedRevision = applied.Revision
	}
	agent, err := NewClientAgent(ClientAgentOptions{
		Config:              config,
		YCYVersion:          clientRunYCYVersion(options.YCYVersion),
		LastAppliedRevision: lastAppliedRevision,
		HTTPClient:          options.HTTPClient,
		WebSocketDialer:     options.WebSocketDialer,
		OnAuthenticated: func() error {
			if options.OnAuthenticated == nil {
				return nil
			}
			if rememberErr := options.OnAuthenticated(); rememberErr != nil {
				options.Logger.Warn("Could not remember Tunnel connection", map[string]any{"error": rememberErr.Error()})
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	options.Logger.Info("Tunnel client started", map[string]any{"server": config.Server.String(), "stateDirectory": instance.StateDirectory})
	newRuntime := options.newRuntime
	if newRuntime == nil {
		newRuntime = defaultClientFRPRuntime
	}
	failures := 0
	for {
		if ctx.Err() != nil {
			return nil
		}
		connection, connectErr := agent.Connect(ctx)
		if connectErr != nil {
			if ctx.Err() != nil {
				return nil
			}
			if clientControlFailureIsFatal(connectErr) {
				return connectErr
			}
			clientReconnect(options.Logger, failures, backoff)
			failures++
			continue
		}
		failures = 0
		connections.Set(connection)
		if ctx.Err() != nil {
			return nil
		}

		if runtime == nil {
			runtime, err = newRuntime(ctx, options.Logger)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				}
				closeClientControlSocket(connection.socket, 4406, "Client failed to prepare frpc")
				return err
			}
			reconciler, err = NewClientReconciler(ClientReconcilerOptions{
				StateDirectory: instance.StateDirectory,
				Runtime:        runtime,
				LogLevel:       options.LogLevel,
			})
			if err != nil {
				closeClientControlSocket(connection.socket, 4406, "Client failed to prepare frpc")
				return err
			}
			if observer, supported := runtime.(clientFRPRuntimeStateObserver); supported {
				reporter.Report(observer.State())
				stopObserving = observer.Observe(reporter.Report)
			}
		}
		reporter.Set(connection)

		connectionErr := runClientControlConnection(ctx, agent, connection, reconciler, reporter)
		reporter.ClearConnection(connection)
		connections.Clear(connection)
		_ = connection.Close()
		if errors.Is(connectionErr, errClientControlRevoked) || ctx.Err() != nil {
			return nil
		}
		if clientControlFailureIsFatal(connectionErr) {
			return connectionErr
		}
		clientReconnect(options.Logger, failures, backoff)
		failures++
	}
}

func clientRunYCYVersion(value string) string {
	if strings.TrimSpace(value) == "" {
		return clientDefaultYCYVersion
	}
	return value
}

func clientReconnectBackoff(configured []time.Duration) ([]time.Duration, error) {
	backoff := configured
	if len(backoff) == 0 {
		backoff = defaultFRPRecoveryBackoff
	}
	backoff = append([]time.Duration(nil), backoff...)
	for _, delay := range backoff {
		if delay < 0 {
			return nil, fmt.Errorf("Tunnel client reconnect delay must not be negative")
		}
	}
	return backoff, nil
}

func clientReconnect(logger logging.Logger, failures int, backoff []time.Duration) {
	index := failures
	if index >= len(backoff) {
		index = len(backoff) - 1
	}
	delay := backoff[index]
	logger.Debug("Scheduling Tunnel control reconnect", map[string]any{"delay": delay.String(), "attempt": failures + 1})
	time.Sleep(delay)
}

func clientControlFailureIsFatal(err error) bool {
	return errors.Is(err, ErrClientAuthentication) || errors.Is(err, ErrClientIncompatible) || errors.Is(err, ErrClientProtocol) || errors.Is(err, errClientControlFatal)
}

func runClientControlConnection(ctx context.Context, agent *ClientAgent, connection *ClientControlConnection, reconciler *ClientReconciler, reporter *clientProcessStateReporter) error {
	configuration := clientDesiredConfigurationFromWelcome(connection.Welcome)
	if err := reportClientApply(ctx, agent, connection, reconciler, reporter, configuration); err != nil {
		return err
	}
	for {
		source, err := connection.ReadMessage()
		if err != nil {
			return err
		}
		message, err := decodeClientControlMessage(source)
		if err != nil {
			if errors.Is(err, ErrClientIncompatible) {
				closeClientControlSocket(connection.socket, 4406, "Client failed to process control message")
			} else {
				closeClientControlSocket(connection.socket, 4400, "Invalid control message")
			}
			return err
		}
		switch message.kind {
		case "desired_state":
			configuration.Snapshot = message.desired.Snapshot
			if err := reportClientApply(ctx, agent, connection, reconciler, reporter, configuration); err != nil {
				return err
			}
		case "restart_frpc":
			if err := reconciler.Restart(); err != nil {
				closeClientControlSocket(connection.socket, 4406, "Client failed to process control message")
				return fmt.Errorf("%w: restart frpc: %v", errClientControlFatal, err)
			}
			if err := reporter.Publish(); err != nil {
				return fmt.Errorf("report Tunnel client process state: %w", err)
			}
		case "revoke":
			return errClientControlRevoked
		default:
			return fmt.Errorf("%w: unexpected control message", ErrClientProtocol)
		}
	}
}

func clientDesiredConfigurationFromWelcome(welcome AgentWelcome) ClientDesiredConfiguration {
	return ClientDesiredConfiguration{
		AdvertisedFRPHost: welcome.AdvertisedFRPHost,
		AdvertisedFRPPort: welcome.AdvertisedFRPPort,
		InternalFRPToken:  welcome.InternalFRPToken,
		Snapshot:          welcome.Snapshot,
	}
}

func reportClientApply(ctx context.Context, agent *ClientAgent, connection *ClientControlConnection, reconciler *ClientReconciler, reporter *clientProcessStateReporter, desired ClientDesiredConfiguration) error {
	err := reconciler.Apply(ctx, desired)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	revision := desired.Snapshot.Revision
	if err == nil {
		agent.RecordAppliedRevision(revision)
		if writeErr := connection.WriteJSON(ApplyResult{
			Type:                  "apply_result",
			TunnelProtocolVersion: TunnelProtocolVersion,
			Revision:              revision,
			Success:               true,
		}); writeErr != nil {
			return fmt.Errorf("acknowledge Tunnel desired state: %w", writeErr)
		}
	} else {
		code := clientReconciliationErrorCode(err)
		if code == "" {
			code = "APPLY_FAILED"
		}
		if writeErr := connection.WriteJSON(ApplyResult{
			Type:                  "apply_result",
			TunnelProtocolVersion: TunnelProtocolVersion,
			Revision:              revision,
			Success:               false,
			Error: &StructuredRuntimeError{
				Code: code, Message: err.Error(), Revision: &revision,
			},
		}); writeErr != nil {
			return fmt.Errorf("report Tunnel desired-state failure: %w", writeErr)
		}
	}
	if writeErr := reporter.Publish(); writeErr != nil {
		return fmt.Errorf("report Tunnel client process state: %w", writeErr)
	}
	return nil
}

type clientControlMessage struct {
	kind    string
	desired DesiredState
}

func decodeClientControlMessage(source []byte) (clientControlMessage, error) {
	var envelope struct {
		Type                  string `json:"type"`
		TunnelProtocolVersion int    `json:"tunnelProtocolVersion"`
	}
	if err := json.Unmarshal(source, &envelope); err != nil {
		return clientControlMessage{}, fmt.Errorf("%w: decode control message", ErrClientProtocol)
	}
	if envelope.Type == "incompatible" {
		var incompatible Incompatible
		if err := json.Unmarshal(source, &incompatible); err != nil || incompatible.TunnelProtocolVersion != TunnelProtocolVersion || strings.TrimSpace(incompatible.Message) == "" {
			return clientControlMessage{}, fmt.Errorf("%w: invalid incompatibility message", ErrClientProtocol)
		}
		return clientControlMessage{}, fmt.Errorf("%w: %s", ErrClientIncompatible, incompatible.Message)
	}
	if envelope.TunnelProtocolVersion != TunnelProtocolVersion {
		return clientControlMessage{}, fmt.Errorf("%w: Control plane uses an unsupported tunnel protocol; upgrade ycy", ErrClientIncompatible)
	}
	switch envelope.Type {
	case "desired_state":
		var desired DesiredState
		if err := json.Unmarshal(source, &desired); err != nil || desired.Snapshot.Revision < 0 || desired.Snapshot.Revision > clientMaximumSafeInteger {
			return clientControlMessage{}, fmt.Errorf("%w: invalid desired-state message", ErrClientProtocol)
		}
		return clientControlMessage{kind: envelope.Type, desired: desired}, nil
	case "restart_frpc":
		return clientControlMessage{kind: envelope.Type}, nil
	case "revoke":
		var revoke Revoke
		if err := json.Unmarshal(source, &revoke); err != nil || (revoke.Reason != "rotated" && revoke.Reason != "deleted") {
			return clientControlMessage{}, fmt.Errorf("%w: invalid revoke message", ErrClientProtocol)
		}
		return clientControlMessage{kind: envelope.Type}, nil
	default:
		return clientControlMessage{}, fmt.Errorf("%w: unexpected control message", ErrClientProtocol)
	}
}

type clientControlConnectionHolder struct {
	mu         sync.Mutex
	connection *ClientControlConnection
}

func (holder *clientControlConnectionHolder) Set(connection *ClientControlConnection) {
	holder.mu.Lock()
	holder.connection = connection
	holder.mu.Unlock()
}

func (holder *clientControlConnectionHolder) Clear(connection *ClientControlConnection) {
	holder.mu.Lock()
	if holder.connection == connection {
		holder.connection = nil
	}
	holder.mu.Unlock()
}

func (holder *clientControlConnectionHolder) Close() error {
	holder.mu.Lock()
	connection := holder.connection
	holder.connection = nil
	holder.mu.Unlock()
	if connection == nil {
		return nil
	}
	return connection.Close()
}

type clientProcessStateReporter struct {
	mu         sync.Mutex
	connection *ClientControlConnection
	state      FRPSupervisorState
}

func (reporter *clientProcessStateReporter) Set(connection *ClientControlConnection) {
	reporter.mu.Lock()
	reporter.connection = connection
	reporter.mu.Unlock()
	_ = reporter.Publish()
}

func (reporter *clientProcessStateReporter) Clear() {
	reporter.mu.Lock()
	reporter.connection = nil
	reporter.mu.Unlock()
}

func (reporter *clientProcessStateReporter) ClearConnection(connection *ClientControlConnection) {
	reporter.mu.Lock()
	if reporter.connection == connection {
		reporter.connection = nil
	}
	reporter.mu.Unlock()
}

func (reporter *clientProcessStateReporter) Report(state FRPSupervisorState) {
	reporter.mu.Lock()
	reporter.state = cloneFRPSupervisorState(state)
	reporter.mu.Unlock()
	_ = reporter.Publish()
}

func (reporter *clientProcessStateReporter) Publish() error {
	reporter.mu.Lock()
	connection := reporter.connection
	state := cloneFRPSupervisorState(reporter.state)
	reporter.mu.Unlock()
	if connection == nil {
		return nil
	}
	return connection.WriteJSON(ProcessState{
		Type:                  "process_state",
		TunnelProtocolVersion: TunnelProtocolVersion,
		State:                 state.State,
		Error:                 state.Error,
	})
}
