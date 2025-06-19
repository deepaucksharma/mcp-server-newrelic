#!/bin/bash

# First check if we have credentials
if [ -z "$NEW_RELIC_API_KEY" ] || [ -z "$NEW_RELIC_ACCOUNT_ID" ]; then
    echo "Please set NEW_RELIC_API_KEY and NEW_RELIC_ACCOUNT_ID environment variables"
    echo "Example:"
    echo "  export NEW_RELIC_API_KEY='your-api-key'"
    echo "  export NEW_RELIC_ACCOUNT_ID='your-account-id'"
    exit 1
fi

# Build and run the test
echo "Building project..."
make build

echo "Running end-to-end test with real NRDB..."
./test_e2e_real.sh