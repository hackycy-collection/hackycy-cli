package main

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func standaloneBinaryOutputPath(binary string) string {
	if runtime.GOOS == "windows" && filepath.Ext(binary) == "" {
		return binary + ".exe"
	}
	return binary
}

// resolveStandaloneBinary accepts both the explicit Windows output and legacy test paths.
func resolveStandaloneBinary(binary string) string {
	if _, err := os.Stat(binary); err == nil {
		return binary
	}
	if runtime.GOOS != "windows" || filepath.Ext(binary) != "" {
		return binary
	}
	withSuffix := binary + ".exe"
	if _, err := os.Stat(withSuffix); err == nil {
		return withSuffix
	}
	return binary
}

func TestResolveStandaloneBinaryUsesWindowsExecutableSuffix(t *testing.T) {
	directory := t.TempDir()
	requested := filepath.Join(directory, "ycy")
	actual := standaloneBinaryOutputPath(requested)
	if err := os.WriteFile(actual, []byte("fixture"), 0o600); err != nil {
		t.Fatalf("write executable fixture: %v", err)
	}
	if got := resolveStandaloneBinary(actual); got != actual {
		t.Fatalf("resolveStandaloneBinary(%q) = %q, want %q", actual, got, actual)
	}
	if got := standaloneBinaryOutputPath(requested); got != actual {
		t.Fatalf("standaloneBinaryOutputPath(%q) = %q, want %q", requested, got, actual)
	}
}
