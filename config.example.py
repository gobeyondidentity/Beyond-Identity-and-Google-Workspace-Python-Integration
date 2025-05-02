#!/usr/bin/env python3
"""
Example configuration settings for Google Workspace and Beyond Identity synchronization.

This module contains all configuration settings required for the synchronization
process between Google Workspace and Beyond Identity. It includes settings for
authentication, domain configuration, and group synchronization.

Note: 
1. Copy this file to config.py
2. Replace the placeholder values with your actual configuration
3. Never commit config.py to version control
"""

from typing import List
import os
from pathlib import Path

# =============================================================================
# Google Workspace Configuration
# =============================================================================

# Domain and Admin Settings
GWS_DOMAIN_NAME: str = "your-domain.com"
GWS_SUPER_ADMIN_EMAIL: str = "admin@your-domain.com"

# Service Account Configuration
SERVICE_ACCOUNT_KEY_PATH: str = "path-to-your-service-account.json"

# =============================================================================
# Beyond Identity Configuration
# =============================================================================

# API Token - Store this securely in environment variables or a secrets manager
BI_TENANT_API_TOKEN: str = os.getenv(
    "BI_TENANT_API_TOKEN",
    "your_api_token_here"  # Replace with actual token or use environment variable
)

# Group Configuration
# This prefix will be added to all groups created in Beyond Identity that are synced from Google Workspace.
# For example, if a Google group is named "engineering@domain.com", it will appear in Beyond Identity
# as "GoogleSCIM_engineering@domain.com". This helps administrators easily identify which groups
# originated from Google Workspace and are being managed by this sync process.
BI_GROUP_PREFIX: str = "GoogleSCIM_"

# Logging Configuration
# Enable or disable logging to file
LOGGING_ENABLED: bool = False

# Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
LOGGING_LEVEL: str = "INFO"

# Directory for log files (relative to script location)
LOGGING_DIR: str = "logs"

# Log file name format (strftime format string)
# Default: sync_YYYYMMDD_HHMMSS.log
LOGGING_FILE_FORMAT: str = "sync_%Y%m%d_%H%M%S.log"

# Test Mode Configuration
# When enabled, the script will only show what actions it would take without actually performing them
TEST_MODE: bool = True

# =============================================================================
# Synchronization Groups
# =============================================================================

# List of Google Workspace groups to synchronize
GWS_GROUPS: List[str] = [
    "group1@your-domain.com",
    "group2@your-domain.com",
    # Add more groups here as needed
]

# =============================================================================
# Validation Functions
# =============================================================================

def validate_config() -> None:
    """
    Validate the configuration settings.
    
    Raises:
        ValueError: If any required configuration is missing or invalid.
    """
    if not GWS_DOMAIN_NAME:
        raise ValueError("GWS_DOMAIN_NAME must be set")
    
    if not GWS_SUPER_ADMIN_EMAIL:
        raise ValueError("GWS_SUPER_ADMIN_EMAIL must be set")
    
    if not Path(SERVICE_ACCOUNT_KEY_PATH).exists():
        raise ValueError(f"Service account key file not found: {SERVICE_ACCOUNT_KEY_PATH}")
    
    if not BI_TENANT_API_TOKEN or BI_TENANT_API_TOKEN == "your_api_token_here":
        raise ValueError("BI_TENANT_API_TOKEN must be set")
    
    if not GWS_GROUPS:
        raise ValueError("At least one group must be specified in GWS_GROUPS")

# Validate configuration on import
validate_config() 