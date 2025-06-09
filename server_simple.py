#!/usr/bin/env python3
"""
Simplified server for testing - uses existing feature modules
"""

import os
from fastmcp import FastMCP

# Create the MCP instance
mcp = FastMCP(
    name="newrelic-mcp",
    version="1.0.0",
    description="Access New Relic platform data through Model Context Protocol"
)

# Import and register existing features
from features import common, entities, apm, synthetics, alerts

print("Registering features...")
common.register(mcp)
entities.register(mcp)
apm.register(mcp)
synthetics.register(mcp)
alerts.register(mcp)

print("Server ready")

if __name__ == "__main__":
    import logging
    logger = logging.getLogger(__name__)
    
    # Determine transport mode
    transport = os.getenv("MCP_TRANSPORT", "stdio")
    
    if transport == "http":
        host = os.getenv("HTTP_HOST", "127.0.0.1")
        port = int(os.getenv("HTTP_PORT", "3000"))
        logger.info(f"Starting in HTTP mode on {host}:{port}")
        mcp.run(transport="http", host=host, port=port)
    else:
        logger.info("Starting in STDIO mode")
        mcp.run()