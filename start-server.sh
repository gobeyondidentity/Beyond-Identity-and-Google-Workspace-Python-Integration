#!/bin/bash
# Start the SCIM sync server
# Make sure to set BI_API_TOKEN environment variable before running this script
# Example: export BI_API_TOKEN="your-token-here"

if [ -z "$BI_API_TOKEN" ]; then
    echo "Error: BI_API_TOKEN environment variable is not set"
    echo "Please set it before running the server:"
    echo "export BI_API_TOKEN=\"your-beyond-identity-api-token\""
    exit 1
fi

echo "Starting SCIM sync server..."
./go-scim-sync server