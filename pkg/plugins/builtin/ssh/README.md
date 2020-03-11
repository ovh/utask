# `ssh` Plugin

This plugin connects to a remote system and performs a block of commands. It can extract variables from the shell back to the output of its enclosing step.

The step will be considered successful if the script returns exit code 0, otherwise, it will be considered as a `SERVER_ERROR` (and will be retried). For unrecoverable errors (for instance, invalid parameters), it is possible to configure a list of exit codes (see `exit_codes_unrecoverable`) that should halt the execution (`CLIENT_ERROR`).

## Configuration

|Fields|Description
|---|---
| `user` | username for the connection
| `target` | address of the remote machine
| `hops` | a list of intermediate addresses (bastions)
| `script` | multiline text, commands to be run on the machine's shell
| `output_mode` | indicates how to retrieve the output values ; valid values are: `auto-result` (default), `disabled`, `manual-delimiters`, `manual-lastline`
| `result` | an object to extract the values of variables from the machine's shell (only used when `output_mode` is configured to `auto-result`)
| `output_manual_delimiters` | array of 2 strings ; look for a JSON formatted string in the script output between specific delimiters (only used when `output_mode` is configured to `manual-delimiters`)
| `ssh_key` | private ssh key, preferrably retrieved from {{.config}}
| `ssh_key_passphrase` | passphrase for the key, if any
| `exit_codes_unrecoverable` | a list of non-zero exit codes (1, 2, 3, ...) or ranges (1-10, ...) which should be considered unrecoverable and halt execution ; these will be returned to the main engine as a `CLIENT_ERROR`

## Example

An action of type `ssh` requires the following kind of configuration:

```yaml
action:
  type: ssh
  configuration:
    # user name connecting to the machine
    user: ubuntu
    # target machine
    target: frontend.ha.example.org
    # intermediate machines
    hops:
    - bastion.example.org
    # the commands to be executed, with variable declarations for value extraction
    script: |-
      PID=$(systemctl show --property MainPID {{.input.serviceName}} | cut -d= -f2)
      SERVICE_UPTIME=$(ps -h -p ${PID} -o etimes)
    # this is the default mode
    output_mode: auto-result
    # value extraction
    result:
      pid: $PID
      uptime: $SERVICE_UPTIME
    # credentials
    ssh_key: '{{.config.mySSHKey}}'
    # optional delimiters to look for an output -- requires output_mode set to manual-delimiters
    #output_manual_delimiters: ["JSON_START", "JSON_END"]
    exit_codes_unrecoverable:
      - "1-10"
      - "100"
      - "110"
```

## Requirements

None by default. Ssh credentials should be retrieved from `{{.config.mySSHKey}}` rather than hardcoded on the template.

## Note

The plugin returns two objects, the `Payload` of the execution as defined in the result field of configuration:

```json
{
  "pid": "3931",
  "uptime": "1715606"
}
```

The `Metadata` to fetch informations about plugin execution:

```json
{
  "output": "Connecting...\nWelcome to Ubuntu 19.04 (GNU/Linux ... x86_64)\n[...]\n{\"pid\":\"3931\",\"service_name\":\"nginx\",\"service_uptime\":\"1715606\"}",
  "exit_code": "0",
  "exit_signal": "0",
  "exit_msg": "exited 0"
}
```
