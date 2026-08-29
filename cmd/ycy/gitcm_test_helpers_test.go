package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os/exec"
	"strings"
	"testing"
)

type gitCMProviderFixture struct {
	calls  int
	bodies []string
}

func newGitCMMessageProvider(t *testing.T, message string) (*httptest.Server, *gitCMProviderFixture) {
	t.Helper()
	fixture := &gitCMProviderFixture{}
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		fixture.calls++
		if request.Method != http.MethodPost || request.URL.Path != "/chat/completions" {
			t.Errorf("provider request = %s %s", request.Method, request.URL.Path)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(request.Body)
		if err != nil {
			t.Errorf("read provider request: %v", err)
			response.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.bodies = append(fixture.bodies, string(body))
		response.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(response, `{"choices":[{"message":{"content":%q}}]}`, message)
	}))
	return server, fixture
}

func gitCMOutput(t *testing.T, directory string, arguments ...string) string {
	t.Helper()
	command := exec.Command("git", arguments...)
	command.Dir = directory
	command.Env = environmentWith(map[string]string{"GIT_CONFIG_NOSYSTEM": "1"})
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(arguments, " "), err, output)
	}
	return string(output)
}
