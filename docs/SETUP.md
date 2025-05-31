# Go SCIM Sync Setup Guide

## Quick Start

### 1. Run the Setup Wizard
The easiest way to get started is using the interactive setup wizard:

```bash
./scim-sync setup wizard
```

This will guide you through:
- Application configuration (log level, test mode)
- Google Workspace setup (domain, admin email, service account)
- Beyond Identity configuration (API token, endpoints)
- Sync settings (groups to sync, retry configuration)
- Server mode settings (port, scheduling)

### 2. Validate Setup
Test your configuration:

```bash
./scim-sync setup validate
```

### 3. Run Your First Sync
Execute a one-time sync:

```bash
./scim-sync run
```

### 4. Start Server Mode (Optional)
For continuous operation with HTTP API:

```bash
./scim-sync server
```

## Manual Configuration

If you prefer to create the configuration manually, create a `config.yaml` file:

```yaml
# Application settings
app:
  log_level: "info"
  test_mode: true

# Google Workspace configuration
google_workspace:
  domain: "yourcompany.com"
  super_admin_email: "admin@yourcompany.com"
  service_account_key_path: "./service-account.json"

# Beyond Identity configuration  
beyond_identity:
  api_token: "your-beyond-identity-api-token"
  scim_base_url: "https://api.byndid.com/scim/v2"
  native_api_url: "https://api.byndid.com/v2"
  group_prefix: "GoogleSCIM_"

# Synchronization settings
sync:
  groups:
    - "engineering@yourcompany.com"
    - "sales@yourcompany.com"
  retry_attempts: 3
  retry_delay_seconds: 30

# Server mode settings
server:
  port: 8080
  schedule_enabled: false
  schedule: "0 */6 * * *"
```

## Prerequisites

### Google Workspace Setup

1. **Create a Google Cloud Project**
   - Go to [Google Cloud Console](https://console.cloud.google.com)
   - Create a new project or select existing one

2. **Enable Admin SDK API**
   - Navigate to APIs & Services > Library
   - Search for "Admin SDK API"
   - Enable the API

3. **Create Service Account**
   - Go to APIs & Services > Credentials
   - Click "Create Credentials" > "Service Account"
   - Fill in the details and create

4. **Generate Service Account Key**
   - Click on your service account
   - Go to "Keys" tab
   - Click "Add Key" > "Create new key"
   - Choose JSON format and download

5. **Enable Domain-wide Delegation**
   - In service account settings, check "Enable domain-wide delegation"
   - Note the Client ID for the next step

6. **Configure Domain-wide Delegation in Google Workspace**
   - Go to [Google Admin Console](https://admin.google.com)
   - Navigate to Security > API Controls > Domain-wide Delegation
   - Add new API client with:
     - Client ID: (from service account)
     - OAuth Scopes: 
       - `https://www.googleapis.com/auth/admin.directory.user`
       - `https://www.googleapis.com/auth/admin.directory.group`
       - `https://www.googleapis.com/auth/admin.directory.group.member`

### Beyond Identity Setup

1. **Get API Token**
   - Log into Beyond Identity Admin Console
   - Navigate to Applications > API Tokens
   - Create new token with SCIM permissions

2. **Note Your SCIM Endpoint**
   - Typically: `https://api.byndid.com/scim/v2`
   - Check your tenant configuration if different

## Security Best Practices

1. **Configuration Security**
   - Never commit API tokens to version control
   - Store config.yaml securely with appropriate file permissions
   - Consider using encrypted storage for production deployments

2. **Service Account Security**
   - Store service account files securely
   - Use minimal required permissions
   - Rotate keys regularly

3. **Test Mode**
   - Always test with `test_mode: true` first
   - Validate sync results before enabling actual changes

4. **Monitoring**
   - Use server mode for monitoring and metrics
   - Set up alerts for sync failures
   - Monitor API rate limits

## Common Configurations

### Development Setup
```yaml
app:
  log_level: "debug"
  test_mode: true

server:
  schedule_enabled: false  # Manual sync only
```

### Production Setup
```yaml
app:
  log_level: "info"
  test_mode: false

server:
  schedule_enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours
```

### High-frequency Sync
```yaml
server:
  schedule_enabled: true
  schedule: "0 */1 * * *"  # Every hour

sync:
  retry_attempts: 5
  retry_delay_seconds: 60
```
