# Installation Guide

This comprehensive guide covers all installation methods for the New Relic MCP Server.

## 📋 Table of Contents

1. [System Requirements](#system-requirements)
2. [Installation Methods](#installation-methods)
3. [Docker Installation](#docker-installation)
4. [Binary Installation](#binary-installation)
5. [From Source](#from-source)
6. [Claude Desktop Integration](#claude-desktop-integration)
7. [Configuration](#configuration)
8. [Verification](#verification)
9. [Upgrading](#upgrading)
10. [Uninstallation](#uninstallation)

## 🖥️ System Requirements

### Minimum Requirements
- **CPU**: 2 cores
- **RAM**: 1GB
- **Disk**: 100MB for binaries
- **OS**: Linux, macOS, Windows (with WSL)
- **Network**: Internet access to New Relic APIs

### Software Requirements
- **For Docker**: Docker 20.10+ and Docker Compose 2.0+
- **For Source Build**: Go 1.21+
- **For Binary**: None (standalone executable)

### New Relic Requirements
- Active New Relic account
- User API Key with query permissions
- Account ID

## 🚀 Installation Methods

Choose the method that best fits your needs:

| Method | Best For | Complexity | Updates |
|--------|----------|------------|---------|
| Docker | Production, Quick Start | Easy | Automated |
| Binary | Simple deployments | Easy | Manual |
| Source | Development, Customization | Medium | Manual |

## 🐳 Docker Installation

### Prerequisites
```bash
# Check Docker version
docker --version  # Should be 20.10+
docker-compose --version  # Should be 2.0+
```

### Step 1: Clone Repository
```bash
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic
```

### Step 2: Configure Environment
```bash
# Copy example configuration
cp .env.example .env

# Edit with your credentials
nano .env  # or your preferred editor
```

Required settings in `.env`:
```env
NEW_RELIC_API_KEY=NRAK-your-key-here
NEW_RELIC_ACCOUNT_ID=your-account-id
NEW_RELIC_REGION=US  # or EU
```

### Step 3: Build and Run

#### Option A: Simple Mode (MCP Server only)
```bash
# Use simple compose file
docker-compose -f docker-compose.simple.yml up -d

# Check status
docker-compose -f docker-compose.simple.yml ps
```

#### Option B: Full Stack (with monitoring)
```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f mcp-server
```

### Docker Commands Reference
```bash
# Start services
docker-compose up -d

# Stop services
docker-compose down

# View logs
docker-compose logs -f [service-name]

# Restart a service
docker-compose restart mcp-server

# Update and rebuild
docker-compose pull
docker-compose up -d --build
```

## 📦 Binary Installation

### Step 1: Download Binary

Download the appropriate binary for your platform:

```bash
# Linux (amd64)
wget https://github.com/deepaucksharma/mcp-server-newrelic/releases/latest/download/mcp-server-linux-amd64
chmod +x mcp-server-linux-amd64
mv mcp-server-linux-amd64 /usr/local/bin/mcp-server

# macOS (Intel)
wget https://github.com/deepaucksharma/mcp-server-newrelic/releases/latest/download/mcp-server-darwin-amd64
chmod +x mcp-server-darwin-amd64
mv mcp-server-darwin-amd64 /usr/local/bin/mcp-server

# macOS (Apple Silicon)
wget https://github.com/deepaucksharma/mcp-server-newrelic/releases/latest/download/mcp-server-darwin-arm64
chmod +x mcp-server-darwin-arm64
mv mcp-server-darwin-arm64 /usr/local/bin/mcp-server

# Windows (use WSL or download manually)
```

### Step 2: Create Configuration

```bash
# Create config directory
mkdir -p ~/.config/mcp-newrelic

# Create environment file
cat > ~/.config/mcp-newrelic/.env << EOF
NEW_RELIC_API_KEY=NRAK-your-key-here
NEW_RELIC_ACCOUNT_ID=your-account-id
NEW_RELIC_REGION=US
EOF
```

### Step 3: Run Binary

```bash
# Run with config file
mcp-server --env-file ~/.config/mcp-newrelic/.env

# Or export environment variables
export NEW_RELIC_API_KEY=NRAK-your-key-here
export NEW_RELIC_ACCOUNT_ID=your-account-id
mcp-server
```

## 🔨 From Source

### Prerequisites
```bash
# Install Go 1.21+
# macOS
brew install go

# Linux
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Verify installation
go version  # Should show 1.21+
```

### Build Steps

```bash
# 1. Clone repository
git clone https://github.com/deepaucksharma/mcp-server-newrelic.git
cd mcp-server-newrelic

# 2. Download dependencies
go mod download

# 3. Build the server
make build

# 4. Verify build
./bin/mcp-server --version

# 5. Install to system (optional)
sudo make install  # Installs to /usr/local/bin
```

### Development Build
```bash
# Build with debug symbols
make build-debug

# Build for all platforms
make build-all

# Run tests before building
make test && make build
```

## 🤖 Claude Desktop Integration

### Step 1: Locate Configuration File

Find your Claude Desktop configuration:

| OS | Location |
|----|----------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

### Step 2: Configure MCP Server

#### Docker Configuration
```json
{
  "mcpServers": {
    "newrelic": {      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-v", "/path/to/.env:/app/.env:ro",
        "mcp-server-newrelic:latest"
      ]
    }
  }
}
```

#### Binary Configuration
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "/usr/local/bin/mcp-server",
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-your-key-here",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id",
        "NEW_RELIC_REGION": "US"
      }
    }
  }
}
```

#### Development Configuration
```json
{
  "mcpServers": {
    "newrelic": {
      "command": "go",
      "args": ["run", "/path/to/mcp-server-newrelic/cmd/server/main.go"],
      "env": {
        "NEW_RELIC_API_KEY": "NRAK-your-key-here",
        "NEW_RELIC_ACCOUNT_ID": "your-account-id"
      }
    }
  }
}
```
## ⚙️ Configuration

### Environment Variables

Create a `.env` file with required settings:

```env
# Required
NEW_RELIC_API_KEY=NRAK-your-key-here
NEW_RELIC_ACCOUNT_ID=your-account-id

# Optional
NEW_RELIC_REGION=US          # US or EU (default: US)
LOG_LEVEL=info               # debug, info, warn, error (default: info)
MCP_TRANSPORT=stdio          # stdio, http, sse (default: stdio)
HTTP_PORT=8080               # HTTP server port (default: 8080)
STATE_STORE=memory           # memory or redis (default: memory)
REDIS_URL=redis://localhost  # If using Redis state store
CACHE_TTL=300                # Cache TTL in seconds (default: 300)
MOCK_MODE=false              # Enable mock mode (default: false)
```

### Configuration File Locations

The server checks for configuration in this order:
1. Command line flags
2. Environment variables
3. `.env` file in current directory
4. `~/.config/mcp-newrelic/.env`
5. `/etc/mcp-newrelic/.env`

## ✅ Verification

### Health Check
```bash
# HTTP mode
curl http://localhost:8080/health

# STDIO mode
echo '{"jsonrpc":"2.0","method":"health","id":1}' | mcp-server
```
### Diagnostics
```bash
# Run built-in diagnostics
make diagnose

# Or with binary
mcp-server --diagnose
```

### Test Query
```bash
# Test basic functionality
echo '{"jsonrpc":"2.0","method":"tools/list","id":1}' | mcp-server
```

## 🔄 Upgrading

### Docker Upgrade
```bash
# Pull latest image
docker-compose pull

# Recreate containers
docker-compose up -d
```

### Binary Upgrade
```bash
# Download new version
wget https://github.com/deepaucksharma/mcp-server-newrelic/releases/latest/download/mcp-server-linux-amd64

# Replace old binary
sudo mv mcp-server-linux-amd64 /usr/local/bin/mcp-server
sudo chmod +x /usr/local/bin/mcp-server

# Verify version
mcp-server --version
```

### Source Upgrade
```bash
# Pull latest changes
git pull origin main

# Rebuild
make clean
make build
```
## 🗑️ Uninstallation

### Docker Uninstall
```bash
# Stop and remove containers
docker-compose down

# Remove images
docker rmi mcp-server-newrelic:latest

# Remove volumes (caution: deletes data)
docker volume prune
```

### Binary Uninstall
```bash
# Remove binary
sudo rm /usr/local/bin/mcp-server

# Remove configuration
rm -rf ~/.config/mcp-newrelic
```

### Source Uninstall
```bash
# If installed with make install
sudo make uninstall

# Remove repository
cd ..
rm -rf mcp-server-newrelic
```

## 🔧 Troubleshooting Installation

### Docker Issues

**"Cannot connect to Docker daemon"**
```bash
# Start Docker service
sudo systemctl start docker

# Add user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

**"Port already in use"**
```bash
# Find process using port
lsof -i :8080

# Change port in docker-compose.yml
```

### Binary Issues

**"Permission denied"**
```bash
# Make executable
chmod +x mcp-server

# Check file permissions
ls -la mcp-server
```

**"Command not found"**
```bash
# Add to PATH
export PATH=$PATH:/path/to/mcp-server

# Or move to standard location
sudo mv mcp-server /usr/local/bin/
```

### Build Issues

**"Go version too old"**
```bash
# Update Go
brew upgrade go  # macOS
# Or download from https://go.dev/dl/
```

**"Missing dependencies"**
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

## 📚 Next Steps

- [Configuration Guide](03_CONFIGURATION.md) - Detailed configuration options
- [Getting Started](01_GETTING_STARTED.md) - First queries
- [Claude Integration](41_GUIDE_CLAUDE_INTEGRATION.md) - AI assistant setup

---

**Need help?** Check the [FAQ](09_FAQ.md) or [open an issue](https://github.com/deepaucksharma/mcp-server-newrelic/issues).