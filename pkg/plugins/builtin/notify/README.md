# `notify` Plugin

This plugin sends a message over any of the notification channels defined in µTask's configuration.

## Configuration

An action of type `notify` requires the following kind of configuration:

```yaml
action:
  type: echo
  configuration:
    # the payload of your notification
    message: Hello World! 
    # a list of extra fields as map of string, to contextualize your message
    fields: 
      randomfield: urgent 
      language: english
    # a list of destination backends as defined in 'utask-cfg' (will be sent to ALL backends if left empty or null)
    backends: [tat-internal, slack-customers] 
```

## Requirements

Configuration for at least one notification backend should be provided in the config item named `utask-cfg` (see [config/README.md](https://github.com/ovh/utask/blob/master/config/README.md)).