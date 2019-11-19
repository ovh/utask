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
    # mandatory, string as uint
    smtp_port: {{.config.smtp.port}}
    # mandatory, string
    smtp_hostname: {{.config.smtp.hostname}}
    # optional, string as boolean
    smtp_skip_tls_verify: "true"
    # mandatory, string
    from_address: foo@example.org
    # optional, string
    from_name: uTask bot
    # mandatory, string collection
    to: [bar@example.org, hey@example.org]
    # mandatory, string
    subject: Hello from µTask
    # mandatory, string
    body: |
      I love baguette
```

## Note

The plugin returns an object to reuse the parameters in a future component:

```json
{
  "from_address":"foo@example.org",
  "from_name":"uTask bot",
  "to": ["bar@example.org", "hey@example.org"],
  "subject":"Hello from µTask",
  "body":"I love baguette"
}
```

Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.