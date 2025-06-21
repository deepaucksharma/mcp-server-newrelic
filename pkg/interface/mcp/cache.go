package mcp

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, bool)
	
	// Set stores a value in the cache with optional TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	
	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error
	
	// Clear removes all values from the cache
	Clear(ctx context.Context) error
}

// Metrics defines the interface for metrics collection
type Metrics interface {
	// IncrementCounter increments a counter metric
	IncrementCounter(name string, tags map[string]string)
	
	// RecordDuration records a duration metric
	RecordDuration(name string, duration time.Duration, tags map[string]string)
	
	// RecordGauge records a gauge metric
	RecordGauge(name string, value float64, tags map[string]string)
	
	// RecordToolExecution records a tool execution
	RecordToolExecution(toolName string, duration time.Duration, success bool)
}