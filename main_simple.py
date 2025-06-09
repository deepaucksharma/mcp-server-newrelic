#!/usr/bin/env python3
"""
New Relic MCP Server - Simplified Entry Point
"""

import os
import sys
import logging
from fastmcp import FastMCP

# Configure logging
logging.basicConfig(
    level=os.getenv('LOG_LEVEL', 'INFO'),
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

# Create the MCP instance
mcp = FastMCP(
    name="newrelic-mcp",
    version="1.0.0",
    description="Access New Relic platform data through Model Context Protocol",
    instructions="""
You have access to New Relic observability data through this MCP server.
Available capabilities include:
- Query application performance metrics (APM)
- Search and inspect entities (services, hosts, etc.)
- View alerts and incidents
- Run NRQL queries for custom analysis
- Explore entity relationships and dependencies

Always specify clear time ranges when querying metrics. Default is last hour.
Entity names are case-sensitive. Use search if unsure of exact name.
"""
)

# Check for required environment variables
if not os.getenv('NEW_RELIC_API_KEY'):
    logger.error("NEW_RELIC_API_KEY environment variable is required")
    sys.exit(1)

# Import and register existing features
try:
    from features import common, entities, apm, synthetics, alerts
    
    logger.info("Registering features...")
    
    # Create minimal services dict for features
    services = {
        "account_id": os.getenv('NEW_RELIC_ACCOUNT_ID'),
        "api_key": os.getenv('NEW_RELIC_API_KEY'),
        "nerdgraph_url": os.getenv('NERDGRAPH_URL', 'https://api.newrelic.com/graphql')
    }
    
    # Register features with simplified approach
    common.register(mcp)
    entities.register(mcp)
    apm.register(mcp)
    synthetics.register(mcp)
    alerts.register(mcp)
    
    logger.info("Server ready")
    
except ImportError as e:
    logger.error(f"Failed to import features: {e}")
    # Continue with basic functionality
    
    @mcp.tool()
    def health_check() -> str:
        """Basic health check for the MCP server"""
        return "New Relic MCP Server is running"

if __name__ == "__main__":
    logger.info("Starting New Relic MCP Server...")
    
    # Determine transport mode
    transport = os.getenv("MCP_TRANSPORT", "stdio")
    
    if transport == "http":
        logger.info("Starting in HTTP mode...")
        # For HTTP mode, you may need additional configuration
        mcp.run(transport="http", host="127.0.0.1", port=3000)
    else:
        logger.info("Starting in STDIO mode...")
        mcp.run()