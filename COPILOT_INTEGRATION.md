# Using New Relic MCP Server with GitHub Copilot

This guide explains how to set up and use the New Relic MCP Server with GitHub Copilot in VS Code.

## What is Model Context Protocol (MCP)?

Model Context Protocol (MCP) allows AI assistants like GitHub Copilot to access external tools and services. This integration lets Copilot query your New Relic observability data directly, providing you with insights about your applications' performance, infrastructure, and alerts.

## Setup

### Prerequisites

1. VS Code with GitHub Copilot and GitHub Copilot Chat extensions installed
2. New Relic account with an API key
3. Python 3.8+ installed

### Configuration

The repository is already configured for use with GitHub Copilot through:
- `.vscode/settings.json` - Connects Copilot to the local MCP server
- `.vscode/tasks.json` - Provides tasks to run the MCP server

### Steps to Use

1. **Set your New Relic API Key**:
   ```bash
   export NEW_RELIC_API_KEY=your_api_key_here
   export NEW_RELIC_ACCOUNT_ID=your_account_id_here
   ```

2. **Start the MCP Server**:
   - Open VS Code Command Palette (Cmd+Shift+P or Ctrl+Shift+P)
   - Type "Tasks: Run Task"
   - Select "Run MCP Server for GitHub Copilot"

3. **Use GitHub Copilot Chat to query New Relic data**:
   - Open GitHub Copilot Chat in VS Code
   - Use the /newrelic command:
   ```
   @copilot /newrelic What's the error rate for my applications?
   ```
   
## Example Queries

- `@copilot /newrelic Show me the health of my applications`
- `@copilot /newrelic Are there any current incidents?`
- `@copilot /newrelic What's the average response time for the checkout service?`
- `@copilot /newrelic Show infrastructure metrics for my production hosts`
- `@copilot /newrelic Run NRQL query SELECT count(*) FROM Transaction WHERE appName = 'MyApp' SINCE 1 hour ago`

## Troubleshooting

If you encounter issues:

1. **Check MCP Server Logs**: Look at the terminal where the MCP server is running for error messages
2. **Verify API Key**: Ensure your New Relic API key has the correct permissions
3. **Transport Mode**: The server uses stdio transport mode to communicate with GitHub Copilot
4. **Python Version**: Make sure you're using Python 3.10 or higher (Python 3.11 recommended)
5. **Restart Server**: Sometimes stopping and restarting the MCP server resolves connectivity issues

## Advanced Configuration

For advanced users, you can modify the `.vscode/settings.json` file to:
- Change the port number
- Adjust environment variables
- Set different initialization options

## Contributing

If you enhance this integration, please update this guide and the configuration files accordingly.
