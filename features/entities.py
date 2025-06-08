import json
from typing import List, Optional, Dict, Any
from fastmcp import FastMCP
import logging
import sys
import os

# Add parent directory to path for imports
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from core.plugin_loader import PluginBase

logger = logging.getLogger(__name__)


class EntitiesPlugin(PluginBase):
    """Entity search and management plugin"""

    @staticmethod
    def register(app: FastMCP, services: Dict[str, Any]):
        nerdgraph = services["nerdgraph"]
        session_manager = services.get("session_manager")
        default_account_id = services.get("account_id")

        @app.tool()
        async def search_entities(
            name: Optional[str] = None,
            entity_type: Optional[str] = None,
            domain: Optional[str] = None,
            tags: Optional[List[Dict[str, str]]] = None,
            target_account_id: Optional[int] = None,
            limit: int = 50,
        ) -> str:
            """Search New Relic entities by various criteria"""
            if not any([name, entity_type, domain, tags, target_account_id]):
                return json.dumps({"errors": [{"message": "At least one search criterion must be provided."}]})

            account_to_use = target_account_id or default_account_id
            conditions = []
            if name:
                escaped = name.replace("'", "\\'")
                conditions.append(f"name LIKE '%{escaped}%'")
            if entity_type:
                conditions.append(f"type = '{entity_type}'")
            if domain:
                conditions.append(f"domain = '{domain}'")
            if tags:
                for tag in tags:
                    if isinstance(tag, dict) and "key" in tag and "value" in tag:
                        val = str(tag["value"]).replace("'", "\\'")
                        conditions.append(f"tags.`{tag['key']}` = '{val}'")
            if account_to_use:
                conditions.append(f"accountId = {account_to_use}")

            search_query = " AND ".join(conditions)
            query = """
            query($searchQuery: String!, $limit: Int) {
              actor {
                entitySearch(query: $searchQuery, options: {limit: $limit}) {
                  results { entities { guid name entityType domain } nextCursor }
                  count
                }
              }
            }
            """
            try:
                variables = {"searchQuery": search_query, "limit": limit}
                result = await nerdgraph.query(query, variables)
                return json.dumps(result, indent=2)
            except Exception as e:
                logger.error(f"Entity search failed: {e}")
                return json.dumps({"errors": [{"message": str(e)}]})

        @app.resource("newrelic://entity/{guid}")
        async def get_entity_details(guid: str) -> str:
            """Get details for a specific entity"""
            if not guid or not isinstance(guid, str):
                return json.dumps({"errors": [{"message": "Valid entity GUID must be provided."}]})

            query = """
            query($guid: EntityGuid!) {
              actor {
                entity(guid: $guid) {
                  guid
                  name
                  entityType
                  reporting
                }
              }
            }
            """
            try:
                result = await nerdgraph.query(query, {"guid": guid})
                if session_manager:
                    session = session_manager.get_or_create_session()
                    session.cache_entity(guid, result)
                return json.dumps(result, indent=2)
            except Exception as e:
                logger.error(f"Failed to get entity details: {e}")
                return json.dumps({"errors": [{"message": str(e)}]})
