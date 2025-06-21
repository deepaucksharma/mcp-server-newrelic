#!/usr/bin/env python3
"""
Seed data script for DISC-MISS-001 scenario
Creates Transaction events with intentionally missing attributes
"""

import os
import sys
import time
import random
import requests
from datetime import datetime, timedelta

# Configuration from environment
ACCOUNT_ID = os.environ.get('E2E_PRIMARY_ACCOUNT_ID')
API_KEY = os.environ.get('E2E_PRIMARY_API_KEY')
MISSING_ATTRS = os.environ.get('SEED_MISSING_ATTRS', '').split(',')
EVENT_COUNT = int(os.environ.get('EVENT_COUNT', '1000'))

# New Relic Event API endpoint
EVENTS_URL = f'https://insights-collector.newrelic.com/v1/accounts/{ACCOUNT_ID}/events'

def generate_transaction_event(timestamp, index):
    """Generate a Transaction event with some attributes missing"""
    event = {
        'eventType': 'Transaction',
        'timestamp': int(timestamp.timestamp()),
        'appName': 'mcp-test-app',
        'transactionName': f'WebTransaction/Test/endpoint_{index % 10}',
        'e2e_test': True,  # Flag for cleanup
        'test_scenario': 'DISC-MISS-001'
    }
    
    # Add attributes that might be missing
    if 'duration' not in MISSING_ATTRS:
        event['duration'] = random.uniform(0.05, 2.0)  # 50ms to 2s
    
    if 'error' not in MISSING_ATTRS:
        event['error'] = random.choice([True, False, False, False])  # 25% errors
    
    if 'error.message' not in MISSING_ATTRS and event.get('error'):
        event['error.message'] = random.choice([
            'Database connection timeout',
            'Service unavailable',
            'Invalid request format'
        ])
    
    if 'http.statusCode' not in MISSING_ATTRS:
        if event.get('error'):
            event['http.statusCode'] = random.choice([400, 404, 500, 503])
        else:
            event['http.statusCode'] = 200
    
    if 'request.method' not in MISSING_ATTRS:
        event['request.method'] = random.choice(['GET', 'POST', 'PUT', 'DELETE'])
    
    if 'http.method' not in MISSING_ATTRS:
        # Alternative attribute that might exist
        event['http.method'] = event.get('request.method', 'GET')
    
    # Add some attributes that are always present
    event['host'] = f'server-{index % 5}.example.com'
    event['aws.region'] = random.choice(['us-east-1', 'us-west-2'])
    
    # Add high-cardinality attributes for testing
    event['trace.id'] = f'trace-{index}-{random.randint(1000, 9999)}'
    event['userId'] = f'user-{index % 100}'
    
    return event

def send_events_batch(events):
    """Send a batch of events to New Relic"""
    headers = {
        'Api-Key': API_KEY,
        'Content-Type': 'application/json'
    }
    
    try:
        response = requests.post(EVENTS_URL, json=events, headers=headers)
        response.raise_for_status()
        return True
    except Exception as e:
        print(f"Error sending events: {e}")
        return False

def main():
    if not ACCOUNT_ID or not API_KEY:
        print("ERROR: E2E_PRIMARY_ACCOUNT_ID and E2E_PRIMARY_API_KEY must be set")
        sys.exit(1)
    
    print(f"Seeding {EVENT_COUNT} Transaction events with missing attributes: {MISSING_ATTRS}")
    print(f"Target account: {ACCOUNT_ID}")
    
    # Generate events over the last hour
    now = datetime.utcnow()
    start_time = now - timedelta(hours=1)
    
    events = []
    batch_size = 100
    success_count = 0
    
    for i in range(EVENT_COUNT):
        # Distribute events over time
        timestamp = start_time + timedelta(
            seconds=(i / EVENT_COUNT) * 3600
        )
        
        event = generate_transaction_event(timestamp, i)
        events.append(event)
        
        # Send in batches
        if len(events) >= batch_size:
            if send_events_batch(events):
                success_count += len(events)
                print(f"Sent {success_count}/{EVENT_COUNT} events...")
            events = []
    
    # Send remaining events
    if events and send_events_batch(events):
        success_count += len(events)
    
    print(f"Successfully sent {success_count}/{EVENT_COUNT} events")
    print(f"Missing attributes: {MISSING_ATTRS}")
    print("Seed data generation complete!")

if __name__ == '__main__':
    main()