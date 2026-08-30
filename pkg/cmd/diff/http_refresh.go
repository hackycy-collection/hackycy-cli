package diff

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type refreshCoordinator struct {
	workspace *Workspace
	mu        sync.Mutex
	active    *RefreshRun
}

func newRefreshCoordinator(workspace *Workspace) *refreshCoordinator {
	return &refreshCoordinator{workspace: workspace}
}

func (handler *diffHTTPHandler) serveRefresh(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost && request.Method != http.MethodDelete {
		writeHTTPError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Use POST or DELETE")
		return
	}
	if !isHTTPRefreshOriginAllowed(request) {
		writeHTTPError(writer, http.StatusForbidden, "ORIGIN_FORBIDDEN", "Refresh requests must be same-origin")
		return
	}
	if request.Method == http.MethodDelete {
		handler.cancelHTTPRefresh()
		writeHTTPNoContent(writer)
		return
	}

	if err := handler.startHTTPRefresh(); err != nil {
		if errors.Is(err, ErrRefreshActive) {
			writeHTTPError(writer, http.StatusConflict, "REFRESH_ACTIVE", "A refresh is already active")
			return
		}
		writeHTTPError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "Refresh could not be started")
		return
	}
	writeHTTPJSON(writer, http.StatusAccepted, struct {
		Accepted bool `json:"accepted"`
	}{Accepted: true})
}

func isHTTPRefreshOriginAllowed(request *http.Request) bool {
	origin := request.Header.Get("Origin")
	if origin == "" {
		return true
	}
	scheme := "http"
	if request.TLS != nil {
		scheme = "https"
	}
	return origin == scheme+"://"+request.Host
}

func (handler *diffHTTPHandler) startHTTPRefresh() error {
	return handler.refresh.Start()
}

func (handler *diffHTTPHandler) cancelHTTPRefresh() {
	handler.refresh.Cancel()
}

func (coordinator *refreshCoordinator) Start() error {
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if coordinator.active != nil {
		return ErrRefreshActive
	}
	run, err := coordinator.workspace.StartRefresh(context.Background())
	if err != nil {
		return err
	}
	coordinator.active = run
	go coordinator.clear(run)
	return nil
}

func (coordinator *refreshCoordinator) Cancel() {
	coordinator.mu.Lock()
	run := coordinator.active
	coordinator.mu.Unlock()
	if run != nil {
		run.Cancel()
	}
}

func (coordinator *refreshCoordinator) CancelAndWait() {
	coordinator.mu.Lock()
	run := coordinator.active
	coordinator.mu.Unlock()
	if run == nil {
		return
	}
	run.Cancel()
	_, _ = run.Wait(context.Background())
}

func (coordinator *refreshCoordinator) clear(run *RefreshRun) {
	_, _ = run.Wait(context.Background())
	coordinator.mu.Lock()
	defer coordinator.mu.Unlock()
	if coordinator.active == run {
		coordinator.active = nil
	}
}
