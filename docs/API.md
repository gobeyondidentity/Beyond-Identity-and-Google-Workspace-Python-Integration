# Go SCIM Sync API Reference

When running in server mode (`./scim-sync server`), the application provides an HTTP API for management and monitoring.

## Base URL

By default, the server runs on port 8080:
```
http://localhost:8080
```

## Endpoints

### Health Check
```http
GET /health
```

Returns server health status and next scheduled sync time.

**Response Example:**
```json
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
```

### Manual Sync
```http
POST /sync
```

Triggers a manual synchronization operation.

**Response Example:**
```json
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
```

### Metrics
```http
GET /metrics
```

Returns synchronization metrics and statistics.

**Response Example:**
```json
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
```

### Version Information
```http
GET /version
```

Returns application version information.

**Response Example:**
```json
{
  "version": "0.1.0",
  "build_time": "2024-01-15T08:00:00Z",
  "mode": "server"
}
```

### Scheduler Control

#### Start Scheduler
```http
POST /scheduler/start
```

Starts the automatic sync scheduler (if configured).

#### Stop Scheduler
```http
POST /scheduler/stop
```

Stops the automatic sync scheduler.

#### Scheduler Status
```http
GET /scheduler/status
```

Returns scheduler status and configuration.

**Response Example:**
```json
{
  "running": true,
  "schedule": "0 */6 * * *",
  "last_sync": "2024-01-15T10:00:00Z",
  "next_sync": "2024-01-15T16:00:00Z"
}
```

## Error Responses

All endpoints return appropriate HTTP status codes:

- `200` - Success
- `400` - Bad Request
- `500` - Internal Server Error

Error response format:
```json
{
  "error": "Error description",
  "details": "Additional error details if available"
}
```

## cURL Examples

### Check Health
```bash
curl http://localhost:8080/health
```

### Trigger Manual Sync
```bash
curl -X POST http://localhost:8080/sync
```

### Get Metrics
```bash
curl http://localhost:8080/metrics
```

### Control Scheduler
```bash
# Start scheduler
curl -X POST http://localhost:8080/scheduler/start

# Stop scheduler  
curl -X POST http://localhost:8080/scheduler/stop

# Check status
curl http://localhost:8080/scheduler/status
```

## Monitoring Integration

The metrics endpoint provides data suitable for monitoring systems like Prometheus, Grafana, or custom dashboards.

Key metrics to monitor:
- `success_rate` - Overall sync success rate
- `last_sync_time` - When the last sync occurred
- `failed_syncs` - Number of failed synchronizations
- `average_sync_duration` - Performance trending

## Rate Limiting

The API does not implement rate limiting by default. Consider adding a reverse proxy (nginx, Apache) for production deployments if rate limiting is needed.
