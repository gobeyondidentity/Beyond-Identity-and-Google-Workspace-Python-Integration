package setup

import (
	"fmt"
	"os"
	"path/filepath"
)

// GenerateDocumentation creates setup documentation files
func GenerateDocumentation(outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate setup guide
	if err := generateSetupGuide(filepath.Join(outputDir, "SETUP.md")); err != nil {
		return err
	}

	// Generate API documentation
	if err := generateAPIGuide(filepath.Join(outputDir, "API.md")); err != nil {
		return err
	}

	// Generate troubleshooting guide
	if err := generateTroubleshootingGuide(filepath.Join(outputDir, "TROUBLESHOOTING.md")); err != nil {
		return err
	}

	fmt.Printf("✅ Documentation generated in %s\n", outputDir)
	return nil
}

func generateSetupGuide(path string) error {
	content := `# Go SCIM Sync Setup Guide

## Quick Start

### 1. Run the Setup Wizard
The easiest way to get started is using the interactive setup wizard:

` + "```bash" + `
./scim-sync setup wizard
` + "```" + `

This will guide you through:
- Application configuration (log level, test mode)
- Google Workspace setup (domain, admin email, service account)
- Beyond Identity configuration (API token, endpoints)
- Sync settings (groups to sync, retry configuration)
- Server mode settings (port, scheduling)

### 2. Validate Setup
Test your configuration:

` + "```bash" + `
./scim-sync setup validate
` + "```" + `

### 3. Run Your First Sync
Execute a one-time sync:

` + "```bash" + `
./scim-sync run
` + "```" + `

### 4. Start Server Mode (Optional)
For continuous operation with HTTP API:

` + "```bash" + `
./scim-sync server
` + "```" + `

## Manual Configuration

If you prefer to create the configuration manually, create a ` + "`config.yaml`" + ` file:

` + "```yaml" + `
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
` + "```" + `

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
       - ` + "`https://www.googleapis.com/auth/admin.directory.user`" + `
       - ` + "`https://www.googleapis.com/auth/admin.directory.group`" + `
       - ` + "`https://www.googleapis.com/auth/admin.directory.group.member`" + `

### Beyond Identity Setup

1. **Get API Token**
   - Log into Beyond Identity Admin Console
   - Navigate to Applications > API Tokens
   - Create new token with SCIM permissions

2. **Note Your SCIM Endpoint**
   - Typically: ` + "`https://api.byndid.com/scim/v2`" + `
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
   - Always test with ` + "`test_mode: true`" + ` first
   - Validate sync results before enabling actual changes

4. **Monitoring**
   - Use server mode for monitoring and metrics
   - Set up alerts for sync failures
   - Monitor API rate limits

## Common Configurations

### Development Setup
` + "```yaml" + `
app:
  log_level: "debug"
  test_mode: true

server:
  schedule_enabled: false  # Manual sync only
` + "```" + `

### Production Setup
` + "```yaml" + `
app:
  log_level: "info"
  test_mode: false

server:
  schedule_enabled: true
  schedule: "0 */6 * * *"  # Every 6 hours
` + "```" + `

### High-frequency Sync
` + "```yaml" + `
server:
  schedule_enabled: true
  schedule: "0 */1 * * *"  # Every hour

sync:
  retry_attempts: 5
  retry_delay_seconds: 60
` + "```" + `
`

	return os.WriteFile(path, []byte(content), 0644)
}

