/**
 * Intelligent Caching System with Freshness Strategies
 * 
 * Implements adaptive caching with context-aware TTL, invalidation strategies,
 * and freshness indicators for optimal performance vs accuracy balance.
 */

import { Logger } from './types.js';

export interface CacheEntry<T> {
  data: T;
  metadata: {
    key: string;
    timestamp: Date;
    ttl: number; // milliseconds
    accessCount: number;
    lastAccessed: Date;
    dataHash?: string; // For change detection
    freshness: 'fresh' | 'recent' | 'stale' | 'expired';
    source: 'primary' | 'fallback' | 'computed';
  };
  strategy: CacheStrategy;
}

export interface CacheStrategy {
  name: string;
  ttl: number; // base TTL in milliseconds
  adaptiveTtl: boolean; // Whether to adjust TTL based on access patterns
  maxAge: number; // absolute maximum age before forced refresh
  refreshThreshold: number; // 0-1, when to trigger background refresh
  priority: 'low' | 'medium' | 'high' | 'critical';
}

export interface CacheStats {
  totalEntries: number;
  hitRate: number;
  missRate: number;
  avgResponseTime: number;
  memoryUsage: number;
  oldestEntry: Date;
  mostAccessed: string;
}

export interface FreshnessPolicy {
  discovery: CacheStrategy;
  goldenMetrics: CacheStrategy;
  entityDetails: CacheStrategy;
  dashboards: CacheStrategy;
  analytics: CacheStrategy;
}

export class IntelligentCache {
  private cache = new Map<string, CacheEntry<any>>();
  private stats = {
    hits: 0,
    misses: 0,
    totalRequests: 0,
    responseTimes: [] as number[],
  };

  private readonly strategies: FreshnessPolicy = {
    discovery: {
      name: 'discovery',
      ttl: 5 * 60 * 1000, // 5 minutes base
      adaptiveTtl: true,
      maxAge: 30 * 60 * 1000, // 30 minutes max
      refreshThreshold: 0.8, // Refresh when 80% of TTL elapsed
      priority: 'high',
    },
    goldenMetrics: {
      name: 'goldenMetrics',
      ttl: 2 * 60 * 1000, // 2 minutes base
      adaptiveTtl: true,
      maxAge: 10 * 60 * 1000, // 10 minutes max
      refreshThreshold: 0.7, // Refresh when 70% of TTL elapsed
      priority: 'critical',
    },
    entityDetails: {
      name: 'entityDetails',
      ttl: 10 * 60 * 1000, // 10 minutes base
      adaptiveTtl: false,
      maxAge: 60 * 60 * 1000, // 1 hour max
      refreshThreshold: 0.9, // Refresh when 90% of TTL elapsed
      priority: 'medium',
    },
    dashboards: {
      name: 'dashboards',
      ttl: 15 * 60 * 1000, // 15 minutes base
      adaptiveTtl: false,
      maxAge: 4 * 60 * 60 * 1000, // 4 hours max
      refreshThreshold: 0.8,
      priority: 'low',
    },
    analytics: {
      name: 'analytics',
      ttl: 30 * 60 * 1000, // 30 minutes base
      adaptiveTtl: true,
      maxAge: 2 * 60 * 60 * 1000, // 2 hours max
      refreshThreshold: 0.75,
      priority: 'medium',
    },
  };

  constructor(private logger: Logger) {
    // Periodic cleanup every 5 minutes
    setInterval(() => this.performMaintenance(), 5 * 60 * 1000);
  }

