# `script` plugin

This plugin execute a script.

*Warn: This plugin will keep running until the execution is done*

*Runtime(s) must be accessible on the host you deploy µTask if you want to execute interpreted scripts: [verify shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) and available packages*

Files must be located under scripts folder, you should set exec permissions (+x). Otherwise the script plugin will try to set the exec permissions.

## Configuration

|Fields|Description
|---|---
| `file_path` | file name under scripts folder
| `argv` | a collection of script argv
| `timeout` | timeout of the script execution
| `stdin` | inject stdin in your script
| `last_line_not_json` | skip or not unmarshaling of last JSON line
| `allow_exit_non_zero` | allow or not non zero exit status code

## Example

An action of type `script` requires the following kind of configuration:

```yaml
action:
  type: script
  configuration:
    # mandatory, string
    # file_path field must be related to you scripts path (./scripts)
    # and could modified /w `scripts-path` flag when you run binary
    file_path: hello-world.sh
    # optional, a collection of string
    argv:
        - world
    # optional, string as Duration
    # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    # default is 2m
    timeout: "25s"
    # optional, boolean
    # default is false, can't be templated
    last_line_not_json: false
    # optional, boolean
    # default is false, can't be templated
    allow_exit_non_zero: false
```

## Note

The plugin returns two objects, the `Payload` who is the last returned line of your script as json:

```json
{"dumb_string":"Hello world!","random_object":{"foo":"bar"}}
```

*Your JSON must be printed on last line*

The `Metadata` to fetch informations about plugin execution:

```js
{
  "exit_code":"0",
  "process_state":"exit status 0",
  // Output combine Stdout and Stderr streams without any distinction
  "output":"Hello world script\n{\"dumb_string\":\"Hello world!\",\"random_object\":{\"foo\":\"bar\"}}\n",
  "execution_time":"846.889µs",
  "error":""
}
```
