# Go SCIM Sync Application - Revised Migration Plan

## Overview
Convert the Python SCIM sync application to a Go application using an **MVP-first approach** focusing on functional parity before adding enhancements.

## Revised Strategy: MVP → Enhanced → Production

### MVP Goals (Phases 1-2)
1. **Functional parity** with Python version
2. **Single binary** deployment
3. **Basic YAML configuration**
4. **One-shot mode only**
5. **Same logging output format**

### Enhanced Goals (Phases 3-4)
1. **Server mode** with scheduling
2. **Interactive configuration wizard**
3. **Advanced configuration management**
4. **Health checks and monitoring**

### Production Goals (Phase 5)
1. **Security hardening**
2. **Performance optimization**
3. **Comprehensive observability**

## MVP Architecture

### Simplified Command Structure
```bash
# MVP Commands (Phase 1-2)
go-scim-sync run                           # One-shot sync
go-scim-sync run --config /path/to/config.yaml
go-scim-sync validate-config               # Validate configuration
go-scim-sync version                       # Version information

# Enhanced Commands (Phase 3+)
go-scim-sync serve                         # Server mode
go-scim-sync config init                   # Create config file
go-scim-sync config setup                  # Interactive wizard
```

### Simplified Project Structure
```
go-scim-sync/
├── cmd/
│   ├── main.go                    # CLI entry point
│   ├── run.go                     # Run command (MVP)
│   └── validate.go                # Config validation
├── internal/
│   ├── config/
│   │   ├── config.go             # Configuration struct and loading
│   │   └── validation.go         # Basic validation
│   ├── google/
│   │   ├── client.go             # Google Workspace API client
│   │   └── service.go            # Google API operations
│   ├── beyondidentity/
│   │   ├── client.go             # BI SCIM API client
│   │   └── scim.go               # SCIM operations
│   ├── sync/
│   │   ├── engine.go             # Main sync logic (port from Python)
│   │   └── types.go              # Data structures
│   └── logger/
│       └── logger.go             # Structured logging
├── configs/
│   └── config.example.yaml       # Single example configuration
├── go.mod
├── go.sum
└── README.md
```

## MVP Configuration Design

### Single YAML Configuration File
```yaml
# Minimal but complete configuration
app:
  log_level: "info"          # debug, info, warn, error
  test_mode: true            # Dry-run mode

# Google Workspace configuration
google_workspace:
  domain: "byndid-mail.com"
  super_admin_email: "nmelo@byndid-mail.com"
  service_account_key_path: "./service-account.json"

# Beyond Identity configuration
beyond_identity:
  api_token: "your-beyond-identity-api-token"  # Configure directly in file
  scim_base_url: "https://api.byndid.com/scim/v2"
  native_api_url: "https://api.byndid.com/v2"
  group_prefix: "GoogleSCIM_"

# Sync configuration
sync:
  groups:
    - "scim_test@byndid-mail.com"
  retry_attempts: 3
  retry_delay_seconds: 30
```

## Revised Implementation Phases

### Phase 1: MVP Foundation (Week 1)
**Goal**: Basic working Go application

1. **Project Setup**
   ```bash
   go mod init github.com/gobeyondidentity/go-scim-sync
   ```

2. **Minimal Dependencies**
   ```go
   github.com/spf13/cobra      // CLI framework
   gopkg.in/yaml.v3            // YAML parsing
   google.golang.org/api       // Google API client
   github.com/sirupsen/logrus  // Structured logging
   ```

3. **Core Structure**
   - Basic CLI with `run` and `validate-config` commands
   - YAML configuration loading with env var substitution
   - Simple validation (required fields only)

4. **Logging Setup**
   - Match Python output format exactly
   - Configurable log levels
   - Console output only (no file logging yet)

### Phase 2: Core Functionality (Week 2-3)
**Goal**: Functional parity with Python version

1. **Google Workspace Client**
   - Service account authentication
   - Group and user operations
   - Direct port of Python logic

2. **Beyond Identity SCIM Client**
   - HTTP client with proper error handling
   - User CRUD operations with required fields (displayName, emails)
   - Group CRUD with PATCH operations for membership

3. **Sync Engine**
   - Direct port of Python sync logic
   - Same error handling and retry mechanisms
   - Identical output messages and behavior

4. **Testing**
   - Unit tests for core functions
   - Integration test with test data
   - Validate against current Python behavior

### Phase 3: Enhanced Features (Week 4)
**Goal**: Add value beyond Python version

1. **Server Mode**
   ```go
   // Minimal server mode
   github.com/robfig/cron/v3   // Scheduling
   net/http                    // Health checks
   ```

2. **Enhanced Configuration**
   ```yaml
   server:
     enabled: false
     port: 8080
     sync_interval: "5m"
     health_check_path: "/health"
   ```

3. **Basic Health Checks**
   - `/health` endpoint
   - `/ready` endpoint
   - Basic metrics

