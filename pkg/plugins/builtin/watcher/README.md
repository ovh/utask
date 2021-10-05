# `watcher` Plugin

This plugin updates the watcher usernames of the current task. New usernames are added to the list of existing one, ignoring any duplicate.

## Configuration

| Fields      | Description        |
| ----------- | ------------------ |
| `usernames` | an array of string |

## Example

An action of type `watcher` requires only one field, the list of watcher usernames to add to the current task.

```yaml
action:
  type: watcher
  configuration:
    usernames:
      - foo
      - bar
```
