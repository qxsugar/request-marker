# Request Marker - Traefik Plugin

A Traefik middleware plugin that marks HTTP requests with headers based on configurable rules. Supports version ranges,
user identification, probabilistic canary routing, and path matching for traffic management, canary deployments, and A/B
testing.

## Features

- **Multiple rule types**: Version ranges, user identification, canary (probabilistic), path matching
- **Dynamic rule loading**: Periodically load and update rules from Redis without restart
- **Priority-based evaluation**: Rules evaluated in priority order (highest first), first match wins
- **Flexible user identification**: Extract user ID from HTTP header, cookie, or query parameter
- **Simple logger**: Compatible with Traefik's restricted environment, supports DEBUG/INFO/ERROR levels

## Installation

### 1. Configure Traefik Plugin

Add plugin declaration to Traefik config:

```yaml
experimental:
  plugins:
    request-marker:
      moduleName: github.com/qxsugar/request-marker
      version: v0.0.2
```

### 2. Configure Middleware

```yaml
middlewares:
  request-marker:
    plugin:
      request-marker:
        tag: api                        # Tag to match rules
        logLevel: DEBUG                 # Log level: DEBUG, INFO, ERROR
        markerKey: X-MARK               # Header key for marking
        versionHeader: X-Version        # Header for version info
        identifyHeader: X-User-ID       # Header for user ID
        identifyCookie: user_id         # Cookie name for user ID
        identifyQuery: uid              # Query param for user ID

        # Redis dynamic rules (optional)
        redisConfig:
          enable: true
          addr: redis.example.com:6379
          password: "your-password"
          db: 0
          ruleListKeys: marker:api:rules
          refreshInterval: 15

        # Static rules (used when Redis disabled)
        staticRules:
          - tag: api
            name: beta-users
            enable: true
            priority: 100
            type: identify
            markValue: beta
            userIds:
              - user001
              - user002

          - tag: api
            name: version-range
            enable: true
            priority: 90
            type: version
            markValue: v2
            minVersion: 2.0.0
            maxVersion: 2.9.9

          - tag: api
            name: canary-30
            enable: true
            priority: 80
            type: canary
            markValue: canary
            canary: 30

          - tag: api
            name: admin-path
            enable: true
            priority: 70
            type: path
            markValue: admin
            path: /admin
```

### 3. Use Middleware in Router

```yaml
routers:
  api-router:
    rule: host(`api.example.com`)
    service: api-service
    entryPoints:
      - web
    middlewares:
      - request-marker

services:
  api-service:
    loadBalancer:
      servers:
        - url: "http://backend-1:8080"
        - url: "http://backend-2:8080"
```

## Rule Types

### Version Rule

Match requests by semantic version range.

```yaml
type: version
minVersion: 2.0.0
maxVersion: 2.9.9
markValue: v2-stable
```

### Identify Rule

Match specific user IDs. User ID extracted from header → cookie → query parameter.

```yaml
type: identify
userIds:
  - user001
  - user002
markValue: beta-tester
```

### Canary Rule

Probabilistic matching based on user ID hash. Canary value is 0-100 (percentage).

```yaml
type: canary
canary: 30
markValue: canary
```

### Path Rule

Match request URL by substring.

```yaml
type: path
path: /admin
markValue: admin-panel
```

## Redis Rule Storage

Rules stored as Redis hashes with hierarchical key structure.

### Key Structure

```
{ruleListKeys}                    # List of rule keys
marker:api:rule:{ruleName}        # Individual rule hash
```

Example (ruleListKeys = `marker:api:rules`):

```
marker:api:rules                  → List ["marker:api:rule:beta-users", ...]
marker:api:rule:beta-users        → Hash {name, enable, priority, ...}
```

### Hash Fields (snake_case)

| Field         | Type   | Description                              |
|---------------|--------|------------------------------------------|
| `name`        | string | Rule name                                |
| `enable`      | 0/1    | Enable flag                              |
| `priority`    | int    | Priority (higher = first)                |
| `type`        | string | Rule type: version/identify/canary/path  |
| `mark_value`  | string | Mark value when matched                  |
| `min_version` | string | Min version (version type)               |
| `max_version` | string | Max version (version type)               |
| `user_ids`    | string | Comma-separated user IDs (identify type) |
| `canary`      | int    | Canary percentage 0-100 (canary type)    |
| `path`        | string | Path substring (path type)               |

## Development

### Build & Test

```bash
make lint      # Run linter
make test      # Run tests with coverage
make vendor    # Vendor dependencies
make clean     # Clean vendor
```

### Run Single Test

```bash
go test -v -run TestName ./...
```

## How It Works

1. **Startup**: Load static rules, sort by priority
2. **Redis refresh**: If enabled, periodically load rules from Redis
3. **Request handling**: For each request, evaluate rules in priority order
4. **Rule matching**: Check enable flag, tag match, then type-specific logic
5. **Marking**: First matching rule sets mark header
6. **Forward**: Request forwarded to backend with mark header

## User Identification Priority

User ID extracted in this order:

1. HTTP header (`identifyHeader`)
2. Cookie (`identifyCookie`)
3. Query parameter (`identifyQuery`)

If no user ID found, canary rules cannot match.

## License

MIT