### Phase 4: Configuration Management (Week 5)
**Goal**: Improve user experience

1. **Interactive Wizard**
   ```go
   github.com/AlecAivazis/survey/v2  // Interactive prompts
   ```

2. **Configuration Commands**
   - `config init` - Create basic config
   - `config setup` - Interactive wizard
   - Enhanced validation with API testing

### Phase 5: Production Readiness (Week 6+)
**Goal**: Enterprise-ready deployment

1. **Security & Operations**
   - Secure credential handling
   - Rate limiting and circuit breakers
   - Graceful shutdown
   - Signal handling

2. **Observability**
   - Structured metrics
   - Distributed tracing
   - Enhanced logging

3. **Performance**
   - Concurrent processing
   - Memory optimization
   - Connection pooling

## MVP Dependencies (Minimal Set)

```go
// Phase 1-2 Dependencies (MVP)
github.com/spf13/cobra          // CLI framework
gopkg.in/yaml.v3                // YAML parsing  
google.golang.org/api           // Google API client
github.com/sirupsen/logrus      // Logging
context                         // Standard library
net/http                        // Standard library
os                              // Standard library
```

## Critical Design Decisions

### 1. Error Handling Strategy
```go
// Custom error types for different failure modes
type SyncError struct {
    Type    ErrorType
    Message string
    Cause   error
}

type ErrorType int
const (
    GoogleAPIError ErrorType = iota
    BeyondIdentityAPIError
    ConfigurationError
    ValidationError
)
```

### 2. State Management (MVP)
```go
// Simple in-memory state tracking
type SyncState struct {
    ProcessedUsers  map[string]UserStatus
    ProcessedGroups map[string]GroupStatus
    Errors          []SyncError
}
```

### 3. Logging Compatibility
```go
// Match Python format exactly
// 2025-05-30 12:21:53,426 - INFO - Starting sync process
logger.WithFields(logrus.Fields{
    "timestamp": time.Now().Format("2006-01-02 15:04:05,000"),
    "level":     "INFO",
}).Info("Starting sync process")
```

## Success Criteria by Phase

### Phase 1 Success Criteria
- [x] Go application compiles and runs
- [x] Loads YAML configuration correctly
- [x] Validates configuration with helpful errors
- [x] CLI commands work as expected

### Phase 2 Success Criteria
- [x] Successfully syncs users from Google → Beyond Identity
- [x] Creates groups and manages memberships
- [x] Handles errors gracefully with same behavior as Python
- [x] Output format matches Python version exactly
- [x] Test mode works correctly

### Phase 3+ Success Criteria
- [x] Server mode runs continuously
- [x] Health checks respond correctly
- [x] Configuration wizard creates valid configs
- [x] Performance meets or exceeds Python version

## Risk Mitigation

### Technical Risks
1. **Google API Differences**: Test thoroughly against Python behavior
2. **SCIM Compatibility**: Use exact same API calls and data structures
3. **Performance**: Profile memory usage with large datasets

### Operational Risks
1. **Breaking Changes**: Maintain exact CLI compatibility
2. **Configuration**: Support importing Python config format
3. **Deployment**: Provide migration guide and rollback plan

## Testing Strategy

### Phase 1-2 Testing
```bash
# Functional parity tests
go test ./... -v                    # Unit tests
./test-integration.sh               # Compare with Python output
./test-production.sh                # Test with real APIs
```

### Phase 3+ Testing
```bash
go test ./... -race                 # Race condition testing
go test ./... -bench=.              # Performance benchmarks
./test-server-mode.sh               # Server mode testing
```

## Delivery Timeline

| Phase | Duration | Deliverable |
|-------|----------|-------------|
| **Phase 1** | 1 week | Basic CLI + Config |
| **Phase 2** | 2 weeks | Full sync functionality |
| **Phase 3** | 1 week | Server mode |
| **Phase 4** | 1 week | Config wizard |
| **Phase 5** | 2+ weeks | Production features |

**Total MVP (Phase 1-2): 3 weeks**
**Total Enhanced (Phase 1-4): 5 weeks**

## Migration Path

### For Current Python Users
1. **Install Go binary** alongside Python version
2. **Convert config** using provided tool
3. **Test in parallel** with Python version
4. **Gradual rollout** starting with test environments
5. **Full migration** after validation period

### Backward Compatibility
- Support reading Python-style config files
- Provide conversion utility
- Maintain same environment variable names
- Same command-line argument structure

## Final Assessment

**This revised plan is MUCH MORE REASONABLE:**

✅ **Focused MVP** delivers value quickly
✅ **Incremental approach** reduces risk  
✅ **Clear success criteria** at each phase
✅ **Minimal dependencies** for MVP
✅ **Realistic timeline** based on scope
✅ **Migration strategy** for existing users

**Recommended next step**: Start with Phase 1 implementation focusing on basic CLI structure and configuration loading.