# `http` Plugin

This plugin permorms an HTTP request.

## Configuration

| Fields                 | Description                                                                                                                                                                                                                                                      |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `url`                  | destination for the http call, including host, path and query params; this all-in-one field conflicts with `host` and `path`                                                                                                                                     |
| `host`                 | destination host for the http call; this field conflicts with the all-in-one field `url`                                                                                                                                                                         |
| `path`                 | path for the http call; to use jointly with the `host` field; this field conflicts with the all-in-one field `url`                                                                                                                                               |
| `method`               | http method (`GET`, `POST`, `PUT`, `DELETE`, `PATCH`)                                                                                                                                                                                                            |
| `body`                 | a string representing the payload to be sent with the request                                                                                                                                                                                                    |
| `headers`              | a list of headers, represented as (`name`, `value`) pairs                                                                                                                                                                                                        |
| `timeout`              | timeout expressed as a duration (e.g. `30s`)                                                                                                                                                                                                                     |
| `auth`                 | a single object composed of either a `basic` object with `user` and `password` fields to enable HTTP basic auth, or a `bearer` field to enable Bearer Token Authorization, or a `mutual_tls` object to enable Mutual TLS authentication                          |
| `follow_redirect`      | if `true` (string) the plugin will follow up to 10 redirects (302, ...)                                                                                                                                                                                          |
| `query_parameters`     | a list of query parameters, represented as (`name`, `value`) pairs; these will appended the query parameters present in the `url` field; parameters can be repeated (in either `url` or `query_parameters`) which will produce e.g. `?param=value1&param=value2` |
| `trim_prefix`          | prefix in the response that must be removed before unmarshalling (optional)                                                                                                                                                                                      |
| `insecure_skip_verify` | If `true` (string), disables server's certificate chain and host verification.                                                                                                                                                                                   |

## Example

An action of type `http` requires the following kind of configuration:

```yaml
action:
  type: http
  configuration:
    # mandatory, string
    url: http://example.org/user?lang=en
    # mandatory, string
    method: POST
    # optional, string as duration
    timeout: "5s"
    # optional, authentication you can use either basic or bearer auth
    auth:
      basic:
        user: {{.config.basicAuth.user}}
        password: {{.config.basicAuth.password}}
      bearer: {{.config.auth.token}}
      mutual_tls:
        # a chain of certificates to identify the caller, first certificate in the chain is considered as the leaf, followed by intermediates
        client_cert: {{.config.mtls.clientCert}}
        # private key corresponding to the certificate
        client_key: {{.config.mtls.clientKey}}
    # optional, string as boolean
    follow_redirect: "true"
    # optional, defines additional root CAs to perform the call. can contains multiple CAs concatained together
    root_ca: {{.config.mtls.rootca}}
    # optional, string as boolean. indicates if server certificate must be validated or not.
    insecure_skip_verify: "false"
    # optional, array of name and value fields
    query_parameters:
    - name: foo
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

None by default. Sensitive data should stored in the configuration and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.

## Resources

The `http` plugin declares automatically resources for its steps:
- `socket` to rate-limit concurrent execution on the number of open outgoing sockets
- `url:hostname` (where `hostname` is e.g. `www.ovh.com`) to rate-limit concurrent execution on a specific destination web-server/API
