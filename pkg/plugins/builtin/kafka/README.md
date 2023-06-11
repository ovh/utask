# `Kafka` plugin

This plugin publishes a message to a Kafka topic.

## Configuration

| Fields          | Description                                                                                                  |
|-----------------|--------------------------------------------------------------------------------------------------------------|
| `brokers`       | List of Kafka brokers (expected format: `HOSTNAME:PORT`).                                                    |
| `kafka_version` | Kafka version. Default version is `1.0.0.0`.                                                                 |
| `with_tls`      | use TLS when connecting to the broker(s).                                                                    |
| `sasl`          | a single object with `user` and `password` fields to enable SASL authentication.                             |
| `timeout`       | Timeout expressed as a duration (e.g `30s`). Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". |
| `message`       | Message to send to Kafka brokers.                                                                            |

## Example

An action of type `kafka` requires the following kind of configuration:

```yaml
action:
  type: kafka
  configuration:
    # mandatory, comma-separated string
    brokers: "localhost:9092,localhost:9093"
    # optional, default version is 1.0.0.0
    kafka_version: "1.0.0.0"
    # optional, if you need to use SASL authentication
    sasl:
      user: {{.config.kafka.sasl.user}}
      password: {{.config.kafka.sasl.password}}
    # optional, boolean
    with_tls: false
    # optional, valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    timeout: 10s
    # mandatory, topic, key and value fields. Key field is optional
    message:
      topic: "utask"
      # Optional, partition key to guarantee message ordering
      key: "hello_world"
      value: |
         {
           "message": "Hello world!"
         }
```

## Requirements

None by default. Sensitive data should stored in the configuration and accessed through `{{.config.[itemKey]}}` rather than hardcoded in your template.

## Resources

The `kafka` plugin declares automatically resources for its steps:
- `socket` to rate-limit concurrent execution on the number of open outgoing sockets
- `url:hostname` (where `hostname` is the broker destination host of the plugin configuration) to rate-limit concurrent execution on a specific broker.


## Return

### Output

| Name                 | Description                           |
|----------------------|---------------------------------------|
| `partition`          | The partition of the produced message |
| `offset`             | The offset of the produced message    |