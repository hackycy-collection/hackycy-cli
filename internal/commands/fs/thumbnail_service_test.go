package fs

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type thumbnailTestConverter struct {
	mu      sync.Mutex
	calls   int
	started chan struct{}
	release chan struct{}
	output  []byte
}

func (converter *thumbnailTestConverter) convert(_ string, _ []byte) ([]byte, error) {
	converter.mu.Lock()
	converter.calls++
	if converter.started != nil {
		close(converter.started)
		converter.started = nil
	}
	release := converter.release
	output := append([]byte(nil), converter.output...)
	converter.mu.Unlock()
	if release != nil {
		<-release
	}
	return output, nil
}

func TestThumbnailServiceValidatesAndCoalescesStableWorkspaceFile(t *testing.T) {
	root := t.TempDir()
	writeThumbnailServicePNG(t, root, "photo.png")
	workspace, err := OpenWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	defer workspace.Close()
	path, err := ParseWorkspacePath("photo.png")
	if err != nil {
		t.Fatal(err)
	}
	converter := &thumbnailTestConverter{started: make(chan struct{}), release: make(chan struct{}), output: []byte("thumbnail")}
	service := newThumbnailService(workspace, converter)
	defer service.Close()
	started := converter.started
	first := make(chan ThumbnailResult, 1)
	firstErr := make(chan error, 1)
	go func() {
		result, getErr := service.Get(path)
		first <- result
		firstErr <- getErr
	}()
	<-started
	second := make(chan ThumbnailResult, 1)
	secondErr := make(chan error, 1)
	go func() {
		result, getErr := service.Get(path)
		second <- result
		secondErr <- getErr
	}()
	close(converter.release)
	result := <-first
	if err := <-firstErr; err != nil || string(result.Bytes) != "thumbnail" {
		t.Fatalf("first result = %#v, %v", result, err)
	}
	result = <-second
	if err := <-secondErr; err != nil || string(result.Bytes) != "thumbnail" {
		t.Fatalf("second result = %#v, %v", result, err)
	}
	converter.mu.Lock()
	calls := converter.calls
	converter.mu.Unlock()
	if calls != 1 {
		t.Fatalf("conversion calls = %d, want 1", calls)
	}
	result.Bytes[0] = 'x'
	cached, err := service.Get(path)
	if err != nil || string(cached.Bytes) != "thumbnail" {
		t.Fatalf("cached result = %#v, %v", cached, err)
	}
}

func TestThumbnailServiceRejectsUnsupportedAndOversizedWorkspaceFiles(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "image.svg"), []byte("<svg/>"), 0o600); err != nil {
		t.Fatal(err)
	}
	large, err := os.OpenFile(filepath.Join(root, "large.png"), os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := large.Truncate(int64(maxThumbnailSourceBytes + 1)); err != nil {
		t.Fatal(err)
	}
	if err := large.Close(); err != nil {
		t.Fatal(err)
	}
	workspace, err := OpenWorkspace(root)
	if err != nil {
		t.Fatal(err)
	}
	defer workspace.Close()
	service := newThumbnailService(workspace, &thumbnailTestConverter{})
	for _, test := range []struct {
		path string
		code string
	}{
		{path: "image.svg", code: "THUMBNAIL_UNSUPPORTED"},
		{path: "large.png", code: "THUMBNAIL_TOO_LARGE"},
	} {
		path, err := ParseWorkspacePath(test.path)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := service.Get(path); !serviceErrorIs(err, test.code) {
			t.Fatalf("Get(%q) error = %v, want %s", test.path, err, test.code)
		}
	}
}

func TestThumbnailServiceEvictsLeastRecentlyUsedEntriesByCountAndBytes(t *testing.T) {
	service := newThumbnailService(nil, &thumbnailTestConverter{})
	for index := range thumbnailCacheEntries + 1 {
		service.storeLocked(string(rune(index)), ThumbnailResult{Bytes: []byte{byte(index)}})
	}
	if service.lru.Len() != thumbnailCacheEntries {
		t.Fatalf("cache entry count = %d", service.lru.Len())
	}
	if _, ok := service.cache[string(rune(0))]; ok {
		t.Fatal("oldest cache entry was retained")
	}
	large := bytes.Repeat([]byte{'x'}, thumbnailCacheBytes/2+1)
	service.storeLocked("first", ThumbnailResult{Bytes: large})
	service.storeLocked("second", ThumbnailResult{Bytes: large})
	if _, ok := service.cache["first"]; ok || service.cacheBytes > thumbnailCacheBytes {
		t.Fatalf("byte-bounded cache = %d bytes with first=%t", service.cacheBytes, ok)
	}
}

func writeThumbnailServicePNG(t *testing.T, root, name string) {
	t.Helper()
	image := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	image.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	var encoded bytes.Buffer
	if err := png.Encode(&encoded, image); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, name), encoded.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
}
