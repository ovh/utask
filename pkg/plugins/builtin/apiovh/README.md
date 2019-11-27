# `apiovh` Plugin

This plugin makes calls to the Public API of OVHCloud: [https://api.ovh.com](https://api.ovh.com).

## Configuration

|Field|Description
|---|---
| `path` | http route + query params
| `method` | http method (GET/POST/PUT/DELETE)
| `body` | a string representing the payload to be sent with the request
| `credentials` | a key to retrieve credentials from configstore

## Example

An action of type `apiovh` requires the following kind of configuration. The `body` field is optional:

```yaml
action:
  type: apiovh
  configuration:
    method: POST
    path: /dbaas/logs/{{.input.serviceName}}/output/graylog/stream
    credendials: ovh-api-credentials
    # body is optional, not used for method GET
    body: | 
      {
        "title": "{{.input.applicationName}}",
        "description": "{{.input.applicationDescription}}",  
        "autoSelectOption": true
      }
```

## Requirements

The `apiovh` plugin requires a config item to be found under the key given in the `credentials` config field. It's content should match the following schema (see [go-ovh](https://github.com/ovh/go-ovh) for more details): 

```js
{
  "endpoint": "ovh-eu",
  "appKey": "XXXX",
  "appSecret": "YYYY",
  "consumerKey": "ZZZZ"
}
```