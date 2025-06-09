#!/usr/bin/env python3
"""
Direct test of New Relic MCP functionality
"""

import asyncio
import os
import sys
import json

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.abspath(__file__)))

from features import alerts, apm, entities, common

async def test_newrelic_api():
    """Test New Relic API functionality directly"""
    
    # Check if we have credentials
    api_key = os.getenv("NEW_RELIC_API_KEY")
    account_id = os.getenv("NEW_RELIC_ACCOUNT_ID")
    
    if not api_key or not account_id:
        print("ERROR: Missing environment variables.")
        print("Please set NEW_RELIC_API_KEY and NEW_RELIC_ACCOUNT_ID.")
        return
    
    print(f"Testing with account ID: {account_id}")
    
    # Create a simple mock for the fastmcp object
    class MockMCP:
        def tool(self):
            def decorator(func):
                return func
            return decorator
        
        def resource(self, path):
            def decorator(func):
                return func
            return decorator
    
    mock_mcp = MockMCP()
    
    # Register the features but capture the functions for direct testing
    registered_functions = {}
    
    # Override the tool decorator to capture functions
    original_tool = mock_mcp.tool
    def capturing_tool():
        def decorator(func):
            registered_functions[func.__name__] = func
            return func
        return decorator
    mock_mcp.tool = capturing_tool
    
    # Register all features
    from core.nerdgraph_client import NerdGraphClient
    
    # Create NerdGraph client
    nerdgraph = NerdGraphClient(api_key=api_key)
    
    services = {
        "nerdgraph": nerdgraph,
        "account_id": account_id
    }
    
    # Register plugins and capture functions
    common.register(mock_mcp)
    alerts.register(mock_mcp)
    apm.register(mock_mcp)
    entities.register(mock_mcp)
    
    # Now test a few functions
    print("\n--- Testing List Alert Policies ---")
    if "list_alert_policies" in registered_functions:
        result = await registered_functions["list_alert_policies"]()
        print(result)
    
    print("\n--- Testing Run NRQL Query ---")
    if "run_nrql_query" in registered_functions:
        result = await registered_functions["run_nrql_query"]("SELECT count(*) FROM Transaction SINCE 1 hour ago")
        print(result)

if __name__ == "__main__":
    asyncio.run(test_newrelic_api())
