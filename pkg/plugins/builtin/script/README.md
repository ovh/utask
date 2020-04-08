# `script` plugin

This plugin execute a script.

*Warn: This plugin will keep running until the execution is done*

*Runtime(s) must be accessible on the host you deploy µTask if you want to execute interpreted scripts: [verify shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) and available packages*

Files must be located under scripts folder, you should set exec permissions (+x). Otherwise the script plugin will try to set the exec permissions.

The step will be considered successful if the script returns exit code 0, otherwise, it will be considered as a `SERVER_ERROR` (and will be retried). For unrecoverable errors (for instance, invalid parameters), it is possible to configure a list of exit codes (see `exit_codes_unrecoverable`) that should halt the execution (`CLIENT_ERROR`).


## Configuration

|Fields|Description
|---|---
| `file_path` | file name under scripts folder
| `argv` | a collection of script argv
| `timeout` | timeout of the script execution
| `stdin` | inject stdin in your script
| `output_mode` | indicates how to retrieve the output values ; valid values are: `manual-lastline` (default), `disabled`, `manual-delimiters`
| `output_manual_delimiters` | array of 2 strings ; look for a JSON formatted string in the script output between specific delimiters (only used when `output_mode` is configured to `manual-delimiters`)
| `exit_codes_unrecoverable` | a list of non-zero exit codes (1, 2, 3, ...) or ranges (1-10, ...) which should be considered unrecoverable and halt execution ; these will be returned to the main engine as a `CLIENT_ERROR`

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
    # this is the default mode
    output_mode: manual-lastline
    # optional delimiters to look for an output -- requires output_mode set to manual-delimiters
    #output_manual_delimiters: ["JSON_START", "JSON_END"]
    exit_codes_unrecoverable:
      - "1-10"
      - "100"
      - "110"
```

## Note

The plugin returns two objects, `output` and `metadata`.

`output` depends on the `output_mode` configuration. It is read from the stdout of the script, either on the last line (`output_mode` set to `manual-lastline`) or between given delimiters (`output_mode` set to `manual-delimiters`, delimiters defined in `output_manual_delimiters`).

```json
{"dumb_string":"Hello world!","random_object":{"foo":"bar"}}
```

`metadata` contains information about the script execution:

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
