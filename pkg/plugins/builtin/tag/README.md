# `tag` Plugin

This plugin updates the tags of the current task. Existing tags are overwritten with the values provided. An empty value deletes the tag.

## Configuration

|Fields|Description
| ------ | --------------- |
| `tags` | key/values tags |

## Example

An action of type `tag` requires only one field, the list of tags to apply to the current task.

```yaml
action:
  type: tag
  configuration:
    tags:
      foo: bar
      bar: # deleted
```
