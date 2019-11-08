# `http` Plugin

This plugin permorms an http request.

## Configuration

An action of type `http` requires the following kind of configuration:

```yaml
action:
  type: http
  configuration:
    url: http://example.org/user
    method: POST
    headers:
    - name:  Authorization
      value: Basic {{.config.basicAuth}}
    # body is optional, not used for method GET
    body: |
      {
        "name": "pablo"
      }
```

## Requirements

None by default. Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.