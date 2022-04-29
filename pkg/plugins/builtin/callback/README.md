# `callback` Plugin

This plugin allows to create callbacks. A callback can be created from a step using this plugin. Once the callback has been created, you can
pause your resolution until the callback has been successfully called.

## Configuration

| Fields               | Description                                                                                                       |
| -------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `action  `           | `create` to create a callback or `wait` to wait a callback                                                        |
| `schema`             | only valid if `action` is `create`: validate the body provided during the call of the callback                    |
| `id`                 | only valid if `action` is `wait`: ID of the callback to wait                                                      |

## Example

First, the callback has to be created:

```yaml
create-cb:
  action:
    type: callback
    configuration:
      action: create
      schema: |-
        {
          "$schema": "http://json-schema.org/schema#",            
          "type": "object",
          "additionalProperties": false,
          "required": ["success"],
          "properties": {
            "success": {
              "type": "boolean"
            }
          }
        }
```

In a second step, you can wait for the callback resolution:

```yaml
wait-cb:
  dependencies:
    - create-cb
  action:
    type: callback
    configuration:
      action: wait
      id: '{{field `step` `create-cb` `output` `id`}}'
```

## Requirements

The base URl for callbacks must be defined in `callback-config` configuration key. The value must be a map with at least the `base_url` key which
contains the base URL to reach the callback API from callers. You can also append a `path_prefix` key to override the default value (`/unsecured/callback/`).

### Examples

In those examples, `<ID>` will be the callback ID and `<T>` the callback token.

| Base URL          | Path prefix | URL                                             |
| ----------------- | ----------- | ----------------------------------------------- |
| `https://foo.bar` | `-`         | `https://foo.bar/unsecured/callback/<ID>?t=<T>` |
| `https://foo.bar` | `/`         | `https://foo.bar/<ID>?t=<T>`                    |
| `https://foo.bar` | `/foobar/`  | `https://foo.bar/foobar/<ID>?t=<T>`             |

## Return

### Callback `create` action output

| Name     | Description                           |
| -------- | ------------------------------------- |
| `id`     | The public identifier of the callback |
| `url`    | The public URL of the callback        |
| `schema` | The sanitized schema                  |

### Callback `wait` action output

| Name   | Description                           |
| ------ | ------------------------------------- |
| `id`   | The public identifier of the callback |
| `date` | The call date                         |
| `body` | The provided body during the call     |
