# `email` plugin

This plugin send an email.

## Configuration

| Fields                | Description                                   |
| --------------------- | --------------------------------------------- |
| `smtp_username`       | username of SMTP server                       |
| `smtp_password`       | password of SMTP server                       |
| `smtp_port`           | port of SMTP server                           |
| `smtp_hostname`       | hostname of SMTP server                       |
| `smtp_skip_tls_verif` | Skip or not TLS insecure verify               |
| `from_address`        | from which email you want to send the message |
| `from_name`           | from which name you want to send the message  |
| `to`                  | receiver(s) of your email                     |
| `subject`             | subject of your email                         |
| `body`                | content of your email                         |
| `attachments`         | file names of files to attach to the message  |

## Example

An action of type `email` requires the following kind of configuration:

```yaml
action:
  type: email
  configuration:
    # optional, string, leave empty for no auth
    smtp_username: {{.config.smtp.username}}
    # optional, string, leave empty for no auth
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
    attachments:
      - /tmp/generated-report.xlsx
```

## Note

The plugin returns an object to reuse the parameters in a future component:

```json
{
  "from_address": "foo@example.org",
  "from_name": "uTask bot",
  "to": ["bar@example.org", "hey@example.org"],
  "subject": "Hello from µTask",
  "body": "I love baguette",
  "attachments": ["/tmp/generated-report.xlsx"]
}
```

Sensitive data should be retrieved from configstore and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.

## Resources

The `email` plugin declares automatically resources for its steps:
- `socket` to rate-limit concurrent execution on the number of open outgoing sockets
- `url:smtp_hostname` (where `smtp_hostname` is the outgoing SMTP server of the plugin configuration) to rate-limit concurrent execution on a specific outgoing SMTP server
