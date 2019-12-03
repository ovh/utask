# `script` plugin

This plugin execute a script.

*Warn: This plugin will keep running until the execution is done*

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
    file: example.py
    # optional, a collection of string
    argv:
        - foo
        - bar
    # optional, string as uint
    # default is 300 seconds, 5 minutes
    timeout: "25"
```

## Note

The plugin returns two objects, the `Payload` who is the last returned line of your script as json:

```json
{
  "foo":"bar",
}
```

The `Metadata` to fetch informations about plugin execution:

```js
{
  "exit_code":"0",
  "exit_signal":"0",
  // Output is combinated /w Stdout and Stderr
  "output": "I'm a super happy script"
}
```