func generateAPIGuide(path string) error {
	content := `# Go SCIM Sync API Reference

When running in server mode (` + "`./scim-sync server`" + `), the application provides an HTTP API for management and monitoring.

## Base URL

By default, the server runs on port 8080:
` + "```" + `
http://localhost:8080
` + "```" + `

## Endpoints

### Health Check
` + "```http" + `
GET /health
` + "```" + `

Returns server health status and next scheduled sync time.

**Response Example:**
` + "```json" + `
{
  "status": "healthy",
  "version": "0.1.0",
  "timestamp": "2024-01-15T10:30:00Z",
  "services": {
    "google_workspace": "ok",
    "beyond_identity": "ok"
  },
  "last_sync": "2024-01-15T10:00:00Z",
  "next_sync": "2024-01-15T16:00:00Z",
  "sync_enabled": true
}
` + "```" + `

### Manual Sync
` + "```http" + `
POST /sync
` + "```" + `

Triggers a manual synchronization operation.

**Response Example:**
` + "```json" + `
{
  "status": "success",
  "message": "Sync operation completed",
  "timestamp": "2024-01-15T10:30:00Z",
  "result": {
    "groups_processed": 3,
    "users_created": 5,
    "users_updated": 2,
    "groups_created": 1,
    "memberships_added": 7,
    "memberships_removed": 1,
    "duration": 5420000000,
    "errors": null
  }
}
` + "```" + `

### Metrics
` + "```http" + `
GET /metrics
` + "```" + `

Returns synchronization metrics and statistics.

**Response Example:**
` + "```json" + `
{
  "total_syncs": 25,
  "successful_syncs": 24,
  "failed_syncs": 1,
  "success_rate": 96.0,
  "total_users_created": 150,
  "total_users_updated": 45,
  "total_groups_created": 8,
  "total_groups_processed": 75,
  "total_memberships_added": 200,
  "total_memberships_removed": 15,
  "last_sync_duration": 5420000000,
  "average_sync_duration": 4890000000,
  "last_sync_time": "2024-01-15T10:00:00Z",
  "uptime": 86400000000000
}
` + "```" + `

### Version Information
` + "```http" + `
GET /version
` + "```" + `

Returns application version information.

**Response Example:**
` + "```json" + `
{
  "version": "0.1.0",
  "build_time": "2024-01-15T08:00:00Z",
  "mode": "server"
}
` + "```" + `

### Scheduler Control

#### Start Scheduler
` + "```http" + `
POST /scheduler/start
` + "```" + `

Starts the automatic sync scheduler (if configured).

#### Stop Scheduler
` + "```http" + `
POST /scheduler/stop
` + "```" + `

Stops the automatic sync scheduler.

#### Scheduler Status
` + "```http" + `
GET /scheduler/status
` + "```" + `

Returns scheduler status and configuration.

**Response Example:**
` + "```json" + `
{
  "running": true,
  "schedule": "0 */6 * * *",
  "last_sync": "2024-01-15T10:00:00Z",
  "next_sync": "2024-01-15T16:00:00Z"
}
` + "```" + `

## Error Responses

All endpoints return appropriate HTTP status codes:

- ` + "`200`" + ` - Success
- ` + "`400`" + ` - Bad Request
- ` + "`500`" + ` - Internal Server Error

Error response format:
` + "```json" + `
{
  "error": "Error description",
  "details": "Additional error details if available"
}
` + "```" + `

## cURL Examples

### Check Health
` + "```bash" + `
curl http://localhost:8080/health
` + "```" + `

### Trigger Manual Sync
` + "```bash" + `
curl -X POST http://localhost:8080/sync
` + "```" + `

### Get Metrics
` + "```bash" + `
curl http://localhost:8080/metrics
` + "```" + `

### Control Scheduler
` + "```bash" + `
# Start scheduler
curl -X POST http://localhost:8080/scheduler/start

# Stop scheduler  
curl -X POST http://localhost:8080/scheduler/stop

# Check status
curl http://localhost:8080/scheduler/status
` + "```" + `

## Monitoring Integration

The metrics endpoint provides data suitable for monitoring systems like Prometheus, Grafana, or custom dashboards.

Key metrics to monitor:
- ` + "`success_rate`" + ` - Overall sync success rate
- ` + "`last_sync_time`" + ` - When the last sync occurred
- ` + "`failed_syncs`" + ` - Number of failed synchronizations
- ` + "`average_sync_duration`" + ` - Performance trending

## Rate Limiting

The API does not implement rate limiting by default. Consider adding a reverse proxy (nginx, Apache) for production deployments if rate limiting is needed.
`

	return os.WriteFile(path, []byte(content), 0644)
}

