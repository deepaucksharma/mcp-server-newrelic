#!/usr/bin/env python3

import os
import json
import asyncio
import sys
from pathlib import Path

# Add project root to path
sys.path.insert(0, str(Path(__file__).parent))

# Import the MCP server
from mcp_server import MCPServer

async def test_real_queries():
    """Test real NRDB queries with actual New Relic account"""
    
    # Initialize server
    server = MCPServer()
    
    print("=== Testing Real NRDB Queries ===")
    
    # Test 1: List Dashboards
    print("\n1. Testing list_dashboards...")
    try:
        result = await server.handle_list_dashboards({})
        print(f"✓ Found {result['total']} dashboards")
        if result['total'] > 0:
            print(f"  First dashboard: {result['dashboards'][0]['name']}")
    except Exception as e:
        print(f"✗ Failed to list dashboards: {e}")
    
    # Test 2: Query NRDB
    print("\n2. Testing query_nrdb...")
    try:
        result = await server.handle_query_nrdb({
            "query": "SELECT count(*) FROM Transaction SINCE 1 hour ago"
        })
        print(f"✓ Query executed successfully")
        print(f"  Result: {json.dumps(result['results'], indent=2)}")
    except Exception as e:
        print(f"✗ Failed to execute query: {e}")
    
    # Test 3: Discover Schemas
    print("\n3. Testing discovery.list_schemas...")
    try:
        result = await server.handle_list_schemas({})
        print(f"✓ Found {len(result['schemas'])} schemas")
        # Show first 5
        for i, schema in enumerate(result['schemas'][:5]):
            print(f"  - {schema['name']} ({schema['sample_count']} samples)")
        if len(result['schemas']) > 5:
            print(f"  ... and {len(result['schemas']) - 5} more")
    except Exception as e:
        print(f"✗ Failed to list schemas: {e}")
    
    # Test 4: List Alerts
    print("\n4. Testing list_alerts...")
    try:
        result = await server.handle_list_alerts({})
        print(f"✓ Found {result['total']} alert conditions")
    except Exception as e:
        print(f"✗ Failed to list alerts: {e}")
    
    # Test 5: Query Builder
    print("\n5. Testing query_builder...")
    try:
        result = await server.handle_query_builder({
            "event_type": "Transaction",
            "select": ["count(*)", "average(duration)"],
            "where": "appName IS NOT NULL",
            "since": "1 hour ago"
        })
        print(f"✓ Query built successfully")
        print(f"  Generated query: {result['query']}")
        
        # Execute the built query
        exec_result = await server.handle_query_nrdb({"query": result['query']})
        print(f"  Execution result: {json.dumps(exec_result['results'], indent=2)}")
    except Exception as e:
        print(f"✗ Failed to build/execute query: {e}")
    
    print("\n=== All Tests Completed ===")

if __name__ == "__main__":
    # Load environment variables
    from dotenv import load_dotenv
    load_dotenv()
    
    # Verify credentials are set
    if not os.getenv("NEW_RELIC_API_KEY"):
        print("ERROR: NEW_RELIC_API_KEY not set in environment")
        sys.exit(1)
    
    if not os.getenv("NEW_RELIC_ACCOUNT_ID"):
        print("ERROR: NEW_RELIC_ACCOUNT_ID not set in environment") 
        sys.exit(1)
    
    # Run tests
    asyncio.run(test_real_queries())