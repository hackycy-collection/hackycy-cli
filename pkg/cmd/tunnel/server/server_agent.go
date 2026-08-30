package server

import (
	"context"
	"errors"
	"fmt"
	tunnelruntime "github.com/hackycy/hackycy-cli/internal/tunnelruntime"
	"strings"
	"sync"
)

var ErrServerAgentGatewayConfiguration = errors.New("Tunnel server agent gateway configuration is invalid")

// ServerAgentFRPSStateProvider keeps agent authorization independent of the
// HTTP state projection while preserving the managed FRPS availability gate.
type ServerAgentFRPSStateProvider interface {
	FRPSState() tunnelruntime.FRPSupervisorState
}

// ServerAgentWelcomeSettings contains the deployment secret and endpoint that
// are safe only for an already authenticated agent welcome frame.
type ServerAgentWelcomeSettings struct {
	AdvertisedFRPHost string
	AdvertisedFRPPort int64
	InternalFRPToken  string
}

// ServerAgentWelcomeSource keeps deployment-secret ownership with the FRPS
// composition instead of the HTTP state projection.
type ServerAgentWelcomeSource interface {
	AgentWelcomeSettings(requestHost string) ServerAgentWelcomeSettings
}

type ServerAgentGatewayOptions struct {
	ControlPlane  *ServerControlPlane
	FRPS          ServerAgentFRPSStateProvider
	WelcomeSource ServerAgentWelcomeSource
}

// ServerAgentGateway owns process-local agent admission, connection runtime
// projection, handshake reads, and desired/revoke presentation. Durable
// desired state remains in ServerControlPlane; later protocol messages stay
// outside this boundary.
type ServerAgentGateway struct {
	controlPlane  *ServerControlPlane
	frps          ServerAgentFRPSStateProvider
	welcomeSource ServerAgentWelcomeSource

	mu       sync.RWMutex
	slots    map[string]serverAgentSlot
	runtime  map[string]serverAgentRuntime
	nextSlot uint64
}

type serverAgentSlot struct {
	generation uint64
	active     bool
	revoking   bool
	connection *ServerAgentConnection
}

type serverAgentRuntime struct {
	processState tunnelruntime.FRPProcessState
	lastError    *tunnelruntime.StructuredRuntimeError
}

func NewServerAgentGateway(options ServerAgentGatewayOptions) (*ServerAgentGateway, error) {
	if options.ControlPlane == nil {
		return nil, fmt.Errorf("%w: control plane is required", ErrServerAgentGatewayConfiguration)
	}
	if options.FRPS == nil {
		return nil, fmt.Errorf("%w: managed frps state is required", ErrServerAgentGatewayConfiguration)
	}
	gateway := &ServerAgentGateway{
		controlPlane:  options.ControlPlane,
		frps:          options.FRPS,
		welcomeSource: options.WelcomeSource,
		slots:         make(map[string]serverAgentSlot),
		runtime:       make(map[string]serverAgentRuntime),
	}
	gateway.controlPlane.Subscribe(gateway.handleControlPlaneEvent)
	return gateway, nil
}

// State combines process-local connection ownership and the latest agent
// process report. Desired and applied revisions remain durable control-plane
// state rather than gateway state.
func (gateway *ServerAgentGateway) State(clientID string) ServerClientRuntimeState {
	state := ServerClientRuntimeState{
		ConnectionState: ServerClientDisconnected,
		ProcessState:    tunnelruntime.FRPProcessStopped,
	}
	if gateway == nil {
		return state
	}
	gateway.mu.RLock()
	slot, connected := gateway.slots[clientID]
	runtime, reported := gateway.runtime[clientID]
	gateway.mu.RUnlock()
	if reported {
		state.ProcessState = runtime.processState
		state.LastError = cloneServerAgentRuntimeError(runtime.lastError)
	}
	if connected && slot.revoking {
		state.ConnectionState = ServerClientRevocationPending
		return state
	}
	if connected && slot.active {
		state.ConnectionState = ServerClientConnected
		return state
	}
	client, err := gateway.controlPlane.GetClient(context.Background(), clientID)
	if err == nil && client.RevocationPending {
		state.ConnectionState = ServerClientRevocationPending
	}
	return state
}

func (gateway *ServerAgentGateway) recordProcessState(clientID string, slot uint64, processState tunnelruntime.FRPProcessState, lastError *tunnelruntime.StructuredRuntimeError) bool {
	if gateway == nil {
		return false
	}
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	current, found := gateway.slots[clientID]
	if !found || current.generation != slot || !current.active || current.revoking {
		return false
	}
	runtime := gateway.runtime[clientID]
	if runtime.processState == "" {
		runtime.processState = tunnelruntime.FRPProcessStopped
	}
	runtime.processState = processState
	if lastError != nil {
		runtime.lastError = cloneServerAgentRuntimeError(lastError)
	} else if runtime.lastError == nil || runtime.lastError.Revision == nil {
		runtime.lastError = nil
	}
	gateway.runtime[clientID] = runtime
	return true
}

