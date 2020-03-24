export default {
    initValue: {
        name: "template-name",
        description: "Example template",
        title_format: "Run a task for {{.input.id}}",

        auto_runnable: false,

        inputs: [
            {
                name: "id",
                description: "Example input",
                optional: true
            }
        ],

        steps: {
            stepOne: {
                description: 'Step number one',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'http',
                    configuration: {
                        url: 'http://localhost:8080',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`,
                        headers: [
                            {
                                name: 'content-type',
                                value: 'application/json'
                            }
                        ]
                    }
                }
            },
            stepTwo: {
                description: 'Step number two, ',
                dependencies: ['stepOne'],
                retry_pattern: 'seconds',
                action: {
                    type: 'http',
                    configuration: {
                        url: 'http://localhost:8080',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`,
                        headers: [
                            {
                                name: 'content-type',
                                value: 'application/json'
                            }
                        ]
                    }
                }
            }
        }
    },
    types: {
        gw: {
            name: 'GW',
            value: {
                description: 'Make a request on a microservice',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'gw',
                    configuration: {
                        serviceName: 'my-service',
                        path: '/foo',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            },
            snippet: {
                description: '${2:Make a request on a microservice}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'gw',
                    configuration: {
                        serviceName: '${3:my-service}',
                        path: '${4:/foo}',
                        method: '${5:POST}',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            }
        },
        apiovh: {
            name: 'OVH API',
            value: {
                description: 'Make a request on the public OVH API',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'apiovh',
                    configuration: {
                        credentials: 'configstore-item-key',
                        path: '/foo',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            },
            snippet: {
                description: '${2:Make a request on the public OVH API}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'apiovh',
                    configuration: {
                        credentials: '${3:configstore-item-key}',
                        path: '${4:/foo}',
                        method: '${5:POST}',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            }
        },
        http: {
            name: 'HTTP',
            value: {
                description: 'Make an http request',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'http',
                    configuration: {
                        url: 'http://localhost:8080',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`,
                        headers: [
                            {
                                name: 'content-type',
                                value: 'application/json'
                            }
                        ]
                    }
                }
            },
            snippet: {
                description: '${3:Make an http request}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'http',
                    configuration: {
                        url: '${4:http://localhost:8080}',
                        method: '${5:POST}',
                        body: `{
    "foo": "bar"
}`,
                        headers: [
                            {
                                name: 'content-type',
                                value: 'application/json'
                            }
                        ]
                    }
                }
            }
        },
        locker: {
            name: 'Locker',
            value: {
                description: 'Make a request on Locker',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'locker',
                    configuration: {
                        credentials: 'locker-user-pass',
                        path: '/foo',
                        method: 'POST',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            },
            snippet: {
                description: '${2:Make a request on Locker}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'locker',
                    configuration: {
                        credentials: '${3:locker-user-pass}',
                        path: '${4:/foo}',
                        method: '${5:POST}',
                        body: `{
    "foo": "bar"
}`
                    }
                }
            }
        },
        notify: {
            name: 'Notify',
            value: {
                description: 'Send a message on TAT',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'notify',
                    configuration: {
                        message: 'This is a notification!',
                        fields: ['foo','bar','baz']
                    }
                }
            },
            snippet: {
                description: '${2:Send a message on TAT}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'notify',
                    configuration: {
                        message: '${3:This is a notification!}',
                        fields: ['foo','bar','baz']
                    }
                }
            }
        },
        ssh: {
            name: 'SSH',
            value: {
                description: 'Execute commands over SSH',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'ssh',
                    configuration: {
                        user: 'ubuntu',
                        target: '1.1.1.1',
                        hops: ['2.2.2.2','3.3.3.3'],
                        script: `UPTIME=$(uptime)
# other shell commands...`,
                        result: {"uptime":"\\$UPTIME"},
                        ssh_key: "{{ .secret.mysshkey }}",
                        ssh_key_passphrase: "{{ .secret.mysshkeypassphrase }}"
                    }
                }
            },
            snippet: {
                description: '${2:Execute commands over SSH}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'ssh',
                    configuration: {
                        user: '${3:ubuntu}',
                        target: '${4:1.1.1.1}',
                        hops: ['2.2.2.2','3.3.3.3'],
                        script: `UPTIME=$(uptime)
# other shell commands...`,
                        result: {"uptime":"\\$UPTIME"},
                        ssh_key: "{{ ${9:.secret.mysshkey} }}",
                        ssh_key_passphrase: "{{ ${10:.secret.mysshkeypassphrase} }}"
                    }
                }
            }
        },
        subtask: {
            name: 'Subtask',
            value: {
                description: 'Spawn a new Task',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'subtask',
                    configuration: {
                        template: 'my-template-name',
                        input: {
                            "foo": "bar"
                        }
                    }
                }
            },
            snippet: {
                description: '${2:Spawn a new Task}',
                dependencies: [],
                retry_pattern: 'seconds',
                action: {
                    type: 'subtask',
                    configuration: {
                        template: '${3:my-template-name}',
                        input: {
                            "${4:foo}": "${5:bar}"
                        }
                    }
                }
            }
        }
    }
};