# `subtask` Plugin

This plugin creates a new task. A step based on this type of action will remain incomplete until the subtask is fully `DONE`.

## Configuration

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
    resolver_usernames: [authorizedUser]  
```

## Requirements

None.