func (gateway *ServerAgentGateway) recordRuntimeError(clientID string, slot uint64, lastError *tunnelruntime.StructuredRuntimeError) bool {
	if gateway == nil {
		return false
	}
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	current, found := gateway.slots[clientID]
	if !found || current.generation != slot || !current.active || current.revoking {
		return false
	}
	runtime := gateway.runtime[clientID]
	if runtime.processState == "" {
		runtime.processState = tunnelruntime.FRPProcessStopped
	}
	runtime.lastError = cloneServerAgentRuntimeError(lastError)
	gateway.runtime[clientID] = runtime
	return true
}

func cloneServerAgentRuntimeError(value *tunnelruntime.StructuredRuntimeError) *tunnelruntime.StructuredRuntimeError {
	if value == nil {
		return nil
	}
	copy := *value
	if value.Revision != nil {
		revision := *value.Revision
		copy.Revision = &revision
	}
	return &copy
}

// RestartFRPC sends the explicit restart request only to a currently active
// agent that has completed its server-frame presentation.
func (gateway *ServerAgentGateway) RestartFRPC(clientID string) bool {
	if gateway == nil {
		return false
	}
	gateway.mu.RLock()
	slot, found := gateway.slots[clientID]
	gateway.mu.RUnlock()
	if !found || !slot.active || slot.revoking || slot.connection == nil {
		return false
	}
	return slot.connection.RestartFRPC()
}

// ServerAgentReservation holds the one pending connection slot for a Client
// Token until the HTTP upgrade either takes ownership or is rejected.
type ServerAgentReservation struct {
	gateway     *ServerAgentGateway
	clientID    string
	slot        uint64
	releaseOnce sync.Once
}

func (reservation *ServerAgentReservation) ClientID() string {
	if reservation == nil {
		return ""
	}
	return reservation.clientID
}

// Release returns a pending slot to the gateway. It is safe to call more than
// once and cannot release a later reservation for the same client.
func (reservation *ServerAgentReservation) Release() {
	if reservation == nil || reservation.gateway == nil {
		return
	}
	reservation.releaseOnce.Do(func() {
		reservation.gateway.release(reservation.clientID, reservation.slot)
	})
}

// Activate transfers this pending slot to the accepted WebSocket lifetime.
// A stale or already-released reservation cannot take ownership of a later
// connection for the same client.
func (reservation *ServerAgentReservation) Activate() *ServerAgentConnection {
	if reservation == nil || reservation.gateway == nil {
		return nil
	}
	var connection *ServerAgentConnection
	reservation.releaseOnce.Do(func() {
		candidate := &ServerAgentConnection{
			gateway:  reservation.gateway,
			clientID: reservation.clientID,
			slot:     reservation.slot,
		}
		if reservation.gateway.activate(reservation.clientID, reservation.slot, candidate) {
			connection = candidate
		}
	})
	return connection
}

// ServerAgentConnection owns one gateway slot while its accepted WebSocket is
// open, keeps the first accepted hello, and serializes server-frame delivery.
type ServerAgentConnection struct {
	gateway            *ServerAgentGateway
	clientID           string
	slot               uint64
	closeOnce          sync.Once
	helloMu            sync.Mutex
	helloAccepted      bool
	hello              tunnelruntime.AgentHello
	presentationMu     sync.Mutex
	presentationActive bool
	closed             bool
	writeFrame         func(any) error
	closeSocket        func(*ServerAgentProtocolError)
}

func (connection *ServerAgentConnection) ClientID() string {
	if connection == nil {
		return ""
	}
	return connection.clientID
}

// AcknowledgeReplacementToken clears durable revocation state only after the
// replacement token has acquired an active agent connection.
func (connection *ServerAgentConnection) AcknowledgeReplacementToken(ctx context.Context) error {
	if connection == nil || connection.gateway == nil {
		return errors.New("Tunnel server agent session is unavailable")
	}
	return connection.gateway.controlPlane.AcknowledgeReplacementToken(ctx, connection.clientID)
}

// Close makes the client eligible for a later authorization. It is idempotent
// and cannot release a newer connection's slot.
func (connection *ServerAgentConnection) Close() {
	if connection == nil || connection.gateway == nil {
		return
	}
	connection.closeOnce.Do(func() {
		connection.presentationMu.Lock()
		connection.closed = true
		connection.presentationMu.Unlock()
		connection.gateway.release(connection.clientID, connection.slot)
	})
}

