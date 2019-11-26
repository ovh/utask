# `ping` plugin

This plugin send a ping.

*Warn: This plugin will keep running until the count is done*

## Configuration

|Field|Description
|---|---
| `hostname` | ping destination
| `count` | number of ping you want execute
| `interval_second` | interval between two pings

## Example

An action of type `ping` requires the following kind of configuration:

```yaml
action:
  type: ping
  configuration:
    # mandatory, string
    hostname: example.org
    # mandatory, string as uint
    count: "2"
    # mandatory, string as uint
    interval_second: "1"
```

## Note

The plugin returns two objects, the `Output` to fetch statistics about ping(s):

```json
{
  "packets_received":1,
  "packets_sent":1,
  "packet_loss": 0.00,
  "ip_addr":"192.168.0.1",
  "rtts":"1s",
  "min_rtt":"1se",
  "max_rtt":"1s",
  "avg_rtt":"1s",
  "std_dev_rtt":"1s"
}
```

The `Metadata` to reuse the parameters in a future component:

```json
{
  "hostname":"example.org",
  "count":"2",
  "interval_second": "1"
}
```
