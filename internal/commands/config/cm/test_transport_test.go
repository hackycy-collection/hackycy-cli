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

func TestExecuteCMTestProviderUsesTheInjectedLocalTransport(t *testing.T) {
	transport := &recordingCMTestTransport{response: &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(`{"choices":[{"message":{"content":"ok"}}]}`)),
	}}
	result, err := executeCMTestProvider(context.Background(), cmTestProfile(), transport)
	if err != nil {
		t.Fatalf("executeCMTestProvider() returned an error: %v", err)
	}
	if result.Content != "ok" {
		t.Fatalf("result content = %q, want ok", result.Content)
	}
	if transport.request == nil || transport.request.URL.String() != "https://provider.test/v1/chat/completions" {
		t.Fatalf("transport request = %#v", transport.request)
	}
}

func TestExecuteCMTestProviderReportsTimeoutUsingTheEffectiveMilliseconds(t *testing.T) {
	profile := cmTestProfile()
	profile.TimeoutMS = 1.5
	transport := cmTestTransportFunc(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	})

	_, err := executeCMTestProvider(context.Background(), profile, transport)
	if err == nil || err.Error() != "Provider request timed out after 1.5ms" {
		t.Fatalf("executeCMTestProvider() error = %v, want timeout", err)
	}
}

func TestExecuteCMTestProviderRetainsHTTPStatusTextAndBody(t *testing.T) {
	transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusTooManyRequests,
			Status:     "429 Provider Busy",
			Body:       io.NopCloser(strings.NewReader(`{"error":"try later"}`)),
		}, nil
	})

	_, err := executeCMTestProvider(context.Background(), cmTestProfile(), transport)
	if err == nil || err.Error() != `429 Provider Busy: {"error":"try later"}` {
		t.Fatalf("executeCMTestProvider() error = %v, want HTTP status and body", err)
	}
}

func TestExecuteCMTestProviderReturnsResponseReadAndJSONFailures(t *testing.T) {
	t.Run("response read", func(t *testing.T) {
		failure := errors.New("read failed")
		transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: failingCMTestBody{err: failure}}, nil
		})

		_, err := executeCMTestProvider(context.Background(), cmTestProfile(), transport)
		if !errors.Is(err, failure) {
			t.Fatalf("executeCMTestProvider() error = %v, want %v", err, failure)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		transport := cmTestTransportFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"choices":`))}, nil
		})

		_, err := executeCMTestProvider(context.Background(), cmTestProfile(), transport)
		if err == nil || !strings.Contains(err.Error(), "decode CM test provider response") {
			t.Fatalf("executeCMTestProvider() error = %v, want JSON decoding failure", err)
		}
	})
}

func cmTestProfile() appconfig.ResolvedCMProfile {
	return appconfig.ResolvedCMProfile{
		Name:            "work",
		BaseURL:         "https://provider.test/v1",
		Model:           "provider-model",
		APIKey:          "test-api-key",
		Temperature:     0.2,
		TimeoutMS:       300000,
		MaxOutputTokens: 1000,
	}
}

type recordingCMTestTransport struct {
	request  *http.Request
	response *http.Response
	err      error
}

func (transport *recordingCMTestTransport) Do(request *http.Request) (*http.Response, error) {
	transport.request = request
	return transport.response, transport.err
}

type cmTestTransportFunc func(*http.Request) (*http.Response, error)

func (transport cmTestTransportFunc) Do(request *http.Request) (*http.Response, error) {
	return transport(request)
}

type failingCMTestBody struct {
	err error
}

func (body failingCMTestBody) Read([]byte) (int, error) {
	return 0, body.err
}

func (failingCMTestBody) Close() error {
	return nil
}
