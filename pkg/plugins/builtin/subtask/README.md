# `subtask` Plugin

This plugin creates a new task. A step based on this type of action will remain incomplete until the subtask is fully `DONE`.

## Configuration

|Fields|Description
|---|---
| `template` | the name of a task template, as accepted through µTask's  API
| `input` | a map of named values, as accepted on µTask's API
| `resolver_usernames` | a string containing a JSON array of additional resolver users for the subtask
| `watcher_usernames` | a string containing a JSON array of additional watcher users for the subtask
| `delay` | a duration indicating if subtask execution needs to be delayed, expects Golang time.Duration format (5s, 1m, ...)

## Example

An action of type `subtask` requires the following kind of configuration:

```yaml
action:
  type: subtask
  configuration:
    # a template that must already be registered on this instance of µTask
    template: another-task-template
    # valid input, as defined by the referred template
    input:
      foo: bar
    # optionally, a list of users which are authorized to resolve this specific task
    resolver_usernames: '["authorizedUser"]'
    watcher_usernames: '["authorizedUser"]'
    delay: 10m
```

## Requirements

None.
