package cm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/hackycy/hackycy-cli/internal/appconfig"
)

func TestCMTestModuleResolvesRunsAndPresentsTheProviderResponse(t *testing.T) {
	resolver := &recordingCMTestResolver{profile: cmTestProfile()}
	presenter := &recordingCMTestPresenter{}
	transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"ok"}}]}`))}, nil
	})
	module := newCMTestModule(t, resolver, transport, presenter)

	result, err := module.Run(context.Background(), TestRequest{Profile: "work"})
	if err != nil {
		t.Fatalf("Run() returned an error: %v", err)
	}
	if resolver.options != (appconfig.CMResolveOptions{ProfileName: "work"}) {
		t.Fatalf("resolver options = %#v", resolver.options)
	}
	if result != (TestResult{Content: "ok"}) || presenter.response != "ok" || presenter.failure != (TestDiagnostic{}) {
		t.Fatalf("Run() = (%#v, presenter=%#v)", result, presenter)
	}
}

func TestCMTestModulePresentsOnlySafeFailureDiagnostics(t *testing.T) {
	profile := cmTestProfile()
	presenter := &recordingCMTestPresenter{}
	failure := errors.New("provider rejected test-api-key")
	transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
		return nil, failure
	})
	module := newCMTestModule(t, &recordingCMTestResolver{profile: profile}, transport, presenter)

	result, err := module.Run(context.Background(), TestRequest{})
	if !errors.Is(err, failure) || strings.Contains(err.Error(), profile.APIKey) || result != (TestResult{}) {
		t.Fatalf("Run() = (%#v, %v), want redacted provider failure", result, err)
	}
	if presenter.failure != (TestDiagnostic{Provider: profile.Name, BaseURL: profile.BaseURL, Model: profile.Model}) || presenter.response != "" {
		t.Fatalf("presenter = %#v", presenter)
	}
}

func TestCMTestModuleDoesNotPresentAResolverFailure(t *testing.T) {
	failure := errors.New("No usable CM profile found")
	presenter := &recordingCMTestPresenter{}
	module := newCMTestModule(t, &recordingCMTestResolver{err: failure}, cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
		t.Fatal("transport was called after a resolution failure")
		return nil, nil
	}), presenter)

	result, err := module.Run(context.Background(), TestRequest{})
	if !errors.Is(err, failure) || result != (TestResult{}) || presenter.response != "" || presenter.failure != (TestDiagnostic{}) {
		t.Fatalf("Run() = (%#v, %v, %#v)", result, err, presenter)
	}
}

func TestCMTestModuleRedactsTheProfileAPIKeyFromSuccessfulProviderContent(t *testing.T) {
	profile := cmTestProfile()
	presenter := &recordingCMTestPresenter{}
	transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"received test-api-key"}}]}`))}, nil
	})
	module := newCMTestModule(t, &recordingCMTestResolver{profile: profile}, transport, presenter)

	result, err := module.Run(context.Background(), TestRequest{})
	if err != nil || strings.Contains(result.Content, profile.APIKey) || strings.Contains(presenter.response, profile.APIKey) || result.Content != "received [REDACTED]" {
		t.Fatalf("Run() = (%#v, %v, %#v)", result, err, presenter)
	}
}

func TestNewCMTestModuleRequiresEachCommandOwnedAdapter(t *testing.T) {
	resolver := &recordingCMTestResolver{profile: cmTestProfile()}
	transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) { return nil, nil })
	presenter := &recordingCMTestPresenter{}
	for _, dependencies := range []TestDependencies{
		{Transport: transport, Presenter: presenter},
		{Resolver: resolver, Presenter: presenter},
		{Resolver: resolver, Transport: transport},
	} {
		if _, err := NewTest(dependencies); err == nil {
			t.Fatalf("NewTest(%#v) returned nil error", dependencies)
		}
	}
}

func newCMTestModule(t *testing.T, resolver TestProfileResolver, transport cmTestProviderTransport, presenter TestPresenter) *TestModule {
	t.Helper()
	module, err := NewTest(TestDependencies{Resolver: resolver, Transport: transport, Presenter: presenter})
	if err != nil {
		t.Fatalf("NewTest() returned an error: %v", err)
	}
	return module
}

type recordingCMTestResolver struct {
	profile appconfig.ResolvedCMProfile
	options appconfig.CMResolveOptions
	err     error
}

func (resolver *recordingCMTestResolver) ResolveCMProfile(options appconfig.CMResolveOptions) (appconfig.ResolvedCMProfile, error) {
	resolver.options = options
	return resolver.profile, resolver.err
}

type recordingCMTestPresenter struct {
	response string
	failure  TestDiagnostic
}

func (presenter *recordingCMTestPresenter) Response(content string) {
	presenter.response = content
}

func (presenter *recordingCMTestPresenter) Failure(diagnostic TestDiagnostic) {
	presenter.failure = diagnostic
}
