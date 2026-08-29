package main

import (
	"io"

	"github.com/hackycy/hackycy-cli/internal/logging"
	commandfactory "github.com/hackycy/hackycy-cli/pkg/cmd/factory"
	rootcommand "github.com/hackycy/hackycy-cli/pkg/cmd/root"
	"github.com/hackycy/hackycy-cli/pkg/cmdutil"
)

type rootTestDependencies struct {
	Out               io.Writer
	Err               io.Writer
	Environment       func(string) string
	EnvironmentLookup func(string) (string, bool)
	Logging           *logging.Runtime
	Diff              rootcommand.DiffHandler
	FS                rootcommand.FSHandler
	TunnelServer      rootcommand.TunnelServerHandler
	TunnelConnect     rootcommand.TunnelConnectHandler
	Upgrade           rootcommand.UpgradeHandler
}

func newRootCommandForTest(version string, dependencies rootTestDependencies) (*rootcommand.App, error) {
	factory := commandfactory.New(commandfactory.Options{
		Version: version,
		IOStreams: cmdutil.IOStreams{
			Out:    dependencies.Out,
			ErrOut: dependencies.Err,
		},
		Environment:       dependencies.Environment,
		EnvironmentLookup: dependencies.EnvironmentLookup,
	})
	if dependencies.Logging != nil {
		factory.Logging = dependencies.Logging
	}
	return rootcommand.New(factory, rootcommand.Dependencies{
		Diff:          dependencies.Diff,
		FS:            dependencies.FS,
		TunnelServer:  dependencies.TunnelServer,
		TunnelConnect: dependencies.TunnelConnect,
		Upgrade:       dependencies.Upgrade,
	})
}
