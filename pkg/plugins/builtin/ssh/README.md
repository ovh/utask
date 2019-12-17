# `ssh` Plugin

This plugin connects to a remote system and performs a block of commands. It can extract variables from the shell back to the output of its enclosing step.

## Configuration

|Fields|Description
|---|---
| `user` | username for the connection
| `target` | address of the remote machine
| `hops` | a list of intermediate addresses (bastions)
| `script` | multiline text, commands to be run on the machine's shell
| `result` | an object to extract the values of variables from the machine's shell
| `ssh_key` | private ssh key, preferrably retrieved from {{.config}}
| `ssh_key_passphrase` | passphrase for the key, if any
| `allow_exit_non_zero` | allow a non-zero exit code to be considered as a successful step (bool default `false`)

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
    # value extraction
    result: 
      pid: $PID
      uptime: $SERVICE_UPTIME
    # credentials
    ssh_key: '{{.config.mySSHKey}}'
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
  "exit_status": "0",
  "exit_signal": "0",
  "exit_msg": "exited 0"
}
```
