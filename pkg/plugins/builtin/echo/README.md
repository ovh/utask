# `echo` Plugin

This plugin returns the output defined in its configuration, without performing any kind of work. It is useful for transforming and aggregating previous results, and for running [tests on the engine](https://github.com/ovh/utask/tree/master/engine/templates_tests).

## Configuration

An action of type `echo` requires the following kind of configuration. The default outcome is a successful step: by adding an error_message, the step will consider the action a failure and set the step in `SERVER_ERROR` state. This is the default error type, which will cause the step to be elligible for retry. To block the task, error_type can be set to `client`, causing the step state to be set to `CLIENT_ERROR`.

```yaml
action:
  type: echo
  configuration:
    output: # an arbitrary yaml object representing the final output of the step
      foo: '{{.input.foo}}-suffix'
      bar: 'prefix-{{.input.bar}}'
    metadata: # an arbitrary yaml object representing the step's returned metadata
      HTTPStatus: 200
    error_message: Epic fail! # an arbitrary error message
    error_type: client # client|server   
```

## Requirements

None.