// AttachCloser keeps physical WebSocket ownership at the HTTP boundary while
// letting the gateway terminate an active session after durable revocation.
func (connection *ServerAgentConnection) AttachCloser(closeSocket func(*ServerAgentProtocolError)) {
	if connection == nil {
		return
	}
	connection.presentationMu.Lock()
	if !connection.closed {
		connection.closeSocket = closeSocket
	}
	connection.presentationMu.Unlock()
}

// Authorize validates one Bearer Client Token and atomically reserves its
// pending control-session slot. WebSocket upgrade handling is intentionally
// separate so a non-upgrade request can release this reservation first.
func (gateway *ServerAgentGateway) Authorize(ctx context.Context, authorization string) (*ServerAgentReservation, error) {
	if gateway.frps.FRPSState().State != tunnelruntime.FRPProcessRunning {
		return nil, serverDomainError("FRPS_UNAVAILABLE", "Managed frps is not running")
	}
	token := parseServerAgentBearerToken(authorization)
	if token == "" {
		return nil, serverDomainError("AUTHENTICATION_FAILED", "Client Token is invalid")
	}
	client, err := gateway.controlPlane.FindClientByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, serverDomainError("AUTHENTICATION_FAILED", "Client Token is invalid")
	}
	slot, reserved := gateway.reserve(client.ID)
	if !reserved {
		return nil, serverDomainError("CLIENT_CONNECTED", "Client Token already has an active control session")
	}
	return &ServerAgentReservation{gateway: gateway, clientID: client.ID, slot: slot}, nil
}

func parseServerAgentBearerToken(authorization string) string {
	separator := strings.IndexByte(authorization, ' ')
	if separator < 1 || !strings.EqualFold(authorization[:separator], "bearer") {
		return ""
	}
	return strings.TrimSpace(authorization[separator+1:])
}

func (gateway *ServerAgentGateway) reserve(clientID string) (uint64, bool) {
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	if _, found := gateway.slots[clientID]; found {
		return 0, false
	}
	gateway.nextSlot++
	if gateway.nextSlot == 0 {
		gateway.nextSlot++
	}
	gateway.slots[clientID] = serverAgentSlot{generation: gateway.nextSlot}
	return gateway.nextSlot, true
}

func (gateway *ServerAgentGateway) activate(clientID string, slot uint64, connection *ServerAgentConnection) bool {
	gateway.mu.Lock()
	defer gateway.mu.Unlock()
	current, found := gateway.slots[clientID]
	if !found || current.generation != slot || current.active || connection == nil {
		return false
	}
	current.active = true
	current.connection = connection
	gateway.slots[clientID] = current
	if _, found := gateway.runtime[clientID]; !found {
		gateway.runtime[clientID] = serverAgentRuntime{processState: tunnelruntime.FRPProcessStopped}
	}
	return true
}

func (gateway *ServerAgentGateway) release(clientID string, slot uint64) {
	gateway.mu.Lock()
	if current, found := gateway.slots[clientID]; found && current.generation == slot {
		delete(gateway.slots, clientID)
	}
	gateway.mu.Unlock()
}

func (gateway *ServerAgentGateway) handleControlPlaneEvent(event ServerControlPlaneEvent) {
	switch event.Type {
	case serverDesiredState:
		gateway.mu.RLock()
		slot, found := gateway.slots[event.ClientID]
		gateway.mu.RUnlock()
		if found && slot.active && !slot.revoking && slot.connection != nil {
			slot.connection.PresentDesiredState()
		}
	case serverClientRotated:
		gateway.revoke(event.ClientID, "rotated", false)
	case serverClientDeleted:
		gateway.revoke(event.ClientID, "deleted", true)
	}
}

func (gateway *ServerAgentGateway) revoke(clientID, reason string, deleted bool) {
	gateway.mu.Lock()
	slot, found := gateway.slots[clientID]
	if !found {
		if deleted {
			delete(gateway.runtime, clientID)
		}
		gateway.mu.Unlock()
		return
	}
	if deleted {
		delete(gateway.slots, clientID)
		delete(gateway.runtime, clientID)
		gateway.mu.Unlock()
		if slot.connection != nil {
			slot.connection.Revoke(reason)
		}
		return
	}
	if !slot.active || slot.connection == nil {
		delete(gateway.slots, clientID)
		gateway.mu.Unlock()
		return
	}
	slot.revoking = true
	gateway.slots[clientID] = slot
	gateway.mu.Unlock()
	slot.connection.Revoke(reason)
}
