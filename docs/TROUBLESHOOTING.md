# Go SCIM Sync Troubleshooting Guide

## Common Issues and Solutions

### Configuration Issues

#### "Configuration validation failed"
**Symptoms:** Validation errors when running `setup validate` or starting the application.

**Solutions:**
1. Run the setup wizard again: `./scim-sync setup wizard`
2. Check required fields in `config.yaml`
3. Ensure all file paths are correct and accessible

#### "Service account file not found"
**Symptoms:** Error about missing service account JSON file.

**Solutions:**
1. Verify the file path in your configuration
2. Check file permissions (should be readable)
3. Use absolute paths if relative paths cause issues
4. Re-download the service account key from Google Cloud Console

### Authentication Issues

#### "API token not configured"
**Symptoms:** Error when trying to connect to Beyond Identity API.

**Solutions:**
1. Configure the API token in your `config.yaml` file under `beyond_identity.api_token`
2. Verify the token is valid and has SCIM permissions
3. Run setup validation to test connectivity

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
   - `https://www.googleapis.com/auth/admin.directory.user`
   - `https://www.googleapis.com/auth/admin.directory.group`
   - `https://www.googleapis.com/auth/admin.directory.group.member`
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
2. Increase retry delay: `retry_delay_seconds: 60`
3. Check network connectivity to both APIs
4. Monitor API rate limits and adjust sync frequency

### Server Mode Issues

#### "Port already in use"
**Symptoms:** Cannot start server mode due to port conflicts.

**Solutions:**
1. Change port in configuration: `server.port: 8081`
2. Kill processes using the port: `lsof -ti:8080 | xargs kill`
3. Use a different port number

#### Scheduler not running
**Symptoms:** Automatic syncs not occurring as scheduled.

**Solutions:**
1. Verify `schedule_enabled: true` in configuration
2. Check cron schedule syntax
3. Look for scheduler errors in logs
4. Restart the server

### Performance Issues

#### High memory usage
**Symptoms:** Application uses excessive memory.

**Solutions:**
1. Reduce the number of groups being synced
2. Increase `retry_delay_seconds` to reduce API pressure
3. Monitor for memory leaks and restart periodically

#### Slow sync performance
**Symptoms:** Syncs take much longer than expected.

**Solutions:**
1. Check network latency to APIs
2. Reduce log level to `warn` or `error`
3. Monitor API rate limits
4. Consider syncing fewer groups per operation

### Logging and Debugging

#### Enable Debug Logging
Add to your configuration:
```yaml
app:
  log_level: "debug"
```

#### Trace API Calls
For detailed API debugging, you can set environment variables:
```bash
export GODEBUG=http2debug=1
```

#### Log File Analysis
Look for these patterns in logs:
- `ERROR` - Critical issues requiring immediate attention
- `WARNING` - Issues that may affect sync quality
- `Failed to` - Operation failures
- `401` or `403` - Authentication/authorization issues

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
3. Test connectivity: `curl https://api.byndid.com`
4. Configure proxy settings if required

### Getting Help

#### Validation Command
Always start troubleshooting with:
```bash
./scim-sync setup validate
```

#### Collect Debug Information
1. Run with debug logging enabled
2. Check configuration: `./scim-sync validate-config`
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
1. Go SCIM sync version: `./scim-sync version`
2. Configuration file (with secrets redacted)
3. Complete error messages
4. Steps to reproduce the issue
5. Environment details (OS, container, etc.)
