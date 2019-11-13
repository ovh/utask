# `email` plugin

This plugin send an email.

## Configuration

An action of type `email` requires the following kind of configuration:

```yaml
action:
  type: email
  configuration:
    # mandatory, string
    smtp_username: {{.config.smtp.username}}
    # mandatory, string
    smtp_password: {{.config.smtp.password}}
    # optional, uint (default is 25)
    smtp_port: {{.config.smtp.port}}
    # mandatory, string
    smtp_hostname: {{.config.smtp.hostname}}
    # mandatory, string
    from: foo@ovh.cloud
    # mandatory, string
    to: [bar@ovh.cloud, hey@ovh.cloud]
    # mandatory, string
    subject: Hello from µTask
    # mandatory, string
    body: |
      I love baguette
```

## Note

Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.