# Deprecated Python Implementation

⚠️ **DEPRECATED** ⚠️

This directory contains the deprecated Python implementation of the Google Workspace to Beyond Identity SCIM sync tool.

## Migration Notice

**The Python version has been replaced with a Go implementation that is now the primary version.**

### Python Version (Deprecated)
- **Location**: This `deprecated/` directory
- **Status**: No longer maintained
- **Files**: 
  - `gwbisync.py` - Main Python script
  - `config.example.py` - Python configuration example
  - `requirements.txt` - Python dependencies
  - `setup.sh` - Python setup script
  - `python-docs/` - Original documentation

### Go Version (Current)
- **Location**: Repository root
- **Status**: Actively maintained and enhanced
- **Binary**: `./scim-sync`
- **Documentation**: `./docs/`

## Migration Benefits

The Go implementation provides:

✅ **Single binary deployment** - No Python dependencies  
✅ **Better performance** - Faster startup and execution  
✅ **Enhanced features** - Interactive setup wizard, server mode, scheduling  
✅ **Better UX** - Comprehensive validation and error reporting  
✅ **Production ready** - Built-in metrics, health checks, and monitoring  

## For Existing Python Users

### Quick Migration Steps

1. **Install Go version**:
   ```bash
   # Build from source
   go build -o scim-sync ./cmd
   
   # Or use pre-built binary
   ./scim-sync
   ```

2. **Migrate configuration**:
   ```bash
   # Use interactive wizard
   ./scim-sync setup wizard
   
   # Or manually convert your Python config
   ```

3. **Test the migration**:
   ```bash
   # Validate setup
   ./scim-sync setup validate
   
   # Run test sync
   ./scim-sync run
   ```

4. **Production deployment**:
   ```bash
   # One-time sync
   ./scim-sync run
   
   # Or server mode with scheduling
   ./scim-sync server
   ```

### Configuration Mapping

| Python Setting | Go Equivalent |
|----------------|---------------|
| `GOOGLE_DOMAIN` | `google_workspace.domain` |
| `SUPER_ADMIN_EMAIL` | `google_workspace.super_admin_email` |
| `SERVICE_ACCOUNT_FILE` | `google_workspace.service_account_key_path` |
| `BI_API_TOKEN` | `beyond_identity.api_token` |
| `GROUPS_TO_SYNC` | `sync.groups` |

### Environment Variables

Both versions support the same environment variables:
- `BI_API_TOKEN` - Beyond Identity API token

## Support

- **Go version support**: Create issues in the main repository
- **Python version**: No longer supported - please migrate to Go version

## Historical Reference

The Python implementation served as the foundation for the Go version and successfully validated the SCIM synchronization approach. All Python functionality has been ported to Go with additional enhancements.

---

**Last Python version**: Archived on migration to Go (May 2025)  
**Recommended action**: Migrate to Go implementation in repository root