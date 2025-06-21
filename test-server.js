#!/usr/bin/env node

import { spawn } from 'child_process';
import { fileURLToPath } from 'url';
import { dirname, join } from 'path';

const __dirname = dirname(fileURLToPath(import.meta.url));

// Check if environment variables are set
const requiredEnvVars = ['NEW_RELIC_API_KEY', 'NEW_RELIC_ACCOUNT_ID'];
const missing = requiredEnvVars.filter(key => !process.env[key]);

if (missing.length > 0) {
  console.error('❌ Missing required environment variables:');
  missing.forEach(key => console.error(`  - ${key}`));
  console.error('\nPlease set these environment variables and try again.');
  console.error('You can copy .env.example to .env and fill in your values.');
  process.exit(1);
}

console.log('✅ Environment variables configured');
console.log(`  Account ID: ${process.env.NEW_RELIC_ACCOUNT_ID}`);
console.log(`  Region: ${process.env.NEW_RELIC_REGION || 'US'}`);
console.log(`  Debug: ${process.env.DEBUG === 'true' ? 'enabled' : 'disabled'}`);

// Test the built server
console.log('\n🚀 Starting MCP Server...');

const server = spawn('node', [join(__dirname, 'dist', 'index.js')], {
  env: process.env,
  stdio: 'inherit'
});

server.on('error', (err) => {
  console.error('❌ Failed to start server:', err.message);
  process.exit(1);
});

server.on('exit', (code) => {
  if (code !== 0) {
    console.error(`❌ Server exited with code ${code}`);
    process.exit(code);
  }
});

// Handle graceful shutdown
process.on('SIGINT', () => {
  console.log('\n👋 Shutting down server...');
  server.kill('SIGINT');
});

process.on('SIGTERM', () => {
  server.kill('SIGTERM');
});