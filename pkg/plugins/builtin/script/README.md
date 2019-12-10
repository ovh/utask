# `script` plugin

This plugin execute a script.

*Warn: This plugin will keep running until the execution is done*

*Runtime(s) must be accessible on the host you deploy µTask if you want to execute interpreted scripts: [verify shebang](https://en.wikipedia.org/wiki/Shebang_(Unix)) and available packages*

*Files must be located under scripts folder with exec (+x) permissions*

## Configuration

|Fields|Description
|---|---
| `file` | file name under scripts folder
| `argv` | a collection of script argv
| `timeout` | timeout of the script execution

## Example

An action of type `script` requires the following kind of configuration:

```yaml
action:
  type: script
  configuration:
    # mandatory, string
    file: hello-world.sh
    # optional, a collection of string
    argv:
        - world
    # optional, string as Duration
    # Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
    # default is 2m
    timeout: "25s"
```

## Note

The plugin returns two objects, the `Payload` who is the last returned line of your script as json:

```json
{"msg":"Hello world!","error":{"message":"nil"}}
```

*Your JSON must be printed on last line*

The `Metadata` to fetch informations about plugin execution:

```js
{
  "exit_code":"0",
  "process_state":"exit status 0",
  // Output is combinated /w Stdout and Stderr
  "output":"Hello world script\n{\"msg\":\"Hello world!\",\"error\":{\"message\":\"nil\"}}\n",
  "execution_time":"846.889µs",
  "error":""
}
```
