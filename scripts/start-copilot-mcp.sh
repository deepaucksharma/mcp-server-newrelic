#!/usr/bin/env bash
# Script to launch the New Relic MCP Server for GitHub Copilot integration

# Source .env file if it exists
if [ -f ../.env ]; then
    echo "Loading environment from .env file..."
    export $(grep -v '^#' ../.env | xargs)
elif [ -f ./.env ]; then
    echo "Loading environment from .env file..."
    export $(grep -v '^#' ./.env | xargs)
fi

# Set default values
TRANSPORT=${MCP_TRANSPORT:-stdio}
LOG_LEVEL=${LOG_LEVEL:-INFO}

# Display banner
echo "============================================="
echo "New Relic MCP Server for GitHub Copilot"
echo "============================================="
echo "Transport: $TRANSPORT"
echo "Log Level: $LOG_LEVEL"
echo "============================================="

# Check if API key is set
if [ -z "$NEW_RELIC_API_KEY" ]; then
    echo "ERROR: NEW_RELIC_API_KEY environment variable is not set."
    echo "Please set it before running this script:"
    echo "export NEW_RELIC_API_KEY=your_api_key_here"
    exit 1
fi

# Run the server with appropriate environment variables
MCP_TRANSPORT=$TRANSPORT \
LOG_LEVEL=$LOG_LEVEL \
USE_ENHANCED_PLUGINS=true \
python3.11 server_simple.py

echo "Server stopped."
