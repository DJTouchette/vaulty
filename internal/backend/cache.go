package backend

import (
	"sync"
	"time"
)

type cachedEntry struct {
	value     string
	fetchedAt time.Time
}

// CachedBackend wraps a SecretBackend with TTL-based caching.
type CachedBackend struct {
	inner SecretBackend
	cache map[string]cachedEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewCachedBackend creates a CachedBackend wrapping inner with the given TTL.
func NewCachedBackend(inner SecretBackend, ttl time.Duration) *CachedBackend {
	return &CachedBackend{
		inner: inner,
		cache: make(map[string]cachedEntry),
		ttl:   ttl,
	}
}

func (c *CachedBackend) Name() string {
	return c.inner.Name()
}

// List delegates to the inner backend (no caching for list).
func (c *CachedBackend) List() ([]string, error) {
	return c.inner.List()
}

// Get returns the cached value if not expired, otherwise fetches from the inner backend.
func (c *CachedBackend) Get(name string) (string, error) {
	c.mu.RLock()
	entry, ok := c.cache[name]
	c.mu.RUnlock()

	if ok && time.Since(entry.fetchedAt) < c.ttl {
		return entry.value, nil
	}

	val, err := c.inner.Get(name)
	if err != nil {
		return "", err
	}

	c.mu.Lock()
	c.cache[name] = cachedEntry{value: val, fetchedAt: time.Now()}
	c.mu.Unlock()

	return val, nil
}

// Zero zeroes all cached secret values and clears the map.
func (c *CachedBackend) Zero() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, entry := range c.cache {
		// Zero the string bytes via a byte slice copy
		b := []byte(entry.value)
		for i := range b {
			b[i] = 0
		}
		delete(c.cache, k)
	}
}
