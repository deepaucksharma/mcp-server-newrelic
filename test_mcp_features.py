#!/usr/bin/env python3
"""
Test script for New Relic MCP Server
This script tests various features of the New Relic MCP Server
"""

import sys
import os
import json
import asyncio
from mcp import MCPClient

async def test_mcp_features():
    """Test the features of the MCP server"""
    
    print("\n=== Testing New Relic MCP Server ===\n")
    
    # Connect to MCP server (stdio transport)
    client = MCPClient(transport="stdio")
    await client.connect()
    
    # Get the server info
    server_info = await client.get_server_info()
    print(f"Connected to: {server_info['name']} v{server_info['version']}")
    print(f"Description: {server_info['description']}")
    
    # Get available tools
    tools = await client.list_tools()
    print(f"\nAvailable tools ({len(tools)}):")
    for tool in tools:
        print(f"  - {tool['name']}")
    
    # Test some of the tools
    try:
        print("\n--- Testing list_apm_applications ---")
        result = await client.call_tool("list_apm_applications")
        print_json_result(result)
        
        print("\n--- Testing list_alert_policies ---")
        result = await client.call_tool("list_alert_policies")
        print_json_result(result)
        
        print("\n--- Testing search_entities ---")
        result = await client.call_tool("search_entities", {"entity_type": "APPLICATION"})
        print_json_result(result)
        
        print("\n--- Testing list_synthetics_monitors ---")
        result = await client.call_tool("list_synthetics_monitors")
        print_json_result(result)
        
    except Exception as e:
        print(f"Error during testing: {e}")
    
    # Close the connection
    await client.disconnect()
    print("\n=== Test completed ===")

def print_json_result(result):
    """Pretty print JSON results"""
    try:
        # If result is already a string, try to parse it as JSON
        if isinstance(result, str):
            data = json.loads(result)
            print(json.dumps(data, indent=2)[:500] + "... (truncated)")
        else:
            print(json.dumps(result, indent=2)[:500] + "... (truncated)")
    except Exception as e:
        print(f"Error formatting result: {e}")
        print(result)

if __name__ == "__main__":
    asyncio.run(test_mcp_features())
