package cm

import (
	"context"
	"errors"
	"strings"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

// TestRequest is the typed CLI request for config cm test.
type TestRequest struct {
	Profile string
}

// TestResult records the secret-safe provider response reported by config cm test.
type TestResult struct {
	Content string
}

// TestProfileResolver is the appconfig boundary used by config cm test.
type TestProfileResolver interface {
	ResolveCMProfile(appconfig.CMResolveOptions) (appconfig.ResolvedCMProfile, error)
}

// TestDiagnostic is the profile projection safe for a failed provider test.
type TestDiagnostic struct {
	Provider string
	BaseURL  string
	Model    string
}

// TestPresenter owns the user-visible config cm test outcomes.
type TestPresenter interface {
	Response(string)
	Failure(TestDiagnostic)
}

// TestDependencies are the command-owned adapters for config cm test.
type TestDependencies struct {
	Resolver  TestProfileResolver
	Transport cmTestProviderTransport
	Presenter TestPresenter
}

// TestModule owns config cm test behavior behind its typed command interface.
type TestModule struct {
	resolver  TestProfileResolver
	transport cmTestProviderTransport
	presenter TestPresenter
}

// NewTest constructs a config cm test command module.
func NewTest(dependencies TestDependencies) (*TestModule, error) {
	if dependencies.Resolver == nil {
		return nil, errors.New("config cm test resolver is required")
	}
	if dependencies.Transport == nil {
		return nil, errors.New("config cm test transport is required")
	}
	if dependencies.Presenter == nil {
		return nil, errors.New("config cm test presenter is required")
	}
	return &TestModule{
		resolver:  dependencies.Resolver,
		transport: dependencies.Transport,
		presenter: dependencies.Presenter,
	}, nil
}

// Run resolves one profile, tests it, and presents only secret-safe outcomes.
func (module *TestModule) Run(context context.Context, request TestRequest) (TestResult, error) {
	profile, err := module.resolver.ResolveCMProfile(appconfig.CMResolveOptions{ProfileName: request.Profile})
	if err != nil {
		return TestResult{}, err
	}
	providerResult, err := executeCMTestProvider(context, profile, module.transport)
	if err != nil {
		module.presenter.Failure(TestDiagnostic{Provider: profile.Name, BaseURL: profile.BaseURL, Model: profile.Model})
		return TestResult{}, redactCMTestError(err, profile.APIKey)
	}
	content := redactCMTestText(providerResult.Content, profile.APIKey)
	module.presenter.Response(content)
	return TestResult{Content: content}, nil
}

func redactCMTestError(err error, apiKey string) error {
	message := redactCMTestText(err.Error(), apiKey)
	if message == err.Error() {
		return err
	}
	return cmTestRedactedError{source: err, message: message}
}

func redactCMTestText(value, apiKey string) string {
	if apiKey == "" {
		return value
	}
	return strings.ReplaceAll(value, apiKey, "[REDACTED]")
}

type cmTestRedactedError struct {
	source  error
	message string
}

func (err cmTestRedactedError) Error() string {
	return err.message
}

func (err cmTestRedactedError) Unwrap() error {
	return err.source
}
