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

The output resulting from this configuration will be:

```js
{
  "pid": "1234",
  "uptime": "876123"
}
```

## Requirements

None by default. Ssh credentials should be retrieved from `{{.config.[sshkey]}}` rather than hardcoded on the template.