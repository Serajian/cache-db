package database

import (
	"encoding/gob"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Serajian/cache-db/model"
)

// Database is a generic in-memory KV store with optional per-key TTL and persistence.
type Database[K comparable, V any] struct {
	lock       sync.RWMutex
	data       map[K]model.Entry[V]
	defaultTTL time.Duration
	basePath   string
}

// NewDatabase creates a new database with an optional default TTL.
// If defaultTTL <= 0, inserted keys won't expire unless SetWithTTL is used.
// basePath: dir for store persist
func NewDatabase[K comparable, V any](defaultTTL time.Duration, basePath string) *Database[K, V] {
	return &Database[K, V]{
		data:       make(map[K]model.Entry[V]),
		defaultTTL: defaultTTL,
		basePath:   basePath,
	}
}

// Set inserts or replaces the value for key, applying default TTL if configured.
func (db *Database[K, V]) Set(key K, value V) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var exp time.Time
	if db.defaultTTL > 0 {
		exp = time.Now().Add(db.defaultTTL)
	}

	db.data[key] = model.Entry[V]{Value: value, ExpiresAt: exp}
}

// SetWithTTL inserts or replaces the value for key with a specific TTL.
// If ttl <= 0, the value never expires (overrides defaultTTL).
func (db *Database[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	db.lock.Lock()
	defer db.lock.Unlock()

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	db.data[key] = model.Entry[V]{Value: value, ExpiresAt: exp}
}

// Get returns the value for key. If the key is expired, it is removed and (zero, false) is returned.
func (db *Database[K, V]) Get(key K) (V, bool) {
	db.lock.RLock()
	e, ok := db.data[key]
	db.lock.RUnlock()

	if !ok {
		var zero V
		return zero, false
	}

	// Fast path: not expired
	if e.ExpiresAt.IsZero() || time.Now().Before(e.ExpiresAt) {
		return e.Value, true
	}

	// Expired: upgrade to write lock and delete
	db.lock.Lock()
	defer db.lock.Unlock()
	// Re-check in case of race
	if e2, ok2 := db.data[key]; ok2 {
		if !e2.ExpiresAt.IsZero() && time.Now().After(e2.ExpiresAt) {
			delete(db.data, key)
		} else {
			return e2.Value, true
		}
	}
	var zero V
	return zero, false
}

// Delete removes a key if it exists.
func (db *Database[K, V]) Delete(key K) {
	db.lock.Lock()
	defer db.lock.Unlock()
	delete(db.data, key)
}

// Clear removes all keys immediately.
func (db *Database[K, V]) Clear() {
	db.lock.Lock()
	defer db.lock.Unlock()
	db.data = make(map[K]model.Entry[V])
}

// CleanExpired removes all expired keys. Useful for periodic maintenance.
func (db *Database[K, V]) CleanExpired() int {
	db.lock.Lock()
	defer db.lock.Unlock()

	now := time.Now()

	removed := 0
	for k, e := range db.data {
		if !e.ExpiresAt.IsZero() && now.After(e.ExpiresAt) {
			delete(db.data, k)
			removed++
		}
	}

	return removed
}

// ******* Persist Methods *******

// Persist writes the database atomically to filename (temp file + rename).
// It captures a consistent snapshot under a read lock, then encodes outside the lock.
func (db *Database[K, V]) Persist(filename string) error {
	// Take a snapshot under RLock to minimize blocking writers.
	db.lock.RLock()
	snap := model.Persisted[K, V]{
		Version:    1,
		DefaultTTL: db.defaultTTL,
		Data:       make(map[K]model.Entry[V], len(db.data)),
	}
	for k, v := range db.data {
		snap.Data[k] = v
	}
	db.lock.RUnlock()

	// Ensure directory exists.
	path := db.getPath(filename)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("ensure dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".db-*.gob.tmp")
	if err != nil {
		return fmt.Errorf("create temp: %w", err)
	}

	enc := gob.NewEncoder(tmp)
	if err = enc.Encode(snap); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("encode gob: %w", err)
	}

	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("close temp: %w", err)
	}

	if err = os.Rename(tmp.Name(), path); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("rename temp: %w", err)
	}
	return nil
}

// Load replaces the in-memory state with the contents of filename.
// It decodes into a temporary value first, then swaps under a write lock.
func (db *Database[K, V]) Load(filename string) error {
	path := db.getPath(filename)
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	var p model.Persisted[K, V]
	dec := gob.NewDecoder(f)
	if err = dec.Decode(&p); err != nil {
		return fmt.Errorf("decode gob: %w", err)
	}

	// Swap state under write lock.
	db.lock.Lock()
	db.data = p.Data
	db.defaultTTL = p.DefaultTTL
	db.lock.Unlock()
	return nil
}

// DeleteFile removes the persisted file. It is idempotent (no error if file is missing).
func (db *Database[K, V]) DeleteFile(filename string) error {
	path := db.getPath(filename)
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (db *Database[K, V]) getPath(fileName string) string {
	path := filepath.Join(db.basePath, fileName)
	return filepath.Clean(path)
}
