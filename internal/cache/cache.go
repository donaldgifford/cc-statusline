// Package cache provides file-based caching with TTL and mtime invalidation.
package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	dirPerms  = 0o700
	filePerms = 0o600
)

type entry struct {
	Data        json.RawMessage `json:"data"`
	ExpiresAt   time.Time       `json:"expires_at"`
	SourceMtime int64           `json:"source_mtime"`
}

// Get reads a cached value. Returns nil if the cache is missing, expired,
// or the source file's mtime has changed since caching.
func Get(key, sourcePath string) json.RawMessage {
	path := cachePath(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var e entry
	if err := json.Unmarshal(data, &e); err != nil {
		return nil
	}

	if time.Now().After(e.ExpiresAt) {
		return nil
	}

	if sourcePath != "" {
		info, err := os.Stat(sourcePath)
		if err != nil {
			return nil
		}
		if info.ModTime().UnixNano() != e.SourceMtime {
			return nil
		}
	}

	return e.Data
}

// Set writes a value to the cache with the given TTL. The source file's
// mtime is recorded for invalidation. Writes atomically via temp file +
// rename.
func Set(key string, data json.RawMessage, ttl time.Duration, sourceMtime time.Time) error {
	dir := cacheDir()
	if err := os.MkdirAll(dir, dirPerms); err != nil {
		return err
	}

	e := entry{
		Data:        data,
		ExpiresAt:   time.Now().Add(ttl),
		SourceMtime: sourceMtime.UnixNano(),
	}

	encoded, err := json.Marshal(e)
	if err != nil {
		return err
	}

	return atomicWrite(cachePath(key), encoded)
}

func atomicWrite(dst string, data []byte) (retErr error) {
	tmp, err := os.CreateTemp(filepath.Dir(dst), "cache-*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	defer func() {
		if retErr != nil {
			os.Remove(tmpName) //nolint:errcheck,gosec // best-effort cleanup on error
		}
	}()

	if _, err := tmp.Write(data); err != nil {
		tmp.Close() //nolint:errcheck,gosec // closing before cleanup removal
		return err
	}

	if err := tmp.Chmod(filePerms); err != nil {
		tmp.Close() //nolint:errcheck,gosec // closing before cleanup removal
		return err
	}

	if err := tmp.Close(); err != nil {
		return err
	}

	return os.Rename(tmpName, dst)
}

func cacheDir() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "cc-statusline")
	}
	return filepath.Join(dir, "cc-statusline")
}

func cachePath(key string) string {
	return filepath.Join(cacheDir(), key+".json")
}