  /**
   * Get cached data with intelligent freshness assessment
   */
  async get<T>(
    key: string,
    strategyType: keyof FreshnessPolicy,
    dataFetcher?: () => Promise<T>,
    forceRefresh: boolean = false
  ): Promise<{ data: T | null; cached: boolean; freshness: string }> {
    const startTime = Date.now();
    this.stats.totalRequests++;

    const strategy = this.strategies[strategyType];
    const entry = this.cache.get(key);

    // Handle force refresh
    if (forceRefresh && dataFetcher) {
      this.logger.debug('Force refresh requested', { key });
      const freshData = await this.fetchAndCache(key, strategy, dataFetcher);
      return {
        data: freshData,
        cached: false,
        freshness: 'fresh',
      };
    }

    // Check if we have cached data
    if (entry) {
      const freshness = this.assessFreshness(entry, strategy);
      entry.metadata.accessCount++;
      entry.metadata.lastAccessed = new Date();

      // Record hit
      this.stats.hits++;
      this.recordResponseTime(Date.now() - startTime);

      // Return cached data if acceptable freshness
      if (freshness !== 'expired') {
        // Trigger background refresh if needed
        if (dataFetcher && this.shouldBackgroundRefresh(entry, strategy)) {
          this.backgroundRefresh(key, strategy, dataFetcher);
        }

        return {
          data: entry.data,
          cached: true,
          freshness,
        };
      }
    }

    // Cache miss or expired - fetch fresh data
    this.stats.misses++;
    
    if (dataFetcher) {
      const freshData = await this.fetchAndCache(key, strategy, dataFetcher);
      this.recordResponseTime(Date.now() - startTime);
      return {
        data: freshData,
        cached: false,
        freshness: 'fresh',
      };
    }

    // No data fetcher provided and no valid cache
    this.recordResponseTime(Date.now() - startTime);
    return {
      data: null,
      cached: false,
      freshness: 'expired',
    };
  }

  /**
   * Set cached data with strategy-specific settings
   */
  set<T>(
    key: string,
    data: T,
    strategyType: keyof FreshnessPolicy,
    customTtl?: number
  ): void {
    const strategy = this.strategies[strategyType];
    const ttl = customTtl || this.calculateAdaptiveTtl(key, strategy);

    const entry: CacheEntry<T> = {
      data,
      metadata: {
        key,
        timestamp: new Date(),
        ttl,
        accessCount: 1,
        lastAccessed: new Date(),
        dataHash: this.calculateDataHash(data),
        freshness: 'fresh',
        source: 'primary',
      },
      strategy,
    };

    this.cache.set(key, entry);
    this.logger.debug('Data cached', { key, ttl, strategy: strategy.name });
  }

  /**
   * Invalidate specific cache entries
   */
  invalidate(pattern: string | RegExp): number {
    let invalidated = 0;
    
    for (const [key] of this.cache) {
      const matches = typeof pattern === 'string' 
        ? key.includes(pattern)
        : pattern.test(key);
      
      if (matches) {
        this.cache.delete(key);
        invalidated++;
      }
    }

    this.logger.info('Cache invalidation completed', { pattern, invalidated });
    return invalidated;
  }

  /**
   * Get cache statistics
   */
  getStats(): CacheStats {
    const entries = Array.from(this.cache.values());
    const hitRate = this.stats.totalRequests > 0 
      ? this.stats.hits / this.stats.totalRequests 
      : 0;
    
    const avgResponseTime = this.stats.responseTimes.length > 0
      ? this.stats.responseTimes.reduce((sum, time) => sum + time, 0) / this.stats.responseTimes.length
      : 0;

    const memoryUsage = this.estimateMemoryUsage();
    
    const oldestEntry = entries.length > 0
      ? new Date(Math.min(...entries.map(e => e.metadata.timestamp.getTime())))
      : new Date();

    const mostAccessed = entries.length > 0
      ? entries.reduce((max, entry) => 
          entry.metadata.accessCount > max.metadata.accessCount ? entry : max
        ).metadata.key
      : '';

    return {
      totalEntries: this.cache.size,
      hitRate: Math.round(hitRate * 100) / 100,
      missRate: Math.round((1 - hitRate) * 100) / 100,
      avgResponseTime: Math.round(avgResponseTime),
      memoryUsage,
      oldestEntry,
      mostAccessed,
    };
  }

