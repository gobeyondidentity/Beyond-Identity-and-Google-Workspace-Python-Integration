# Go SCIM Sync

A Go application for synchronizing users and groups from Google Workspace to Beyond Identity using SCIM protocol.

## Phase 1 - MVP Implementation Status

✅ **Completed:**
- Basic CLI structure with Cobra
- YAML configuration loading with environment variable substitution
- Configuration validation with helpful error messages
- Python-compatible logging format
- Commands: `run`, `validate-config`, `version`

🚧 **In Progress:**
- Phase 2: Core sync functionality (Google API + Beyond Identity SCIM)

## Quick Start

### 1. Build the application
```bash
go mod tidy
go build -o go-scim-sync ./cmd
```

### 2. Create configuration
```bash
cp configs/config.example.yaml config.yaml
# Edit config.yaml with your actual values
```

### 3. Set environment variables
```bash
export BI_API_TOKEN="your-beyond-identity-api-token"
```

### 4. Validate configuration
```bash
./go-scim-sync validate-config
```

### 5. Run synchronization
```bash
./go-scim-sync run
```

## Commands

### `go-scim-sync run`
Run SCIM synchronization once and exit.

**Options:**
- `--config PATH` - Specify configuration file path (default: ./config.yaml)

### `go-scim-sync validate-config`
Validate configuration file syntax and required fields.

### `go-scim-sync version`
Print version information.

## Configuration

The application uses a YAML configuration file. See `configs/config.example.yaml` for a complete example.

### Required Settings

```yaml
google_workspace:
  domain: "your-domain.com"
  super_admin_email: "admin@your-domain.com"
  service_account_key_path: "./service-account.json"

beyond_identity:
  api_token: "${BI_API_TOKEN}"

sync:
  groups:
    - "group1@your-domain.com"
```

### Environment Variables

- `BI_API_TOKEN` - Beyond Identity API token (required)

### Configuration File Locations

The application searches for configuration files in this order:
1. `./config.yaml`
2. `./config.yml`
3. `~/.config/go-scim-sync/config.yaml`
4. `~/.config/go-scim-sync/config.yml`

## Development Status

This is Phase 1 of the migration from Python to Go. Current implementation includes:

- ✅ CLI framework
- ✅ Configuration management
- ✅ Logging setup
- ⏳ Google Workspace API client (Phase 2)
- ⏳ Beyond Identity SCIM client (Phase 2)
- ⏳ Sync engine (Phase 2)
- ⏳ Server mode (Phase 3)
- ⏳ Interactive configuration wizard (Phase 4)

## Architecture

```
go-scim-sync/
├── cmd/                    # CLI commands
├── internal/
│   ├── config/            # Configuration management
│   └── logger/            # Logging setup
└── configs/               # Example configurations
```

## Contributing

This project follows the Go standard project layout and uses:
- `github.com/spf13/cobra` for CLI
- `gopkg.in/yaml.v3` for configuration
- `github.com/sirupsen/logrus` for logging

## License

[Add license information]