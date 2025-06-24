# MCP Server New Relic - Hello World

This is a minimal "Hello World" implementation of an MCP server for New Relic. It demonstrates basic connectivity to the New Relic NerdGraph API.

## Quick Start

### Prerequisites
- Node.js 18+ installed
- New Relic API key (User or Personal API key with NerdGraph access)
- npm or yarn package manager

### Installation

1. Clone this repository:
```bash
git clone https://github.com/yourusername/mcp-server-newrelic.git
cd mcp-server-newrelic
```

2. Install dependencies:
```bash
npm install
```

3. Set up your API key:
```bash
# Copy the .env template and add your API key
cp .env .env.local
# Edit .env.local and replace 'your-api-key-here' with your actual API key
```

4. Build the TypeScript code:
```bash
npm run build
```

5. Test the server:
```bash
npm start
```

### Using with MCP Client

Add this server to your MCP client configuration:

```json
{
  "mcpServers": {
    "newrelic": {
      "command": "node",
      "args": ["/absolute/path/to/mcp-server-newrelic/dist/index.js"],
      "env": {
        "NEW_RELIC_API_KEY": "your-api-key-here"
      }
    }
  }
}
```

### Available Tools

#### hello_newrelic
Tests the connection to New Relic and returns basic account information.

**Parameters:**
- `account_id` (number, required): The New Relic account ID to query

**Example:**
```json
{
  "name": "hello_newrelic",
  "arguments": {
    "account_id": 1234567
  }
}
```

**Response:**
```
Hello from New Relic! 👋

Account Information:
- ID: 1234567
- Name: Production Account
- Transactions (last hour): 125432
```

### Troubleshooting

1. **"Cannot find module" error**: Run `npm run build` before starting
2. **"API Key Required" error**: Ensure NEW_RELIC_API_KEY is set in environment
3. **"GraphQL Error" response**: Verify your API key has correct permissions
4. **Connection timeout**: Check network connectivity to api.newrelic.com

### Development

To run in development mode:
```bash
npm run dev
```

This will compile and run the server in one step.

## GitHub Copilot Integration

This MCP server can be used with GitHub Copilot to provide New Relic data directly in your IDE.

### Visual Studio Code Setup

1. **Repository-specific configuration** (Recommended for team sharing):
   - The `.vscode/mcp.json` file is already included in this repository
   - Set your API key as an environment variable: `export NEW_RELIC_API_KEY=your-api-key`
   - Reload VS Code window

2. **User-specific configuration** (For personal use):
   - Copy `.vscode/settings.json.example` to your VS Code user settings
   - Update the path to point to your installation
   - Add your API key

### Using with Copilot Chat

Once configured, you can use the New Relic MCP server in Copilot Chat:

1. Open Copilot Chat in VS Code
2. The New Relic tools will be available automatically
3. Ask Copilot to query your New Relic account:
   - "Using the hello_newrelic tool, show me account 1234567"
   - "Check New Relic account information for account 1234567"

### Supported IDEs

MCP support is available in:
- Visual Studio Code (GA)
- JetBrains IDEs (Public Preview)
- Xcode (Public Preview)  
- Eclipse (Public Preview)

### Next Steps

This is just a Hello World implementation. See `hello_world_tech_spec.md` for the roadmap to expand functionality.