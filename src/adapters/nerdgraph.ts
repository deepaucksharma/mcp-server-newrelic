/**
 * New Relic NerdGraph GraphQL Client
 * 
 * Provides typed access to New Relic's GraphQL API with proper error handling,
 * retries, and rate limiting compliance.
 */

import { GraphQLClient } from 'graphql-request';
import { NerdGraphClient, Logger } from '../core/types.js';

export interface NerdGraphClientOptions {
  apiKey: string;
  region: 'US' | 'EU';
  timeout?: number;
  maxRetries?: number;
  logger?: Logger;
}

/**
 * Production NerdGraph client implementation
 */
export class NerdGraphClientImpl implements NerdGraphClient {
  private client: GraphQLClient;
  private logger: Logger;
  private maxRetries: number;

  constructor(private options: NerdGraphClientOptions) {
    const endpoint = this.getEndpoint(options.region);
    
    this.client = new GraphQLClient(endpoint, {
      headers: {
        'API-Key': options.apiKey,
        'Content-Type': 'application/json',
        'User-Agent': 'mcp-server-newrelic/1.0.0',
      },
    });

    this.logger = options.logger || console;
    this.maxRetries = options.maxRetries || 3;
  }

  /**
   * Execute a GraphQL query with retry logic
   */
  async request<T = any>(query: string, variables?: Record<string, any>): Promise<T> {
    let lastError: Error | undefined;

    for (let attempt = 0; attempt <= this.maxRetries; attempt++) {
      try {
        this.logger.debug('Executing GraphQL query', {
          attempt: attempt + 1,
          hasVariables: !!variables,
          queryLength: query.length,
        });

        const result = await this.client.request<T>(query, variables);
        
        if (attempt > 0) {
          this.logger.info('GraphQL query succeeded after retry', {
            attempt: attempt + 1,
          });
        }

        return result;

      } catch (error: any) {
        lastError = error;
        
        // Don't retry on authentication errors
        if (this.isAuthError(error)) {
          this.logger.error('Authentication failed - check API key', {
            region: this.options.region,
            error: error.message,
          });
          throw new Error('New Relic authentication failed: Check your API key and region');
        }

        // Don't retry on client errors (4xx)
        if (this.isClientError(error)) {
          this.logger.error('Client error in GraphQL query', {
            error: error.message,
            query: query.substring(0, 200),
          });
          throw error;
        }

        // Retry on network/server errors with exponential backoff
        if (attempt < this.maxRetries) {
          const delay = Math.min(1000 * Math.pow(2, attempt), 10000);
          this.logger.warn('GraphQL query failed, retrying', {
            attempt: attempt + 1,
            maxRetries: this.maxRetries,
            delay,
            error: error.message,
          });
          
          await this.sleep(delay);
          continue;
        }
      }
    }

    this.logger.error('GraphQL query failed after all retries', {
      maxRetries: this.maxRetries,
      error: lastError?.message,
    });

    throw lastError || new Error('Unknown error occurred');
  }

  /**
   * Execute an NRQL query via NerdGraph
   */
  async nrql(accountId: number, query: string): Promise<any> {
    const graphqlQuery = `
      query($accountId: Int!, $nrql: Nrql!) {
        actor {
          account(id: $accountId) {
            nrql(query: $nrql) {
              results
              metadata {
                timeWindow {
                  begin
                  end
                }
                rawResponse
              }
            }
          }
        }
      }
    `;

    try {
      this.logger.debug('Executing NRQL query', {
        accountId,
        query: query.substring(0, 100) + (query.length > 100 ? '...' : ''),
      });

      const result = await this.request(graphqlQuery, {
        accountId,
        nrql: query,
      });

      const nrqlResult = result.actor?.account?.nrql;
      if (!nrqlResult) {
        throw new Error('Invalid NRQL response structure');
      }

      // Transform the result to a more convenient format
      return {
        results: nrqlResult.results || [],
        facets: this.extractFacets(nrqlResult.results),
        metadata: nrqlResult.metadata,
      };

    } catch (error: any) {
      this.logger.error('NRQL query failed', {
        accountId,
        query,
        error: error.message,
      });

      // Re-throw with more context
      throw new Error(`NRQL query failed: ${error.message}`);
    }
  }

  /**
   * Get account information
   */
  async getAccountInfo(accountId: number): Promise<any> {
    const query = `
      query($accountId: Int!) {
        actor {
          account(id: $accountId) {
            id
            name
            licenseKey
            reportingEventTypes
          }
        }
      }
    `;

    const result = await this.request(query, { accountId });
    return result.actor?.account;
  }

  /**
   * Execute entity search
   */
  async searchEntities(query: string, accountId?: number): Promise<any> {
    const searchQuery = `
      query($query: String!, $accountId: Int) {
        actor {
          entitySearch(query: $query) {
            results {
              entities {
                guid
                name
                type
                domain
                reporting
                ... on ApmApplicationEntity {
                  alertSeverity
                  runningAgentVersions {
                    maxVersion
                    minVersion
                  }
                }
              }
            }
          }
        }
      }
    `;

    const result = await this.request(searchQuery, { query, accountId });
    return result.actor?.entitySearch?.results?.entities || [];
  }

  /**
   * Get the correct GraphQL endpoint for the region
   */
  private getEndpoint(region: 'US' | 'EU'): string {
    return region === 'EU' 
      ? 'https://api.eu.newrelic.com/graphql'
      : 'https://api.newrelic.com/graphql';
  }

  /**
   * Check if error is authentication-related
   */
  private isAuthError(error: any): boolean {
    return error?.response?.status === 401 || 
           error?.response?.status === 403 ||
           error?.message?.toLowerCase().includes('authentication') ||
           error?.message?.toLowerCase().includes('unauthorized') ||
           error?.message?.toLowerCase().includes('forbidden');
  }

  /**
   * Check if error is a client error (4xx)
   */
  private isClientError(error: any): boolean {
    const status = error?.response?.status;
    return status >= 400 && status < 500;
  }

  /**
   * Sleep for specified milliseconds
   */
  private sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  /**
   * Extract facets from NRQL results for easier processing
   */
  private extractFacets(results: any[]): any[] {
    if (!Array.isArray(results) || results.length === 0) {
      return [];
    }

    // Check if this looks like a faceted query result
    const firstResult = results[0];
    if (!firstResult || typeof firstResult !== 'object') {
      return [];
    }

    // Look for facet patterns in the results
    const facets: any[] = [];
    
    results.forEach(result => {
      // Find non-aggregation fields (likely facets)
      Object.keys(result).forEach(key => {
        if (!key.includes('.') && !['count', 'average', 'sum', 'min', 'max'].includes(key)) {
          const existingFacet = facets.find(f => f.name === result[key]);
          if (!existingFacet) {
            facets.push({
              name: result[key],
              results: [result],
            });
          } else {
            existingFacet.results.push(result);
          }
        }
      });
    });

    return facets;
  }
}

/**
 * Factory function to create a NerdGraph client
 */
export function createNerdGraphClient(options: NerdGraphClientOptions): NerdGraphClient {
  return new NerdGraphClientImpl(options);
}