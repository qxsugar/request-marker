# Request Marker Example

This directory contains example configurations and tools for the request-marker plugin.

## Files

- `config.yaml` - Example configuration file with static rules

## Loading Rules into Redis

Use the `load-config` tool to load rules from a YAML file into Redis:

```bash
./load-config -config example/config.yaml -redis localhost:6379
```

### Options

- `-config` - Path to YAML config file (default: `config.yaml`)
- `-redis` - Redis address (default: `localhost:6379`)
- `-password` - Redis password (optional)
- `-db` - Redis database number (default: `0`)

### Example Usage

Load config with authentication:
```bash
./load-config -config example/config.yaml -redis redis.example.com:6379 -password secret -db 1
```

### Output

```
✓ Loaded rule: beta-users
✓ Loaded rule: version-v2
✓ Loaded rule: canary-30
✓ Loaded rule: admin-panel
✓ Loaded rule: api-v1

✓ Successfully loaded 5 rules to Redis
  Rule list key: marker:api:rules
  Redis address: localhost:6379
  Redis database: 0
```

## Configuration Format

The YAML config file defines:

```yaml
tag: api                          # Tag to match rules
logLevel: DEBUG                   # Log level
markerKey: X-MARK                 # Header key for marking
versionHeader: X-Version          # Version header name
identifyHeader: X-User-ID         # User ID header name
identifyCookie: user_id           # User ID cookie name
identifyQuery: uid                # User ID query param

redisConfig:
  enable: true
  addr: localhost:6379
  password: ""
  db: 0
  ruleListKeys: marker:api:rules
  refreshInterval: 15

staticRules:
  - tag: api
    name: rule-name
    enable: true
    priority: 100
    type: identify|version|canary|path
    markValue: mark-value
    # Type-specific fields:
    userIds: [user1, user2]        # For identify type
    minVersion: 1.0.0              # For version type
    maxVersion: 2.0.0              # For version type
    canary: 30                      # For canary type (0-100)
    path: /admin                    # For path type
```

## Redis Storage

After loading, rules are stored in Redis as:

```
marker:api:rules                  # List of rule keys
marker:api:rule:beta-users        # Hash with rule data
marker:api:rule:version-v2        # Hash with rule data
...
```

View rules in Redis:

```bash
redis-cli
> LRANGE marker:api:rules 0 -1
> HGETALL marker:api:rule:beta-users
```

## Testing

Run tests to verify the plugin works correctly:

```bash
go test -v -cover ./...
```

Expected output: 12 tests passing with ~40% coverage.
