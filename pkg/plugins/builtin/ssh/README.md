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

None by default. Ssh credentials should be retrieved from `{{.config.[sshkey]}}` rather than hardcoded on the template.

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
  "output": "Connecting...
    Pseudo-terminal will not be allocated because stdin is not a terminal.
    Welcome to Ubuntu 19.04 (GNU/Linux 5.0.0-31-generic x86_64)

     * Documentation:  https://help.ubuntu.com
     * Management:     https://landscape.canonical.com
     * Support:        https://ubuntu.com/advantage

      System information as of Mon Dec  9 11:03:03 UTC 2019

      System load:  0.0               Processes:           111
      Usage of /:   23.7% of 9.52GB   Users logged in:     0
      Memory usage: 18%               IP address for ens3: 10.00.00.01
      Swap usage:   0%

     * Overheard at KubeCon: \"microk8s.status just blew my mind\".

         https://microk8s.io/docs/commands#microk8s.status

    26 updates can be installed immediately.
    0 of these updates are security updates.


    *** System restart required ***
    {\"pid":"3931\",\"service_name\":\"nginx\",\"service_uptime\":\"1715606\"}",
  "exit_status": "0",
  "exit_signal": "0",
  "exit_msg": "exited 0"
}
```
