package model

import (
	"time"
)

// Entry holds a value with an optional expiration timestamp.
type Entry[V any] struct {
	Value     V
	ExpiresAt time.Time // zero means no expiration
}

// Persisted is the on-disk format with a version for future migrations.
type Persisted[K comparable, V any] struct {
	Version    int
	DefaultTTL time.Duration
	Data       map[K]Entry[V]
}
