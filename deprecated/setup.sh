#!/bin/bash

# Exit on error
set -e

echo "Setting up Google Workspace - Beyond Identity Sync..."

# Create virtual environment if it doesn't exist
if [ ! -d "venv" ]; then
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# Activate virtual environment
echo "Activating virtual environment..."
source venv/bin/activate

# Install requirements
echo "Installing dependencies..."
pip install -r requirements.txt

# Create necessary directories if they don't exist
mkdir -p logs

echo "Setup complete! To start using the sync tool:"
echo "1. Configure your settings in config.py"
echo "2. Place your Google service account key in the root directory"
echo "3. Activate the virtual environment: source venv/bin/activate"
echo "4. Run the script: python gwbisync.py" 