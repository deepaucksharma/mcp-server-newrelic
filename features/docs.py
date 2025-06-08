import json
import logging
import os
import sys
from typing import Dict, Any
from fastmcp import FastMCP

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from core.plugin_loader import PluginBase

logger = logging.getLogger(__name__)


class DocsPlugin(PluginBase):
    """Tools for searching and retrieving New Relic documentation"""

    metadata = {
        "name": "DocsPlugin",
        "version": "1.0.0",
        "description": "Search the New Relic docs repository",
        "dependencies": [],
        "required_services": ["docs_cache"],
        "provides_services": [],
        "enabled": True,
        "priority": 80,
    }

    @staticmethod
    def register(app: FastMCP, services: Dict[str, Any]):
        docs_cache = services.get("docs_cache")

        @app.tool()
        async def search_docs(keyword: str, limit: int = 5) -> str:
            """Search the local New Relic docs for a keyword"""
            if not keyword or not docs_cache:
                return json.dumps([])
            results = docs_cache.search(keyword, limit=limit)
            return json.dumps(results, indent=2)

        @app.resource("newrelic://docs/{path}")
        async def get_doc_content(path: str) -> str:
            """Get the content of a documentation file"""
            if not path or not docs_cache:
                return ""
            return docs_cache.get_content(path)

