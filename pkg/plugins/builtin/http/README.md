# `http` Plugin

This plugin permorms an http request.

## Configuration

An action of type `http` requires the following kind of configuration:

```yaml
action:
  type: http
  configuration:
    # mandatory
    url: http://example.org/user
    # mandatory
    method: POST
    # optional
    timeout_seconds: 5
    # optional
    basic_auth:
      user: {{.config.basicAuth.user}}
      password: {{.config.basicAuth.password}}
    # optional
    deny_redirects: false
    # optional
    parameters:
    - key: foo
      value: bar
    # optional
    headers:
    - name:  x-request-id
      value: xxx-yyy-zzz
    # optional
    body: |
      {
        "name": "pablo"
      }
```

## Requirements

None by default. Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.