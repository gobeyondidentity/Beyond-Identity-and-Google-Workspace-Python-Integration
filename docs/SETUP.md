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
- Sync settings (groups to sync, enrollment group configuration, retry settings)
- Server mode settings (port, scheduling)

### 2. Configure API Token
After the wizard, ensure your API token is configured in the `config.yaml` file under `beyond_identity.api_token`.

### 3. Validate Setup
Test your configuration:

```bash
./scim-sync setup validate
```

### 4. Run Your First Sync
Execute a one-time sync:

```bash
./scim-sync run
```

### 5. Start Server Mode (Optional)
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
  enrollment_group_email: "byid-enrolled@yourcompany.com"  # Optional: Google group for BI enrolled users
  enrollment_group_name: "BYID Enrolled"                   # Optional: Display name for enrollment group
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

1. **Environment Variables**
   - Always use environment variables for sensitive data
   - Never commit API tokens to version control

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

## Bi-directional Sync Configuration

The application supports bi-directional synchronization:

### GWS → BI Sync (Standard)
- Creates users and groups in Beyond Identity
- Syncs group memberships from Google Workspace
- Handles user lifecycle (creation, updates)

### BI → GWS Sync (Enrollment Status)
- Monitors Beyond Identity user activation status
- Automatically manages enrollment group membership in Google Workspace
- Adds users to enrollment group when they activate in Beyond Identity
- Removes users from enrollment group when they become inactive

### Enrollment Group Configuration

The enrollment group settings are optional. If not configured, defaults will be used:

```yaml
sync:
  enrollment_group_email: "byid-enrolled@yourcompany.com"
  enrollment_group_name: "BYID Enrolled"
```

**Default behavior:**
- If `enrollment_group_email` is not specified, it defaults to `byid-enrolled@{domain}`
- If `enrollment_group_name` is not specified, it defaults to `"BYID Enrolled"`
- The enrollment group will be created automatically if it doesn't exist

### Enrollment Sync Process

During each sync operation:

1. **Check User Status**: Query Beyond Identity for activation status of all users in sync scope
2. **Compare States**: Compare BI status with current enrollment group membership
3. **Update Membership**: 
   - Add newly activated users to enrollment group
   - Remove deactivated users from enrollment group
4. **Logging**: All enrollment changes are logged for audit purposes

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
