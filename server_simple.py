#!/usr/bin/env python3
"""
Simplified server for testing - uses existing feature modules
Updated for FastMCP 2.0 compatibility
"""

import os
import logging
from fastmcp import FastMCP
from fastmcp.utilities import setup_logging

# Configure logging with FastMCP 2.0 utilities
setup_logging(level=os.getenv('LOG_LEVEL', 'INFO'))
logger = logging.getLogger(__name__)

# Create the MCP instance with FastMCP 2.0 features
mcp = FastMCP(
    name="newrelic-mcp",
    version="2.0.0", 
    description="Access New Relic platform data through Model Context Protocol",
    instructions="""
Simple New Relic MCP Server for testing and development.
Available capabilities include:
- Query application performance metrics (APM)
- Search and inspect entities (services, hosts, etc.)
- View alerts and incidents
- Run NRQL queries for custom analysis
- Monitor synthetic checks

This is a simplified version optimized for quick testing and development.
"""
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
    # Determine transport mode with FastMCP 2.0 support
    transport = os.getenv("MCP_TRANSPORT", "stdio")
    
    if transport == "http":
        # Use streamable HTTP transport for production
        host = os.getenv("HTTP_HOST", "127.0.0.1")
        port = int(os.getenv("HTTP_PORT", "3000"))
        logger.info(f"Starting in HTTP mode on {host}:{port}")
        mcp.run(transport="http", host=host, port=port)
    elif transport == "sse":
        # Server-sent events transport (legacy support)
        host = os.getenv("SSE_HOST", "127.0.0.1")
        port = int(os.getenv("SSE_PORT", "3001"))
        logger.info(f"Starting in SSE mode on {host}:{port}")
        mcp.run(transport="sse", host=host, port=port)
    else:
        # STDIO mode for desktop AI assistants (Claude Desktop, GitHub Copilot)
        logger.info("Starting in STDIO mode for desktop AI assistants")
        mcp.run()