  /**
   * Get cache health assessment
   */
  getHealthAssessment(): {
    status: 'healthy' | 'warning' | 'critical';
    issues: string[];
    recommendations: string[];
  } {
    const stats = this.getStats();
    const issues: string[] = [];
    const recommendations: string[] = [];

    // Check hit rate
    if (stats.hitRate < 0.3) {
      issues.push(`Low cache hit rate: ${(stats.hitRate * 100).toFixed(1)}%`);
      recommendations.push('Consider increasing TTL values or reviewing cache strategies');
    }

    // Check memory usage
    if (stats.memoryUsage > 100 * 1024 * 1024) { // 100MB
      issues.push(`High memory usage: ${Math.round(stats.memoryUsage / 1024 / 1024)}MB`);
      recommendations.push('Implement more aggressive cache eviction or reduce TTL');
    }

    // Check response time
    if (stats.avgResponseTime > 1000) {
      issues.push(`High average response time: ${stats.avgResponseTime}ms`);
      recommendations.push('Optimize data fetching or implement background refresh');
    }

    // Check cache size
    if (stats.totalEntries > 1000) {
      issues.push(`Large cache size: ${stats.totalEntries} entries`);
      recommendations.push('Implement LRU eviction or reduce cache scope');
    }

    const status = issues.length === 0 ? 'healthy' 
                 : issues.length <= 2 ? 'warning' 
                 : 'critical';

    return { status, issues, recommendations };
  }

  /**
   * Private methods for cache management
   */

  private async fetchAndCache<T>(
    key: string,
    strategy: CacheStrategy,
    dataFetcher: () => Promise<T>
  ): Promise<T> {
    try {
      const data = await dataFetcher();
      this.set(key, data, strategy.name as keyof FreshnessPolicy);
      return data;
    } catch (error: any) {
      this.logger.error('Failed to fetch data for cache', { key, error: error.message });
      
      // Try to return stale data if available
      const staleEntry = this.cache.get(key);
      if (staleEntry) {
        this.logger.warn('Returning stale data due to fetch failure', { key });
        staleEntry.metadata.source = 'fallback';
        return staleEntry.data;
      }
      
      throw error;
    }
  }

  private assessFreshness(entry: CacheEntry<any>, strategy: CacheStrategy): string {
    const now = Date.now();
    const age = now - entry.metadata.timestamp.getTime();
    
    if (age > strategy.maxAge) {
      entry.metadata.freshness = 'expired';
      return 'expired';
    }
    
    if (age > entry.metadata.ttl) {
      entry.metadata.freshness = 'stale';
      return 'stale';
    }
    
    if (age > entry.metadata.ttl * 0.7) {
      entry.metadata.freshness = 'recent';
      return 'recent';
    }
    
    entry.metadata.freshness = 'fresh';
    return 'fresh';
  }

  private shouldBackgroundRefresh(entry: CacheEntry<any>, strategy: CacheStrategy): boolean {
    const age = Date.now() - entry.metadata.timestamp.getTime();
    const refreshPoint = entry.metadata.ttl * strategy.refreshThreshold;
    
    return age >= refreshPoint && strategy.priority !== 'low';
  }

  private async backgroundRefresh<T>(
    key: string,
    strategy: CacheStrategy,
    dataFetcher: () => Promise<T>
  ): Promise<void> {
    try {
      this.logger.debug('Background refresh triggered', { key });
      const freshData = await dataFetcher();
      this.set(key, freshData, strategy.name as keyof FreshnessPolicy);
    } catch (error: any) {
      this.logger.warn('Background refresh failed', { key, error: error.message });
      // Don't throw - this is background operation
    }
  }

  private calculateAdaptiveTtl(key: string, strategy: CacheStrategy): number {
    if (!strategy.adaptiveTtl) {
      return strategy.ttl;
    }

    const entry = this.cache.get(key);
    if (!entry) {
      return strategy.ttl;
    }

    // Adjust TTL based on access frequency
    const accessFrequency = entry.metadata.accessCount / 
      Math.max(1, (Date.now() - entry.metadata.timestamp.getTime()) / 1000 / 60); // per minute

    // High access frequency = longer TTL (up to 2x)
    // Low access frequency = shorter TTL (down to 0.5x)
    const multiplier = Math.min(2, Math.max(0.5, 1 + (accessFrequency - 1) * 0.2));
    
    return Math.round(strategy.ttl * multiplier);
  }

