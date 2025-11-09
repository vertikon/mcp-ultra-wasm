# NATS Subjects Documentation

## Overview

This document describes all NATS subjects used in the mcp-ultra-wasm application.

## Subject Naming Convention

Subjects follow the pattern: `<domain>.<entity>.<action>`

## Subjects

### Task Domain

| Subject | Description | Publisher | Subscriber | Payload |
|---------|-------------|-----------|------------|---------|
| `task.created` | Task creation event | Task Service | Event Handlers | TaskCreatedEvent |
| `task.updated` | Task update event | Task Service | Event Handlers | TaskUpdatedEvent |
| `task.deleted` | Task deletion event | Task Service | Event Handlers | TaskDeletedEvent |
| `task.completed` | Task completion event | Task Service | Notification Service | TaskCompletedEvent |

### System Domain

| Subject | Description | Publisher | Subscriber | Payload |
|---------|-------------|-----------|------------|---------|
| `system.health.check` | Health check request | Health Monitor | All Services | HealthCheckRequest |
| `system.health.response` | Health check response | All Services | Health Monitor | HealthCheckResponse |

## Event Payloads

### TaskCreatedEvent

```json
{
  "task_id": "uuid",
  "title": "string",
  "created_by": "uuid",
  "created_at": "timestamp",
  "priority": "string",
  "tags": ["string"]
}
```

### TaskUpdatedEvent

```json
{
  "task_id": "uuid",
  "updated_fields": ["string"],
  "updated_by": "uuid",
  "updated_at": "timestamp"
}
```

### TaskDeletedEvent

```json
{
  "task_id": "uuid",
  "deleted_by": "uuid",
  "deleted_at": "timestamp"
}
```

### TaskCompletedEvent

```json
{
  "task_id": "uuid",
  "completed_by": "uuid",
  "completed_at": "timestamp",
  "duration_seconds": number
}
```

## Subject Subscriptions

### Event Handlers

- `task.*` - All task events
- `system.health.*` - Health check events

### Notification Service

- `task.completed` - Send completion notifications
- `task.created` - Send creation notifications (if enabled)

## Notes

- All subjects use wildcard subscriptions where appropriate
- Events are published asynchronously
- Subscribers should implement idempotency
- Failed event processing is retried with exponential backoff

## Future Subjects (Planned)

- `user.created` - User registration events
- `notification.send` - Notification dispatch events
- `analytics.track` - Analytics tracking events
