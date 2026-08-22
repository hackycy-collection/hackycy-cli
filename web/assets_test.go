package webassets

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadValidatesFixedShells(t *testing.T) {
	if err := Validate(); err != nil {
		t.Fatalf("Validate returned an error: %v", err)
	}
	if _, err := Load("unknown"); err == nil {
		t.Fatal("Load accepted an unknown application")
	}
}

func TestServeAssetUsesImmutableHeaders(t *testing.T) {
	site, err := Load("fs")
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	asset := firstAsset(t, site.files)
	response := httptest.NewRecorder()
	if !site.ServeAsset(response, httptest.NewRequest(http.MethodGet, "/"+asset, nil)) {
		t.Fatal("ServeAsset did not serve an embedded asset")
	}
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "public, max-age=31536000, immutable" || response.Header().Get("Referrer-Policy") != "no-referrer" || response.Header().Get("X-Content-Type-Options") != "nosniff" || response.Body.Len() == 0 {
		t.Fatalf("unexpected asset response: code=%d headers=%v body=%d", response.Code, response.Header(), response.Body.Len())
	}
	if site.ServeAsset(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/"+asset, nil)) {
		t.Fatal("ServeAsset accepted a non-GET/HEAD request")
	}
	if site.ServeAsset(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)) {
		t.Fatal("ServeAsset accepted a missing asset")
	}
}

func TestServeShellUsesNoStoreHeaders(t *testing.T) {
	site, err := Load("diff")
	if err != nil {
		t.Fatalf("Load returned an error: %v", err)
	}
	response := httptest.NewRecorder()
	site.ServeShell(response, httptest.NewRequest(http.MethodHead, "/", nil), "default-src 'self'")
	if response.Code != http.StatusOK || response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("Content-Security-Policy") != "default-src 'self'" || response.Body.Len() != 0 {
		t.Fatalf("unexpected shell response: code=%d headers=%v body=%d", response.Code, response.Header(), response.Body.Len())
	}
}

func firstAsset(t *testing.T, files fs.FS) string {
	t.Helper()
	entries, err := fs.ReadDir(files, "assets")
	if err != nil {
		t.Fatalf("read assets: %v", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".js") {
			return "assets/" + entry.Name()
		}
	}
	t.Fatal("no JavaScript asset was embedded")
	return ""
}