func generateTroubleshootingGuide(path string) error {
	content := `# Go SCIM Sync Troubleshooting Guide

## Common Issues and Solutions

### Configuration Issues

#### "Configuration validation failed"
**Symptoms:** Validation errors when running ` + "`setup validate`" + ` or starting the application.

**Solutions:**
1. Run the setup wizard again: ` + "`./scim-sync setup wizard`" + `
2. Check required fields in ` + "`config.yaml`" + `
3. Ensure all file paths are correct and accessible

#### "Service account file not found"
**Symptoms:** Error about missing service account JSON file.

**Solutions:**
1. Verify the file path in your configuration
2. Check file permissions (should be readable)
3. Use absolute paths if relative paths cause issues
4. Re-download the service account key from Google Cloud Console

### Authentication Issues

#### "Beyond Identity API token not set in config.yaml"
**Symptoms:** Error when trying to connect to Beyond Identity API.

**Solutions:**
1. Set the API token in your config.yaml file under beyond_identity.api_token
2. Run the setup wizard again: ` + "`./scim-sync setup wizard`" + `
3. Verify the token is valid and has SCIM permissions

#### "Authentication failed" with Beyond Identity
**Symptoms:** 401 Unauthorized errors when accessing Beyond Identity API.

**Solutions:**
1. Verify your API token is correct
2. Check token permissions in Beyond Identity Admin Console
3. Ensure token hasn't expired
4. Try generating a new API token

#### "Domain-wide delegation" errors with Google Workspace
**Symptoms:** OAuth errors when accessing Google Workspace APIs.

**Solutions:**
1. Verify domain-wide delegation is enabled for your service account
2. Check OAuth scopes in Google Admin Console:
   - ` + "`https://www.googleapis.com/auth/admin.directory.user`" + `
   - ` + "`https://www.googleapis.com/auth/admin.directory.group`" + `
   - ` + "`https://www.googleapis.com/auth/admin.directory.group.member`" + `
3. Ensure the Client ID matches your service account
4. Wait a few minutes for changes to propagate

### Sync Issues

#### "Group not found" errors
**Symptoms:** 404 errors when trying to sync specific groups.

**Solutions:**
1. Verify group email addresses are correct
2. Check that groups exist in Google Workspace
3. Ensure the service account has access to the groups
4. Remove non-existent groups from configuration

#### "The authorization token is missing required scopes"
**Symptoms:** 403 errors from Beyond Identity API.

**Solutions:**
1. Regenerate API token with proper SCIM permissions
2. Check Beyond Identity Admin Console for required scopes
3. Contact Beyond Identity support if scope issues persist

#### Sync takes too long or times out
**Symptoms:** Sync operations hang or timeout.

**Solutions:**
1. Reduce the number of groups in configuration
2. Increase retry delay: ` + "`retry_delay_seconds: 60`" + `
3. Check network connectivity to both APIs
4. Monitor API rate limits and adjust sync frequency

### Server Mode Issues

#### "Port already in use"
**Symptoms:** Cannot start server mode due to port conflicts.

**Solutions:**
1. Change port in configuration: ` + "`server.port: 8081`" + `
2. Kill processes using the port: ` + "`lsof -ti:8080 | xargs kill`" + `
3. Use a different port number

#### Scheduler not running
**Symptoms:** Automatic syncs not occurring as scheduled.

**Solutions:**
1. Verify ` + "`schedule_enabled: true`" + ` in configuration
2. Check cron schedule syntax
3. Look for scheduler errors in logs
4. Restart the server

### Performance Issues

#### High memory usage
**Symptoms:** Application uses excessive memory.

**Solutions:**
1. Reduce the number of groups being synced
2. Increase ` + "`retry_delay_seconds`" + ` to reduce API pressure
3. Monitor for memory leaks and restart periodically

#### Slow sync performance
**Symptoms:** Syncs take much longer than expected.

**Solutions:**
1. Check network latency to APIs
2. Reduce log level to ` + "`warn`" + ` or ` + "`error`" + `
3. Monitor API rate limits
4. Consider syncing fewer groups per operation

### Logging and Debugging

#### Enable Debug Logging
Add to your configuration:
` + "```yaml" + `
app:
  log_level: "debug"
` + "```" + `

#### Trace API Calls
For detailed API debugging, you can set environment variables:
` + "```bash" + `
export GODEBUG=http2debug=1
` + "```" + `

#### Log File Analysis
Look for these patterns in logs:
- ` + "`ERROR`" + ` - Critical issues requiring immediate attention
- ` + "`WARNING`" + ` - Issues that may affect sync quality
- ` + "`Failed to`" + ` - Operation failures
- ` + "`401`" + ` or ` + "`403`" + ` - Authentication/authorization issues

### Environment-Specific Issues

#### Docker/Container Issues
**Symptoms:** Application works locally but fails in containers.

**Solutions:**
1. Ensure environment variables are passed to container
2. Mount configuration files and service account keys properly
3. Check container networking for API access
4. Verify file permissions in container

#### Network/Firewall Issues
**Symptoms:** Cannot connect to Google or Beyond Identity APIs.

**Solutions:**
1. Check firewall rules for outbound HTTPS (443)
2. Verify DNS resolution for API endpoints
3. Test connectivity: ` + "`curl https://api.byndid.com`" + `
4. Configure proxy settings if required

### Getting Help

#### Validation Command
Always start troubleshooting with:
` + "```bash" + `
./scim-sync setup validate
` + "```" + `

#### Collect Debug Information
1. Run with debug logging enabled
2. Check configuration: ` + "`./scim-sync validate-config`" + `
3. Test individual components with setup validation
4. Capture relevant log snippets

#### Common Log Patterns to Share
- Complete error messages with stack traces
- API response codes and messages
- Configuration validation output
- Network connectivity test results

#### When to Contact Support
- API tokens and service accounts are correctly configured
- Configuration passes validation
- Network connectivity is confirmed
- Issue persists across multiple attempts

#### Information to Include
1. Go SCIM sync version: ` + "`./scim-sync version`" + `
2. Configuration file (with secrets redacted)
3. Complete error messages
4. Steps to reproduce the issue
5. Environment details (OS, container, etc.)
`

	return os.WriteFile(path, []byte(content), 0644)
}