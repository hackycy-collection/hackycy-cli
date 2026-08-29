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
	ExportEnv         rootcommand.ExportEnvHandler
	ConfigForkList    rootcommand.ConfigForkListHandler
	ConfigForkAdd     rootcommand.ConfigForkAddHandler
	ConfigForkRemove  rootcommand.ConfigForkRemoveHandler
	ConfigCMList      rootcommand.ConfigCMListHandler
	ConfigCMAdd       rootcommand.ConfigCMAddHandler
	ConfigCMUse       rootcommand.ConfigCMUseHandler
	ConfigCMSet       rootcommand.ConfigCMSetHandler
	ConfigCMRemove    rootcommand.ConfigCMRemoveHandler
	ConfigCMTest      rootcommand.ConfigCMTestHandler
	RM                rootcommand.RmHandler
	Run               rootcommand.RunHandler
	GitHeat           rootcommand.GitHeatHandler
	GitPulse          rootcommand.GitPulseHandler
	GitFork           rootcommand.GitForkHandler
	GitCM             rootcommand.GitCMHandler
	ZIP               rootcommand.ZipHandler
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
		ExportEnv:        dependencies.ExportEnv,
		ConfigForkList:   dependencies.ConfigForkList,
		ConfigForkAdd:    dependencies.ConfigForkAdd,
		ConfigForkRemove: dependencies.ConfigForkRemove,
		ConfigCMList:     dependencies.ConfigCMList,
		ConfigCMAdd:      dependencies.ConfigCMAdd,
		ConfigCMUse:      dependencies.ConfigCMUse,
		ConfigCMSet:      dependencies.ConfigCMSet,
		ConfigCMRemove:   dependencies.ConfigCMRemove,
		ConfigCMTest:     dependencies.ConfigCMTest,
		RM:               dependencies.RM,
		Run:              dependencies.Run,
		GitHeat:          dependencies.GitHeat,
		GitPulse:         dependencies.GitPulse,
		GitFork:          dependencies.GitFork,
		GitCM:            dependencies.GitCM,
		ZIP:              dependencies.ZIP,
		Diff:             dependencies.Diff,
		FS:               dependencies.FS,
		TunnelServer:     dependencies.TunnelServer,
		TunnelConnect:    dependencies.TunnelConnect,
		Upgrade:          dependencies.Upgrade,
	})
}
