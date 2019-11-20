# `http` Plugin

This plugin permorms an http request.

## Configuration

An action of type `http` requires the following kind of configuration:

```yaml
action:
  type: http
  configuration:
    # mandatory, string
    url: http://example.org/user
    # mandatory, string
    method: POST
    # optional, string as uint16
    timeout_seconds: "5"
    # optional, object of user and password fields
    basic_auth:
      user: {{.config.basicAuth.user}}
      password: {{.config.basicAuth.password}}
    # optional, string as boolean
    deny_redirects: "false"
    # optional, array of key and value fields
    parameters:
    - key: foo
      value: bar
    # optional, array of name and value fields
    headers:
    - name:  x-request-id
      value: xxx-yyy-zzz
    # optional, string
    body: |
      {
        "name": "pablo"
      }
```

## Requirements

None by default. Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.