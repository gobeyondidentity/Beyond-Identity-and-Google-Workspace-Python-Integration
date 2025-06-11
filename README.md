# Google Workspace ↔ Beyond Identity Bi-directional Sync

A high-performance Go application for bi-directional synchronization between Google Workspace and Beyond Identity using SCIM protocol.

> 🆕 **Go Implementation Now Primary** - The Python version has been moved to `deprecated/` folder. This Go implementation provides better performance, enhanced features, and production-ready capabilities.

## ✨ Key Features

✅ **Bi-directional Sync** - GWS → BI provisioning + BI → GWS enrollment status management  
✅ **Complete SCIM Synchronization** - Full user and group sync with membership management  
✅ **Enrollment Group Management** - Automatic Google group updates based on BI activation status  
✅ **Interactive Setup Wizard** - Guided configuration with validation  
✅ **Server Mode** - HTTP API with automatic scheduling  
✅ **Comprehensive Validation** - Connectivity testing and error reporting  
✅ **Production Ready** - Health checks, metrics, and monitoring  
✅ **Single Binary** - No dependencies, easy deployment

## 🚀 Quick Start

### Option 1: Interactive Setup (Recommended)

```bash
# Build the application
go build -o scim-sync ./cmd

# Run interactive setup wizard
./scim-sync setup wizard

# Validate your setup
./scim-sync setup validate

# Run your first sync
./scim-sync run
```

### Option 2: Manual Configuration

```bash
# Build the application
go build -o scim-sync ./cmd

# Create configuration from example
cp configs/config.example.yaml config.yaml

# Edit config.yaml with your values, then validate
./scim-sync validate-config

# Run synchronization
./scim-sync run
```

## 📋 Commands

### Core Operations
- `./scim-sync run` - Run one-time synchronization
- `./scim-sync server` - Start server mode with scheduling and HTTP API

### Setup & Configuration  
- `./scim-sync setup wizard` - Interactive configuration wizard
- `./scim-sync setup validate` - Validate setup and test connectivity
- `./scim-sync setup docs` - Generate documentation

### Utilities
- `./scim-sync validate-config` - Validate configuration file
- `./scim-sync version` - Show version information

### Server Mode API
When running `./scim-sync server`, these endpoints are available:
- `GET /health` - Health check and status
- `POST /sync` - Trigger manual sync
- `GET /metrics` - Sync metrics and statistics
- `GET /version` - Version information

## Configuration

The application uses a YAML configuration file. See `configs/config.example.yaml` for a complete example.

### Required Settings

```yaml
google_workspace:
  domain: "your-domain.com"
  super_admin_email: "admin@your-domain.com"
  service_account_key_path: "./service-account.json"

beyond_identity:
  api_token: "your-beyond-identity-api-token"

sync:
  groups:
    - "group1@your-domain.com"
  enrollment_group_email: "byid-enrolled@your-domain.com"  # Optional: Auto-managed enrollment group
```

### API Token Configuration

The Beyond Identity API token should be configured in the `config.yaml` file under `beyond_identity.api_token`.

### Configuration File Locations

The application searches for configuration files in this order:
1. `./config.yaml`
2. `./config.yml`
3. `~/.config/scim-sync/config.yaml`
4. `~/.config/scim-sync/config.yml`

## 🔄 Bi-directional Sync

The application performs synchronization in both directions:

### GWS → BI Sync (Provisioning)
- **Users**: Creates/updates user accounts in Beyond Identity
- **Groups**: Creates groups with configured prefix (e.g., `GoogleSCIM_Engineering`)
- **Memberships**: Syncs group membership from Google Workspace to Beyond Identity
- **Lifecycle**: Handles user activation, deactivation, and updates

### BI → GWS Sync (Enrollment Status)
- **Status Monitoring**: Checks Beyond Identity user activation status via SCIM API
- **Enrollment Group**: Automatically manages a Google Workspace group for enrolled users
- **Real-time Updates**: 
  - Users who **activate** in BI → **Added** to enrollment group
  - Users who **deactivate** in BI → **Removed** from enrollment group
- **Audit Trail**: All enrollment changes are logged for compliance

### Enrollment Group Configuration

```yaml
sync:
  enrollment_group_email: "byid-enrolled@your-domain.com"  # Default: byid-enrolled@{domain}
  enrollment_group_name: "BYID Enrolled"                   # Default: "BYID Enrolled"
```

The enrollment group is automatically created if it doesn't exist. Users in the configured `sync.groups` are monitored for Beyond Identity activation status changes.

## 🎯 Implementation Status

**✅ COMPLETE** - All phases of the migration from Python to Go have been implemented:

### Phase 1-2: Core Functionality ✅
- ✅ CLI framework with Cobra
- ✅ Configuration management with YAML and env vars
- ✅ Google Workspace API client with service account auth
- ✅ Beyond Identity SCIM API client with full CRUD operations
- ✅ Complete sync engine ported from Python
- ✅ Comprehensive error handling and retry logic

### Phase 3: Server Mode ✅
- ✅ HTTP API server with health checks
- ✅ Automatic sync scheduling with cron expressions
- ✅ Metrics collection and exposure
- ✅ Manual sync triggers via API

### Phase 4: Setup Management ✅
- ✅ Interactive configuration wizard
- ✅ Setup validation with connectivity testing
- ✅ Automatic documentation generation
- ✅ Enhanced error reporting and guidance

## 🏗️ Architecture

```
scim-sync/
├── cmd/                    # CLI entry point and commands
├── internal/
│   ├── config/            # Configuration management and validation
│   ├── gws/               # Google Workspace API client
│   ├── bi/                # Beyond Identity SCIM API client  
│   ├── sync/              # Synchronization engine
│   ├── server/            # HTTP server and scheduling
│   ├── wizard/            # Interactive setup wizard
│   ├── setup/             # Setup validation and docs generation
│   └── logger/            # Structured logging
├── configs/               # Example configurations
├── docs/                  # Generated documentation
└── deprecated/            # Legacy Python implementation
```

## 📚 Documentation

Complete documentation is available in the `docs/` directory:

- **[Setup Guide](docs/SETUP.md)** - Comprehensive setup instructions with prerequisites
- **[API Reference](docs/API.md)** - Complete HTTP API documentation for server mode  
- **[Troubleshooting](docs/TROUBLESHOOTING.md)** - Common issues and solutions

Generate fresh documentation anytime with:
```bash
./scim-sync setup docs
```

## 🔄 Migration from Python

The Python implementation has been moved to `deprecated/` folder. See `deprecated/README.md` for migration instructions.

**Migration benefits:**
- ⚡ **10x faster** startup time
- 📦 **Single binary** deployment (no Python dependencies)
- 🛠️ **Enhanced features** (wizard, server mode, validation)
- 📊 **Built-in monitoring** (health checks, metrics)
- 🚀 **Production ready** (scheduling, error handling)

## 🤝 Contributing

This project follows Go standard practices and uses:
- `github.com/spf13/cobra` - CLI framework
- `google.golang.org/api` - Google Workspace APIs  
- `github.com/robfig/cron/v3` - Scheduling
- `github.com/sirupsen/logrus` - Structured logging

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

Copyright (c) 2024 Beyond Identity