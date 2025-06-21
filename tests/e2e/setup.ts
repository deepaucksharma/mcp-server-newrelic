/**
 * E2E Test Setup - Per Test File
 * 
 * Configures individual test files for E2E testing
 */

import { beforeAll, afterAll } from 'vitest';

// Global test configuration
declare global {
  var E2E_TEST_CONFIG: {
    timeout: number;
    parallelAccounts: number;
    cacheTimeoutSeconds: number;
    enableDebugLogging: boolean;
  };
}

beforeAll(() => {
  // Set global test configuration
  globalThis.E2E_TEST_CONFIG = {
    timeout: parseInt(process.env.E2E_TIMEOUT_MS || '120000'),
    parallelAccounts: parseInt(process.env.E2E_PARALLEL_ACCOUNTS || '3'),
    cacheTimeoutSeconds: parseInt(process.env.E2E_CACHE_TTL_SECONDS || '3600'),
    enableDebugLogging: process.env.DEBUG?.includes('mcp') || false,
  };
  
  // Configure console output for E2E tests
  if (globalThis.E2E_TEST_CONFIG.enableDebugLogging) {
    console.log('🔍 Debug logging enabled for E2E tests');
  }
});

afterAll(() => {
  // Cleanup global configuration
  if (globalThis.E2E_TEST_CONFIG) {
    delete globalThis.E2E_TEST_CONFIG;
  }
});

// Utility functions for E2E tests
export function getTestTimeout(): number {
  return globalThis.E2E_TEST_CONFIG?.timeout || 120000;
}

export function getParallelAccountLimit(): number {
  return globalThis.E2E_TEST_CONFIG?.parallelAccounts || 3;
}

export function getCacheTimeout(): number {
  return globalThis.E2E_TEST_CONFIG?.cacheTimeoutSeconds || 3600;
}

export function isDebugEnabled(): boolean {
  return globalThis.E2E_TEST_CONFIG?.enableDebugLogging || false;
}

// Test environment validation helpers
export function requireEnvVar(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Required environment variable ${name} is not set`);
  }
  return value;
}

export function getOptionalEnvVar(name: string, defaultValue?: string): string | undefined {
  return process.env[name] || defaultValue;
}

export function validateAccountId(accountId: string | number): number {
  const id = typeof accountId === 'string' ? parseInt(accountId) : accountId;
  if (isNaN(id) || id <= 0) {
    throw new Error(`Invalid account ID: ${accountId}`);
  }
  return id;
}

// Test result helpers
export function logTestResult(testName: string, success: boolean, details?: any): void {
  const status = success ? '✅' : '❌';
  console.log(`${status} ${testName}`);
  
  if (details && globalThis.E2E_TEST_CONFIG?.enableDebugLogging) {
    console.log('   Details:', JSON.stringify(details, null, 2));
  }
}

export function formatLatency(ms: number): string {
  if (ms < 1000) {
    return `${Math.round(ms)}ms`;
  } else if (ms < 60000) {
    return `${(ms / 1000).toFixed(1)}s`;
  } else {
    return `${(ms / 60000).toFixed(1)}m`;
  }
}

export function formatBytes(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes}B`;
  } else if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)}KB`;
  } else {
    return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
  }
}

// Mock data helpers for testing fallback scenarios
export function createMockEventType(name: string, sampleCount: number = 1000) {
  return {
    name,
    sampleCount,
    lastSeen: new Date(),
  };
}

export function createMockAttribute(name: string, type: string = 'string') {
  return {
    name,
    type,
    cardinality: 'medium' as const,
    sampleValues: ['sample1', 'sample2', 'sample3'],
  };
}

export function createMockEntity(guid: string, name: string, type: string = 'APPLICATION') {
  return {
    guid,
    name,
    type,
    domain: 'APM',
    entityType: type,
    reporting: true,
    tags: [],
  };
}