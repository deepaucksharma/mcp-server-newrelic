# Testing MCP Server New Relic with GitHub Copilot

## Prerequisites

1. GitHub Copilot subscription active
2. Visual Studio Code with GitHub Copilot extension installed
3. New Relic API key (User or Personal API key)
4. Node.js 18+ installed

## Setup Instructions

### 1. Install and Build

```bash
# Clone and install
git clone https://github.com/yourusername/mcp-server-newrelic.git
cd mcp-server-newrelic
npm install
npm run build
```

### 2. Configure Environment

```bash
# Set your New Relic API key
export NEW_RELIC_API_KEY=your-actual-api-key-here
```

### 3. Test Standalone

```bash
# Run the test script
node test-hello.js
```

Expected output:
- ✅ Connected to MCP server
- 📋 Available tools listed
- 📊 Result with account information (or error if invalid API key)

### 4. Configure VS Code

The repository includes `.vscode/mcp.json` which will automatically configure the MCP server when you open the project in VS Code.

### 5. Test in Copilot Chat

1. Open VS Code in the project directory
2. Open Copilot Chat (Cmd/Ctrl + Shift + I)
3. Test commands:

```
@workspace Can you list the available MCP tools?
```

```
@workspace Using the hello_newrelic tool, check account 1234567
```

```
@workspace What New Relic account information can you access?
```

## Expected Behaviors

### Success Case
- Copilot recognizes the `hello_newrelic` tool
- Tool executes and returns account information
- Response includes account ID, name, and transaction count

### Error Cases
- Invalid API key: Returns error message about authentication
- Invalid account ID: Returns error about account access
- No API key set: Server fails to start

## Troubleshooting

### MCP Server Not Available in Copilot

1. Check VS Code Output panel for MCP errors
2. Ensure `.vscode/mcp.json` exists
3. Reload VS Code window (Cmd/Ctrl + R)
4. Check environment variable is set

### Authentication Errors

1. Verify API key has NerdGraph access
2. Test API key with curl:
```bash
curl -X POST https://api.newrelic.com/graphql \
  -H 'API-Key: YOUR_API_KEY' \
  -H 'Content-Type: application/json' \
  -d '{"query":"{ actor { user { email } } }"}'
```

### Build Errors

1. Ensure TypeScript is installed: `npm install`
2. Clean and rebuild: `rm -rf dist && npm run build`
3. Check Node.js version: `node --version` (should be 18+)

## Advanced Testing

### Testing with Multiple Accounts

Create a test script to verify multiple accounts:

```javascript
// test-multiple.js
const accounts = [1234567, 7654321, 9999999];
for (const id of accounts) {
  console.log(`Testing account ${id}...`);
  // Use the tool with each account
}
```

### Performance Testing

Monitor response times:
- Tool discovery: <100ms
- Tool execution: <3s (depends on API latency)
- Memory usage: <50MB

## Integration Verification Checklist

- [ ] MCP server builds without errors
- [ ] Test script runs successfully
- [ ] VS Code recognizes the MCP server
- [ ] Copilot Chat can list available tools
- [ ] hello_newrelic tool executes in Copilot
- [ ] Error handling works properly
- [ ] Environment variables are respected

## Next Steps

Once basic testing is complete:
1. Add more tools (see hello_world_tech_spec.md)
2. Test with real New Relic accounts
3. Share with team for feedback
4. Consider publishing to npm for easier distribution