  private calculateDataHash(data: any): string {
    // Simple hash function for change detection
    const str = JSON.stringify(data);
    let hash = 0;
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // Convert to 32-bit integer
    }
    return hash.toString(36);
  }

  private recordResponseTime(time: number): void {
    this.stats.responseTimes.push(time);
    
    // Keep only last 100 response times
    if (this.stats.responseTimes.length > 100) {
      this.stats.responseTimes = this.stats.responseTimes.slice(-100);
    }
  }

  private estimateMemoryUsage(): number {
    let totalSize = 0;
    
    for (const [key, entry] of this.cache) {
      // Rough estimation
      totalSize += key.length * 2; // UTF-16 characters
      totalSize += JSON.stringify(entry).length * 2;
    }
    
    return totalSize;
  }

  private performMaintenance(): void {
    const startSize = this.cache.size;
    let cleaned = 0;

    // Remove expired entries
    for (const [key, entry] of this.cache) {
      const age = Date.now() - entry.metadata.timestamp.getTime();
      
      if (age > entry.strategy.maxAge) {
        this.cache.delete(key);
        cleaned++;
      }
    }

    // Implement LRU eviction if cache is still too large
    if (this.cache.size > 500) { // Max 500 entries
      const entries = Array.from(this.cache.entries())
        .sort(([, a], [, b]) => a.metadata.lastAccessed.getTime() - b.metadata.lastAccessed.getTime());
      
      const toRemove = entries.slice(0, this.cache.size - 400); // Keep 400 most recent
      
      for (const [key] of toRemove) {
        this.cache.delete(key);
        cleaned++;
      }
    }

    if (cleaned > 0) {
      this.logger.info('Cache maintenance completed', { 
        startSize, 
        endSize: this.cache.size, 
        cleaned 
      });
    }
  }

  /**
   * Preload cache with anticipated data
   */
  async preload<T>(
    key: string,
    strategyType: keyof FreshnessPolicy,
    dataFetcher: () => Promise<T>
  ): Promise<void> {
    try {
      const strategy = this.strategies[strategyType];
      await this.fetchAndCache(key, strategy, dataFetcher);
      this.logger.debug('Cache preloaded', { key, strategy: strategy.name });
    } catch (error: any) {
      this.logger.warn('Cache preload failed', { key, error: error.message });
    }
  }

  /**
   * Warm up cache with common queries
   */
  async warmUp(accountId: number, commonEntityGuids: string[]): Promise<void> {
    this.logger.info('Starting cache warm-up', { accountId, entities: commonEntityGuids.length });

    const warmUpTasks = [
      // Preload discovery data
      this.preload(`discovery:${accountId}`, 'discovery', async () => {
        // This would call the discovery service
        return { warmedUp: true, timestamp: new Date() };
      }),
    ];

    // Preload common entity data
    for (const guid of commonEntityGuids.slice(0, 10)) { // Limit to 10 entities
      warmUpTasks.push(
        this.preload(`entity:${guid}`, 'entityDetails', async () => {
          return { guid, warmedUp: true };
        }),
        this.preload(`metrics:${guid}`, 'goldenMetrics', async () => {
          return { guid, metrics: {}, warmedUp: true };
        })
      );
    }

    await Promise.allSettled(warmUpTasks);
    this.logger.info('Cache warm-up completed');
  }

  /**
   * Clear all cache entries
   */
  clear(): void {
    const size = this.cache.size;
    this.cache.clear();
    this.stats = { hits: 0, misses: 0, totalRequests: 0, responseTimes: [] };
    this.logger.info('Cache cleared', { entriesRemoved: size });
  }
}