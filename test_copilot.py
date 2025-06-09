"""
This is a simple test file to verify that GitHub Copilot is working with the New Relic MCP server.
Type some comments below and see if GitHub Copilot provides completions based on your New Relic data.
"""

# Function to query New Relic APM data
def get_apm_metrics(app_name, time_range="1 hour"):
    """
    Get APM metrics for a specific application
    """
    # Try NRQL: SELECT average(duration) FROM Transaction WHERE appName = 

# Function to check New Relic alerts
def get_current_incidents(priority=None):
    """
    Get all open incidents from New Relic Alerts
    """
    # Implementation using the alerts.list_open_incidents tool

# Parse New Relic logs for errors
def search_logs_for_errors(query, since="30 minutes ago"):
    """
    Search New Relic logs for specific error patterns
    """
    # NRQL: SELECT * FROM Log WHERE message CONTAINS 

# Get infrastructure metrics from New Relic
def get_host_metrics(hostname=None):
    """
    Get infrastructure metrics for a specific host or all hosts
    """
    # NRQL: SELECT average(cpuPercent), average(memoryUsedPercent) FROM SystemSample WHERE 

# Create a dashboard for New Relic data
def create_dashboard(title, widgets=[]):
    """
    Create a New Relic dashboard with the specified widgets
    """
    # NerdGraph query to create dashboard
