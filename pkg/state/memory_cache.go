package state

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/utils"
)

// MemoryCache implements an in-memory result cache with TTL
type MemoryCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	
	// Configuration
	maxEntries      int
	maxMemory       int64
	defaultTTL      time.Duration
	
	// Metrics
	hits            int64
	misses          int64
	evictions       int64
	currentMemory   int64
	
	// Cleanup
	cleanupInterval time.Duration
	stopCleanup     chan struct{}
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(maxEntries int, maxMemory int64, defaultTTL time.Duration) *MemoryCache {
	mc := &MemoryCache{
		entries:         make(map[string]*CacheEntry),
		maxEntries:      maxEntries,
		maxMemory:       maxMemory,
		defaultTTL:      defaultTTL,
		cleanupInterval: 1 * time.Minute,
		stopCleanup:     make(chan struct{}),
	}
	
	// Start cleanup goroutine with panic recovery
	utils.SafeGoWithRestart("MemoryCache.cleanupLoop", mc.cleanupLoop, 3)
	
	return mc
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, bool) {
	mc.mu.RLock()
	entry, exists := mc.entries[key]
	
	if !exists {
		mc.mu.RUnlock()
		atomic.AddInt64(&mc.misses, 1)
		return nil, false
	}
	
	// Check if entry has expired while holding read lock
	if mc.isExpired(entry) {
		mc.mu.RUnlock()
		
		// Acquire write lock to delete expired entry
		mc.mu.Lock()
		// Double-check the entry still exists and is expired
		if e, ok := mc.entries[key]; ok && mc.isExpired(e) {
			delete(mc.entries, key)
			atomic.AddInt64(&mc.currentMemory, -e.Size)
		}
		mc.mu.Unlock()
		
		atomic.AddInt64(&mc.misses, 1)
		return nil, false
	}
	
	// Update access count and return value
	atomic.AddInt64(&entry.AccessCount, 1)
	value := entry.Value
	mc.mu.RUnlock()
	
	atomic.AddInt64(&mc.hits, 1)
	return value, true
}

// Set stores a value in the cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = mc.defaultTTL
	}
	
	// Calculate size before creating entry
	size := mc.estimateSize(value)
	
	entry := &CacheEntry{
		Key:         key,
		Value:       value,
		CreatedAt:   time.Now(),
		TTL:         ttl,
		AccessCount: 0,
		Size:        size,
	}
	
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	// Check if we need to evict entries
	if len(mc.entries) >= mc.maxEntries {
		mc.evictLRU()
	}
	
	// Add or update entry
	if existing, exists := mc.entries[key]; exists {
		// Update existing entry
		atomic.AddInt64(&mc.currentMemory, -existing.Size)
	}
	
	mc.entries[key] = entry
	atomic.AddInt64(&mc.currentMemory, entry.Size)
	
	// Check memory limit
	for atomic.LoadInt64(&mc.currentMemory) > mc.maxMemory && len(mc.entries) > 0 {
		mc.evictLRU()
	}
	
	return nil
}

// Delete removes an entry from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	if entry, exists := mc.entries[key]; exists {
		atomic.AddInt64(&mc.currentMemory, -entry.Size)
		delete(mc.entries, key)
	}
	
	return nil
}

// Clear removes all entries from the cache
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	mc.entries = make(map[string]*CacheEntry)
	atomic.StoreInt64(&mc.currentMemory, 0)
	
	return nil
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats(ctx context.Context) (CacheStats, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	
	return CacheStats{
		TotalEntries: int64(len(mc.entries)),
		MemoryUsage:  atomic.LoadInt64(&mc.currentMemory),
		HitCount:     atomic.LoadInt64(&mc.hits),
		MissCount:    atomic.LoadInt64(&mc.misses),
		EvictCount:   atomic.LoadInt64(&mc.evictions),
	}, nil
}

// isExpired checks if a cache entry has expired
func (mc *MemoryCache) isExpired(entry *CacheEntry) bool {
	return time.Since(entry.CreatedAt) > entry.TTL
}

// evictLRU evicts the least recently used entry (must be called with lock held)
func (mc *MemoryCache) evictLRU() {
	if len(mc.entries) == 0 {
		return
	}
	
	var lruKey string
	var lruEntry *CacheEntry
	var minAccessCount int64 = -1
	
	// Find LRU entry (simple implementation, could be optimized with heap)
	for key, entry := range mc.entries {
		if minAccessCount == -1 || entry.AccessCount < minAccessCount {
			lruKey = key
			lruEntry = entry
			minAccessCount = entry.AccessCount
		}
	}
	
	if lruKey != "" {
		atomic.AddInt64(&mc.currentMemory, -lruEntry.Size)
		delete(mc.entries, lruKey)
		atomic.AddInt64(&mc.evictions, 1)
	}
}

// estimateSize estimates the memory size of a value more accurately
func (mc *MemoryCache) estimateSize(value interface{}) int64 {
	return mc.calculateSize(value, make(map[uintptr]bool))
}

// calculateSize recursively calculates the size of a value
func (mc *MemoryCache) calculateSize(value interface{}, visited map[uintptr]bool) int64 {
	if value == nil {
		return 0
	}

	var size int64 = 0

	switch v := value.(type) {
	case string:
		// String header is 16 bytes + string content
		size = 16 + int64(len(v))
	
	case []byte:
		// Slice header is 24 bytes + byte content
		size = 24 + int64(len(v))
	
	case int, int32, int64, uint, uint32, uint64:
		size = 8
	
	case float32:
		size = 4
	
	case float64:
		size = 8
	
	case bool:
		size = 1
	
	case map[string]interface{}:
		// Map header is about 48 bytes
		size = 48
		for k, v := range v {
			// Add key size
			size += 16 + int64(len(k))
			// Add value size recursively
			size += mc.calculateSize(v, visited)
		}
	
	case []interface{}:
		// Slice header is 24 bytes
		size = 24
		for _, elem := range v {
			size += mc.calculateSize(elem, visited)
		}
	
	case map[string]string:
		// Map header is about 48 bytes
		size = 48
		for k, v := range v {
			// String key and value
			size += 16 + int64(len(k)) + 16 + int64(len(v))
		}
	
	case []string:
		// Slice header is 24 bytes
		size = 24
		for _, s := range v {
			size += 16 + int64(len(s))
		}
	
	default:
		// For unknown types, use a conservative estimate
		// This includes struct overhead
		size = 64
	}

	return size
}

// cleanupLoop runs periodic cleanup of expired entries
func (mc *MemoryCache) cleanupLoop() {
	ticker := time.NewTicker(mc.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mc.cleanupExpired()
		case <-mc.stopCleanup:
			return
		}
	}
}

// cleanupExpired removes all expired entries
func (mc *MemoryCache) cleanupExpired() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	
	now := time.Now()
	for key, entry := range mc.entries {
		if now.Sub(entry.CreatedAt) > entry.TTL {
			atomic.AddInt64(&mc.currentMemory, -entry.Size)
			delete(mc.entries, key)
			atomic.AddInt64(&mc.evictions, 1)
		}
	}
}

// Close stops the cleanup loop
func (mc *MemoryCache) Close() error {
	close(mc.stopCleanup)
	return nil
}

// HitRate returns the cache hit rate
func (mc *MemoryCache) HitRate() float64 {
	hits := atomic.LoadInt64(&mc.hits)
	misses := atomic.LoadInt64(&mc.misses)
	total := hits + misses
	
	if total == 0 {
		return 0
	}
	
	return float64(hits) / float64(total)
}