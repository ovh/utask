# `cache` Plugin

This plugin provides a key-value cache backed by the database. Values can be stored with an optional TTL (time-to-live) in seconds. When a TTL is set, the entry is automatically cleaned up on read (lazy expiration) and at service startup (bulk purge). Setting a TTL of `0` (or omitting it) means the entry never expires.

## Configuration

| Field    | Description                                                                                                  |
| -------- | ------------------------------------------------------------------------------------------------------------ |
| `action` | `set` to store a value, `get` to retrieve a value, or `delete` to remove a value                             |
| `key`    | the cache key (required for all actions)                                                                      |
| `value`  | the value to store (only used with `set`; can be any JSON-compatible type: string, number, object, array...) |
| `ttl`    | time-to-live in seconds (only used with `set`; `0` or omitted means no expiration)                            |

## Example

Store a value with a 1-hour TTL:

```yaml
cache-set:
  action:
    type: cache
    configuration:
      action: set
      key: "my-key"
      value:
        foo: bar
        count: 42
      ttl: 3600
```

Retrieve the value:

```yaml
cache-get:
  dependencies:
    - cache-set
  action:
    type: cache
    configuration:
      action: get
      key: "my-key"
```

Delete the value:

```yaml
cache-delete:
  dependencies:
    - cache-get
  action:
    type: cache
    configuration:
      action: delete
      key: "my-key"
```

## Requirements

None.

## Return

### `set` action output

| Name     | Description                      |
| -------- | -------------------------------- |
| `key`    | the cache key that was set       |
| `cached` | `true` if the value was stored   |

### `get` action output

| Name    | Description                                                        |
| ------- | ------------------------------------------------------------------ |
| `key`   | the cache key that was looked up                                   |
| `hit`   | `true` if the key was found and not expired, `false` otherwise     |
| `value` | the cached value (or `null` if `hit` is `false`)                   |

### `delete` action output

| Name      | Description                      |
| --------- | -------------------------------- |
| `key`     | the cache key that was deleted   |
| `deleted` | `true` if the delete was issued  |
