# State Management Architecture

This document specifies the state management and caching architecture of the New Relic MCP Server.

## Table of Contents

1. [Overview](#overview)
2. [State Categories](#state-categories)
3. [Storage Backends](#storage-backends)
4. [Cache Architecture](#cache-architecture)
5. [Session Management](#session-management)
6. [State Synchronization](#state-synchronization)
7. [Performance Optimization](#performance-optimization)
8. [Failure Handling](#failure-handling)
9. [Security Considerations](#security-considerations)
10. [Future Enhancements](#future-enhancements)

## Overview

The State Management system SHALL provide efficient, scalable, and reliable state storage for the MCP Server across multiple deployment scenarios.

### Design Goals

1. **Performance**: Sub-millisecond access for hot data
2. **Scalability**: Support from single instance to global deployment
3. **Reliability**: Graceful degradation and recovery
4. **Flexibility**: Multiple storage backend support
5. **Security**: Encrypted sensitive state

### State Lifecycle

```
Create → Store → Access → Update → Expire → Delete
   ↓        ↓       ↓        ↓        ↓       ↓
Validate  Cache  Authorize  Lock   Notify  Cleanup
```

## State Categories

### 1. Request State
**Lifetime**: Single request
**Scope**: Request context
**Storage**: In-memory

```go
type RequestState struct {
    RequestID    string
    UserID       string
    StartTime    time.Time
    TraceContext map[string]string
    Cache        map[string]interface{}
}
```

### 2. Session State
**Lifetime**: User session (30 minutes default)
**Scope**: User interactions
**Storage**: Memory + Redis

```go
type SessionState struct {
    SessionID     string
    UserID        string
    Discoveries   map[string]Discovery
    Preferences   UserPreferences
    History       []HistoryEntry
    LastActivity  time.Time
}
```

### 3. Discovery State
**Lifetime**: Hours to days
**Scope**: Account-wide
**Storage**: Distributed cache

```go
type DiscoveryState struct {
    AccountID    string
    EventTypes   []EventType
    Schemas      map[string]Schema
    Profiles     map[string]Profile
    LastUpdated  time.Time
    Version      int
}
```

### 4. Application State
**Lifetime**: Application lifecycle
**Scope**: Server instance
**Storage**: Memory

```go
type ApplicationState struct {
    ServerID     string
    StartTime    time.Time
    Connections  int
    ToolRegistry map[string]Tool
    Config       Configuration
}
```

## Storage Backends

### Memory Store (Default)

**Implementation**: In-process concurrent map
**Use Case**: Single instance, development
**Characteristics**:
- Zero network latency
- No persistence
- Limited by process memory
- Lost on restart

```go
type MemoryStore struct {
    data sync.Map
    ttls map[string]time.Time
    mu   sync.RWMutex
}

func (m *MemoryStore) Get(key string) (interface{}, error) {
    if m.isExpired(key) {
        m.Delete(key)
        return nil, ErrNotFound
    }
    value, ok := m.data.Load(key)
    if !ok {
        return nil, ErrNotFound
    }
    return value, nil
}
```

### Redis Store

**Implementation**: Redis client with connection pooling
**Use Case**: Production, multi-instance
**Characteristics**:
- Network latency (< 1ms typical)
- Persistent with configurable durability
- Distributed access
- Horizontal scalability

```go
type RedisStore struct {
    client  *redis.Client
    prefix  string
    codec   Codec
}

func (r *RedisStore) Get(key string) (interface{}, error) {
    data, err := r.client.Get(r.prefix + key).Bytes()
    if err != nil {
        return nil, err
    }
    return r.codec.Decode(data)
}
```

### Hybrid Store

**Implementation**: Memory + Redis with write-through
**Use Case**: High performance with durability
**Characteristics**:
- Best of both worlds
- Automatic failover
- Configurable consistency

## Cache Architecture

### Multi-Layer Cache Design

```
┌─────────────────────────────────────────┐
│          L1: Request Cache              │
│         (Thread-local, < 1μs)           │
├─────────────────────────────────────────┤
│          L2: Process Cache              │
│         (In-memory, < 100μs)            │
├─────────────────────────────────────────┤
│        L3: Distributed Cache            │
│          (Redis, < 1ms)                 │
├─────────────────────────────────────────┤
│        L4: Persistent Store             │
│      (Database/S3, > 10ms)              │
└─────────────────────────────────────────┘
```

### Cache Strategy

#### Read Path
```
1. Check L1 (Request Cache)
2. Check L2 (Process Cache)  
3. Check L3 (Redis)
4. Load from source
5. Populate L3, L2, L1
```

#### Write Path
```
1. Write to source
2. Invalidate/Update L1
3. Invalidate/Update L2
4. Invalidate/Update L3
```

### Cache Configuration

```yaml
cache:
  request:
    enabled: true
    max_size: 100MB
    
  process:
    enabled: true
    max_size: 1GB
    ttl: 300s
    eviction: lru
    
  distributed:
    enabled: true
    backend: redis
    ttl: 3600s
    compression: true
```

### Cache Patterns

#### Pattern 1: Cache-Aside
```go
func GetWithCache(key string) (interface{}, error) {
    // Try cache first
    if value, err := cache.Get(key); err == nil {
        return value, nil
    }
    
    // Load from source
    value, err := loadFromSource(key)
    if err != nil {
        return nil, err
    }
    
    // Update cache
    cache.Set(key, value, ttl)
    return value, nil
}
```

#### Pattern 2: Write-Through
```go
func SetWithCache(key string, value interface{}) error {
    // Write to source
    if err := writeToSource(key, value); err != nil {
        return err
    }
    
    // Update cache
    cache.Set(key, value, ttl)
    return nil
}
```

#### Pattern 3: Refresh-Ahead
```go
func RefreshAhead(key string) {
    if cache.TTL(key) < refreshThreshold {
        go func() {
            value, _ := loadFromSource(key)
            cache.Set(key, value, ttl)
        }()
    }
}
```

## Session Management

### Session Lifecycle

```
Create Session
     │
     ▼
Authenticate ──► FAIL ──► Reject
     │
  SUCCESS
     │
     ▼
Initialize State
     │
     ▼
Process Requests ◄─┐
     │             │
     ▼             │
Update Activity ───┘
     │
     ▼
Expire/Logout
     │
     ▼
Cleanup State
```

### Session Implementation

```go
type SessionManager struct {
    store       StateStore
    ttl         time.Duration
    maxSessions int
}

type Session struct {
    ID           string
    UserID       string
    CreatedAt    time.Time
    LastActivity time.Time
    Data         map[string]interface{}
    mu           sync.RWMutex
}

func (sm *SessionManager) Create(userID string) (*Session, error) {
    session := &Session{
        ID:           generateID(),
        UserID:       userID,
        CreatedAt:    time.Now(),
        LastActivity: time.Now(),
        Data:         make(map[string]interface{}),
    }
    
    key := fmt.Sprintf("session:%s", session.ID)
    if err := sm.store.Set(key, session, sm.ttl); err != nil {
        return nil, err
    }
    
    return session, nil
}
```

### Session Features

1. **Activity Tracking**: Update last activity on each request
2. **Sliding Expiration**: Extend TTL on activity
3. **Concurrent Access**: Thread-safe operations
4. **Partial Updates**: Update specific session data
5. **Bulk Operations**: List/cleanup multiple sessions

## State Synchronization

### Distributed State Challenges

1. **Consistency**: Eventual vs strong consistency
2. **Partitioning**: Network splits
3. **Clock Skew**: Time synchronization
4. **Race Conditions**: Concurrent updates

### Synchronization Patterns

#### Pattern 1: Optimistic Locking
```go
type VersionedState struct {
    Data    interface{}
    Version int64
}

func UpdateOptimistic(key string, updater func(interface{}) interface{}) error {
    for retries := 0; retries < maxRetries; retries++ {
        // Get current state with version
        current, err := store.GetVersioned(key)
        if err != nil {
            return err
        }
        
        // Apply update
        newData := updater(current.Data)
        
        // Try to save with version check
        if err := store.SetVersioned(key, newData, current.Version); err == nil {
            return nil
        }
    }
    return ErrConcurrentUpdate
}
```

#### Pattern 2: Distributed Locks
```go
func UpdateWithLock(key string, updater func(interface{}) interface{}) error {
    lock := dlock.New(key, lockTTL)
    if err := lock.Lock(); err != nil {
        return err
    }
    defer lock.Unlock()
    
    current, _ := store.Get(key)
    newData := updater(current)
    return store.Set(key, newData, ttl)
}
```

#### Pattern 3: Event Sourcing
```go
type StateEvent struct {
    ID        string
    Type      string
    Data      interface{}
    Timestamp time.Time
}

func ApplyEvents(state interface{}, events []StateEvent) interface{} {
    for _, event := range events {
        state = applyEvent(state, event)
    }
    return state
}
```

## Performance Optimization

### Memory Management

1. **Bounded Caches**: Prevent unbounded growth
```go
type BoundedCache struct {
    data     map[string]CacheEntry
    maxSize  int64
    currSize int64
    eviction EvictionPolicy
}
```

2. **Object Pooling**: Reuse frequently allocated objects
```go
var statePool = sync.Pool{
    New: func() interface{} {
        return &RequestState{
            Cache: make(map[string]interface{}),
        }
    },
}
```

3. **Compression**: Reduce memory footprint
```go
func (s *Store) Set(key string, value interface{}) error {
    data, _ := json.Marshal(value)
    compressed := compress(data)
    return s.backend.Set(key, compressed)
}
```

### Access Patterns

1. **Batch Operations**: Reduce round trips
```go
func GetMulti(keys []string) (map[string]interface{}, error) {
    return store.MGet(keys...)
}
```

2. **Pipeline Commands**: For Redis backend
```go
pipe := redis.Pipeline()
for _, key := range keys {
    pipe.Get(key)
}
results, _ := pipe.Exec()
```

3. **Preloading**: Warm caches proactively
```go
func PreloadDiscoveryCache(accountID string) {
    go func() {
        schemas := discoverSchemas(accountID)
        for _, schema := range schemas {
            cache.Set(schemaKey(schema), schema, longTTL)
        }
    }()
}
```

## Failure Handling

### Failure Modes

1. **Cache Miss**: Fallback to source
2. **Store Unavailable**: Use degraded mode
3. **Network Partition**: Local state only
4. **Data Corruption**: Validate and rebuild

### Resilience Patterns

#### Circuit Breaker
```go
type CircuitBreaker struct {
    failures    int
    threshold   int
    timeout     time.Duration
    lastFailure time.Time
    state       State
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == Open {
        if time.Since(cb.lastFailure) > cb.timeout {
            cb.state = HalfOpen
        } else {
            return ErrCircuitOpen
        }
    }
    
    err := fn()
    if err != nil {
        cb.recordFailure()
    } else {
        cb.reset()
    }
    return err
}
```

#### Fallback Chain
```go
func GetWithFallback(key string) (interface{}, error) {
    // Try primary store
    if value, err := primary.Get(key); err == nil {
        return value, nil
    }
    
    // Try secondary store
    if value, err := secondary.Get(key); err == nil {
        return value, nil
    }
    
    // Use default
    return getDefault(key), nil
}
```

## Security Considerations

### Data Protection

1. **Encryption at Rest**: For sensitive state
```go
func (s *SecureStore) Set(key string, value interface{}) error {
    data, _ := json.Marshal(value)
    encrypted := encrypt(data, s.key)
    return s.backend.Set(key, encrypted)
}
```

2. **Access Control**: Per-key permissions
3. **Audit Logging**: Track state access
4. **Data Masking**: Hide sensitive fields

### Security Patterns

#### Secure Session Storage
```go
type SecureSession struct {
    ID        string
    UserID    string
    Token     string // Encrypted
    ExpiresAt time.Time
}

func (s *SecureSession) Validate() error {
    if time.Now().After(s.ExpiresAt) {
        return ErrSessionExpired
    }
    if !validateToken(s.Token) {
        return ErrInvalidToken
    }
    return nil
}
```

#### State Isolation
- **Tenant Isolation**: Separate state by account
- **User Isolation**: Prevent cross-user access
- **Request Isolation**: No request state leakage

## Future Enhancements

### Advanced Caching

1. **Predictive Caching**: ML-based cache warming
2. **Adaptive TTL**: Dynamic expiration based on access patterns
3. **Geo-distributed Cache**: Regional cache nodes
4. **Cache Analytics**: Usage patterns and optimization

### State Persistence

1. **Event Sourcing**: Complete state history
2. **Snapshotting**: Periodic state backups
3. **State Migration**: Zero-downtime upgrades
4. **State Replication**: Cross-region replication

### Performance Features

1. **NUMA Awareness**: CPU-local caches
2. **Zero-copy Operations**: Avoid data duplication
3. **Lock-free Structures**: Concurrent data structures
4. **Compression Algorithms**: Adaptive compression

### Monitoring and Observability

```yaml
metrics:
  state_operations_total: Counter by operation, status
  state_operation_duration: Histogram by operation
  cache_hit_ratio: Gauge by cache_level
  state_size_bytes: Gauge by state_type
  session_count: Gauge
  lock_contention: Histogram
```

## Best Practices

1. **Use Appropriate TTLs**: Balance freshness vs performance
2. **Monitor Cache Hit Ratios**: Aim for > 80% hit rate
3. **Implement Graceful Degradation**: Handle cache failures
4. **Regular Cleanup**: Prevent state accumulation
5. **Profile Memory Usage**: Identify leaks early
6. **Test Failure Scenarios**: Ensure resilience

## Conclusion

The state management architecture provides a flexible, performant, and reliable foundation for the MCP Server. It supports various deployment scenarios from single-instance development to globally distributed production systems.

## Related Documentation

- [Architecture Overview](10_ARCHITECTURE_OVERVIEW.md) - System architecture
- [Discovery-First Architecture](11_ARCHITECTURE_DISCOVERY_FIRST.md) - Discovery caching
- [Performance Tuning](15_ARCHITECTURE_SCALABILITY.md) - Scaling strategies
- [Security Architecture](14_ARCHITECTURE_SECURITY.md) - Security details

---

**Implementation Note**: The current implementation uses a hybrid approach with memory caching and optional Redis support. See the implementation in `pkg/state/` for details.