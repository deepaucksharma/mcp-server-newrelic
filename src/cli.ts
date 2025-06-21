#!/usr/bin/env node
/**
 * CLI Tool for MCP Server New Relic
 * 
 * Provides command-line access to discovery and testing capabilities
 */

import { createRequestContext } from './core/context.js';
import { DiscoveryEngine } from './core/discovery/engine.js';

async function discoverCommand(accountId: string, apiKey: string, region: 'US' | 'EU' = 'US') {
  console.error('🔍 Starting discovery process...');
  
  try {
    const ctx = await createRequestContext({
      accountId: parseInt(accountId),
      apiKey,
      region,
    });

    const discoveryEngine = new DiscoveryEngine(ctx);
    const worldModel = await discoveryEngine.buildDiscoveryGraph(parseInt(accountId));

    console.log('\n📊 Discovery Results:');
    console.log('===================');
    console.log(`Account ID: ${worldModel.accountId}`);
    console.log(`Confidence: ${(worldModel.confidence * 100).toFixed(1)}%`);
    console.log(`Schemas discovered: ${worldModel.schemas.length}`);
    console.log(`Attributes profiled: ${Object.keys(worldModel.attributes).length}`);
    console.log(`Error indicators: ${worldModel.errorIndicators.length}`);
    console.log(`Metrics available: ${worldModel.metrics.length}`);

    console.log('\n📋 Top Schemas:');
    worldModel.schemas.slice(0, 5).forEach((schema, index) => {
      console.log(`${index + 1}. ${schema.eventType} (${schema.count.toLocaleString()} records)`);
    });

    if (worldModel.serviceIdentifier.confidence > 0.5) {
      console.log(`\n🔗 Service Identifier: ${worldModel.serviceIdentifier.field} (${(worldModel.serviceIdentifier.confidence * 100).toFixed(1)}% confidence)`);
    }

    if (worldModel.errorIndicators.length > 0) {
      console.log('\n❌ Error Indicators:');
      worldModel.errorIndicators.slice(0, 3).forEach((indicator, index) => {
        console.log(`${index + 1}. ${indicator.field} in ${indicator.eventType} (${(indicator.confidence * 100).toFixed(1)}% confidence)`);
      });
    }

    console.log('\n🧠 Explainability Trace:');
    console.log(ctx.explainabilityTrace.toMarkdown());

  } catch (error: any) {
    console.error('❌ Discovery failed:', error.message);
    process.exit(1);
  }
}

async function main() {
  const args = process.argv.slice(2);
  const command = args[0];

  if (command === 'discover') {
    const accountId = process.env['NEW_RELIC_ACCOUNT_ID'] || args[1];
    const apiKey = process.env['NEW_RELIC_API_KEY'] || args[2];
    const region = (process.env['NEW_RELIC_REGION'] as 'US' | 'EU') || 'US';

    if (!accountId || !apiKey) {
      console.error('Usage: npm run discover [account_id] [api_key]');
      console.error('Or set NEW_RELIC_ACCOUNT_ID and NEW_RELIC_API_KEY environment variables');
      process.exit(1);
    }

    await discoverCommand(accountId, apiKey, region);
  } else {
    console.error('Available commands:');
    console.error('  discover - Run discovery engine against New Relic account');
    console.error('');
    console.error('Environment variables:');
    console.error('  NEW_RELIC_ACCOUNT_ID - Your New Relic account ID');
    console.error('  NEW_RELIC_API_KEY - Your New Relic API key');
    console.error('  NEW_RELIC_REGION - US or EU (default: US)');
    process.exit(1);
  }
}

main().catch(error => {
  console.error('CLI error:', error);
  process.exit(1);
});