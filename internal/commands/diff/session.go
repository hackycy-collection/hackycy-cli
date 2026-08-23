package diff

import "context"

type comparisonSession struct {
	workspace      *Workspace
	server         *RunningServer
	bindingAddress string
}

func startComparison(input Input) (*comparisonSession, error) {
	workspace, err := NewWorkspace(WorkspaceOptions{
		BaselineDirectory: input.BaselineDirectory,
		TargetDirectory:   input.TargetDirectory,
		NoGitIgnore:       input.NoGitIgnore,
		Exclusions:        input.Exclusions,
	})
	if err != nil {
		return nil, err
	}
	bindingAddress := diffBindingAddress(input.Public)
	server, err := StartServer(workspace, bindingAddress, input.Port)
	if err != nil {
		return nil, err
	}
	return &comparisonSession{
		workspace:      workspace,
		server:         server,
		bindingAddress: bindingAddress,
	}, nil
}

func (session *comparisonSession) wait(ctx context.Context) error {
	if ctx == nil {
		ctx = context.Background()
	}
	select {
	case <-ctx.Done():
		return session.server.Close()
	case <-session.server.done:
		return session.server.Wait()
	}
}

func diffBindingAddress(public bool) string {
	if public {
		return "0.0.0.0"
	}
	return "127.0.0.1"
}
