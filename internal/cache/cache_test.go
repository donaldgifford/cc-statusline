package cache

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setCacheHome overrides HOME so cacheDir() resolves to a temp directory.
// On macOS, os.UserCacheDir() returns $HOME/Library/Caches.
// Cannot use t.Parallel() with t.Setenv().
func setCacheHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	// Create the Library/Caches subdirectory for macOS.
	if err := os.MkdirAll(filepath.Join(dir, "Library", "Caches"), dirPerms); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestSetAndGet(t *testing.T) {
	setCacheHome(t)

	data := json.RawMessage(`{"value":"hello"}`)
	if err := Set("test-key", data, 5*time.Minute, time.Time{}); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got := Get("test-key", "")
	if got == nil {
		t.Fatal("Get() returned nil, want cached data")
	}
	if string(got) != string(data) {
		t.Errorf("Get() = %q, want %q", string(got), string(data))
	}
}

func TestGet_CacheMiss(t *testing.T) {
	setCacheHome(t)

	got := Get("nonexistent-key", "")
	if got != nil {
		t.Errorf("Get() = %q, want nil for cache miss", string(got))
	}
}

func TestGet_Expired(t *testing.T) {
	setCacheHome(t)

	data := json.RawMessage(`{"value": "expired"}`)
	// Write with a very short TTL, then verify expiry.
	if err := Set("expire-key", data, 1*time.Nanosecond, time.Time{}); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	// The entry should be expired immediately.
	time.Sleep(1 * time.Millisecond)

	got := Get("expire-key", "")
	if got != nil {
		t.Errorf("Get() = %q, want nil for expired entry", string(got))
	}
}

func TestGet_MtimeInvalidation(t *testing.T) {
	dir := setCacheHome(t)

	// Create a source file.
	sourceFile := filepath.Join(dir, "source.txt")
	if err := os.WriteFile(sourceFile, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(sourceFile)
	if err != nil {
		t.Fatal(err)
	}

	data := json.RawMessage(`{"value": "cached"}`)
	if err := Set("mtime-key", data, 5*time.Minute, info.ModTime()); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Should hit cache since mtime matches.
	got := Get("mtime-key", sourceFile)
	if got == nil {
		t.Fatal("Get() returned nil, want cached data (mtime matches)")
	}

	// Modify the source file to change its mtime.
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(sourceFile, []byte("modified"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Should miss cache since mtime changed.
	got = Get("mtime-key", sourceFile)
	if got != nil {
		t.Errorf("Get() = %q, want nil for changed mtime", string(got))
	}
}

func TestGet_SourceFileDeleted(t *testing.T) {
	setCacheHome(t)

	data := json.RawMessage(`{"value": "cached"}`)
	if err := Set("deleted-key", data, 5*time.Minute, time.Now()); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Source file doesn't exist — should invalidate.
	got := Get("deleted-key", "/nonexistent/file.txt")
	if got != nil {
		t.Errorf("Get() = %q, want nil for missing source file", string(got))
	}
}

func TestGet_NoSourcePath(t *testing.T) {
	setCacheHome(t)

	data := json.RawMessage(`{"value": "no-source"}`)
	if err := Set("nosrc-key", data, 5*time.Minute, time.Time{}); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Empty source path skips mtime check.
	got := Get("nosrc-key", "")
	if got == nil {
		t.Fatal("Get() returned nil, want cached data (no source path)")
	}
}

func TestSet_CreatesDirectory(t *testing.T) {
	setCacheHome(t)

	data := json.RawMessage(`{"value": "first-write"}`)
	if err := Set("newdir-key", data, 5*time.Minute, time.Time{}); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Verify the cache directory was created.
	dir := cacheDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("cache directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("expected directory at %s", dir)
	}
}

func TestAtomicWrite(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	dst := filepath.Join(dir, "atomic.json")

	data := []byte(`{"key": "value"}`)
	if err := atomicWrite(dst, data); err != nil {
		t.Fatalf("atomicWrite() error: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("file contents = %q, want %q", got, data)
	}

	// Verify no temp files remain.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("expected 1 file, got %d", len(entries))
	}
}

func TestGet_CorruptedCache(t *testing.T) {
	setCacheHome(t)

	// Write invalid JSON directly to the cache file.
	cacheFile := cachePath("corrupt-key")
	if err := os.MkdirAll(filepath.Dir(cacheFile), dirPerms); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cacheFile, []byte("not json"), filePerms); err != nil {
		t.Fatal(err)
	}

	got := Get("corrupt-key", "")
	if got != nil {
		t.Errorf("Get() = %q, want nil for corrupted cache", string(got))
	}
}
