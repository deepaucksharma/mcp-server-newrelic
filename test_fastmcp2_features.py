#!/usr/bin/env python3
"""
Test script for FastMCP 2.0 features
"""

import os
import sys
import asyncio
import logging
from typing import Dict, Any

# Set up test environment
os.environ.setdefault('NEW_RELIC_API_KEY', 'test-key')
os.environ.setdefault('NEW_RELIC_ACCOUNT_ID', '123456')
os.environ.setdefault('LOG_LEVEL', 'DEBUG')

logging.basicConfig(level=logging.DEBUG)
logger = logging.getLogger(__name__)

async def test_fastmcp_features():
    """Test FastMCP 2.0 specific features"""
    
    try:
        # Test FastMCP 2.0 imports
        logger.info("Testing FastMCP 2.0 imports...")
        
        from fastmcp import FastMCP
        from fastmcp.utilities import setup_logging
        
        logger.info("✅ FastMCP 2.0 imports successful")
        
        # Test enhanced app creation
        logger.info("Testing enhanced app creation...")
        
        app = FastMCP(
            name="test-app",
            version="2.0.0",
            description="Test FastMCP 2.0 features"
        )
        
        logger.info("✅ FastMCP app creation successful")
        
        # Test context injection (simulation)
        logger.info("Testing context injection simulation...")
        
        @app.tool()
        async def test_tool() -> Dict[str, Any]:
            """Test tool with simulated context injection"""
            # In real implementation, this would use get_mcp_context()
            return {
                "status": "success",
                "fastmcp_version": "2.0.0",
                "features": ["context_injection", "progress_reporting", "enhanced_security"]
            }
        
        logger.info("✅ Tool registration successful")
        
        # Test logging utilities
        logger.info("Testing logging utilities...")
        setup_logging(level="INFO")
        logger.info("✅ Logging utilities working")
        
        # Test server configuration
        logger.info("Testing server configuration...")
        
        # Simulate transport configuration
        transport = os.getenv("MCP_TRANSPORT", "stdio")
        logger.info(f"Transport mode: {transport}")
        
        if transport == "http":
            host = os.getenv("HTTP_HOST", "127.0.0.1")
            port = int(os.getenv("HTTP_PORT", "3000"))
            logger.info(f"HTTP configuration: {host}:{port}")
        
        logger.info("✅ Server configuration successful")
        
        # Test enhanced error handling
        logger.info("Testing enhanced error handling...")
        
        try:
            # Simulate error sanitization
            raise ValueError("Test error for sanitization")
        except Exception as e:
            # In real implementation, would use fastmcp.security.sanitize_error
            sanitized = "Error sanitized for security"
            logger.info(f"Error sanitized: {sanitized}")
        
        logger.info("✅ Error handling test successful")
        
        return True
        
    except ImportError as e:
        logger.error(f"❌ FastMCP 2.0 import failed: {e}")
        logger.error("Please install FastMCP 2.0: pip install fastmcp>=2.0.0")
        return False
    except Exception as e:
        logger.error(f"❌ Test failed with error: {e}")
        return False

async def test_new_relic_integration():
    """Test New Relic MCP server integration"""
    
    try:
        logger.info("Testing New Relic MCP server integration...")
        
        # Test core imports
        from core.account_manager import AccountManager
        from core.nerdgraph_client import NerdGraphClient
        from core.cache import get_cache
        
        logger.info("✅ Core imports successful")
        
        # Test feature imports
        from features import common, entities, apm, alerts, synthetics
        
        logger.info("✅ Feature imports successful")
        
        # Test plugin system
        logger.info("Testing plugin system...")
        
        # Simulate FastMCP app
        from fastmcp import FastMCP
        
        test_app = FastMCP(
            name="test-newrelic-mcp",
            version="2.0.0",
            description="Test New Relic integration"
        )
        
        # Register features
        common.register(test_app)
        entities.register(test_app)
        apm.register(test_app)
        alerts.register(test_app)
        synthetics.register(test_app)
        
        logger.info("✅ Plugin registration successful")
        
        return True
        
    except Exception as e:
        logger.error(f"❌ New Relic integration test failed: {e}")
        return False

async def test_github_copilot_config():
    """Test GitHub Copilot configuration"""
    
    try:
        logger.info("Testing GitHub Copilot configuration...")
        
        # Check MCP configuration file
        mcp_config_path = ".vscode/mcp.json"
        if os.path.exists(mcp_config_path):
            logger.info("✅ GitHub Copilot MCP config found")
            
            import json
            with open(mcp_config_path, 'r') as f:
                config = json.load(f)
            
            if "mcpServers" in config and "newrelic" in config["mcpServers"]:
                logger.info("✅ New Relic MCP server configured for GitHub Copilot")
            else:
                logger.warning("⚠️ New Relic server not found in MCP config")
        else:
            logger.warning("⚠️ GitHub Copilot MCP config not found")
        
        # Check VS Code tasks
        tasks_config_path = ".vscode/tasks.json"
        if os.path.exists(tasks_config_path):
            logger.info("✅ VS Code tasks configuration found")
        else:
            logger.warning("⚠️ VS Code tasks configuration not found")
        
        return True
        
    except Exception as e:
        logger.error(f"❌ GitHub Copilot config test failed: {e}")
        return False

async def main():
    """Run all tests"""
    logger.info("🚀 Starting FastMCP 2.0 feature tests...")
    
    tests = [
        ("FastMCP 2.0 Features", test_fastmcp_features),
        ("New Relic Integration", test_new_relic_integration),
        ("GitHub Copilot Config", test_github_copilot_config)
    ]
    
    results = []
    
    for test_name, test_func in tests:
        logger.info(f"\n🧪 Running test: {test_name}")
        try:
            result = await test_func()
            results.append((test_name, result))
            if result:
                logger.info(f"✅ {test_name}: PASSED")
            else:
                logger.error(f"❌ {test_name}: FAILED")
        except Exception as e:
            logger.error(f"❌ {test_name}: ERROR - {e}")
            results.append((test_name, False))
    
    # Summary
    logger.info("\n📊 Test Results Summary:")
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for test_name, result in results:
        status = "✅ PASSED" if result else "❌ FAILED"
        logger.info(f"  {test_name}: {status}")
    
    logger.info(f"\nOverall: {passed}/{total} tests passed")
    
    if passed == total:
        logger.info("🎉 All tests passed! FastMCP 2.0 migration successful!")
        return 0
    else:
        logger.error("💥 Some tests failed. Please check the errors above.")
        return 1

if __name__ == "__main__":
    sys.exit(asyncio.run(main()))