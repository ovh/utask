# µTask config keys and files 🔨

## Configstore

For more complex data, µTask takes its configuration from the [configstore](https://github.com/ovh/configstore) library.

The source of configuration can be set through the `CONFIGURATION_FROM` environment variable. Read `docker-compose.yaml` file to see an example of configuration coming from the environment.

Configuration is stored in `items` with text content, each found under a `key`.

## Mandatory items

### Database

`database` key is a postgres connection string from configstore.

```
postgres://user:pass@db/utask?sslmode=disable
```

### Encryption-key

`encryption-key` key is an encryption key labelled `storage` formatted for the [symmecrypt library](https://github.com/ovh/symmecrypt) from configstore.

```js
{
    "identifier": "storage",
    "cipher": "aes-gcm",
    "timestamp": 1535627466,
    "key": "e5f45aef9f072e91f735547be63f3434e6de49695b178e3868b23b0e32269800"
}
```

### Utask-cfg

`utask-cfg` key is a json-formatted structure with global configuration values from configstore.

```js
{
    // application_name is a publicly visible, human readable identifier for this instance of µTask
    "application_name": "µTask Foo",
    // admin_usernames is a list of usernames with admin privileges over µTask resources, ie. the ability to view and execute any task, and to hotfix resolutions if a problem arises
    "admin_usernames": ["admin1", "admin2"],
    // admin_groups is a list of user groups with admin privileges over µTask resources, ie. the ability to view and execute any task, and to hotfix resolutions if a problem arises
    "admin_groups": ["administrators", "maintainers"],
    // completed_task_expiration is a textual representation of how long a task is kept in DB after its completion
    "completed_task_expiration": "720h", // default == 720h == 30 days
    // notify_config contains a map of named notification configurations, composed of a type and config data,
    // implemented notifiers include:
    // - opsgenie (https://www.atlassian.com/software/opsgenie); available zones are: global, eu, sandbox
    // - slack webhook (https://api.slack.com/messaging/webhooks)
    // - generic webhook (custom URL, with HTTP POST method)
    // notification strategies can be declared per backend:
    // - template_notification_strategies is an array of strategy per template
    // - default_notification_strategy is the strategy that will apply, if none matched above
    // available strategies are: always, failure_only, silent
    "notify_config": {
        "opsgenie-eu": {
            "type": "opsgenie",
            "config": {
                "zone": "eu",
                "api_key": "very-secret", 
                "timeout": "30s"
            },
            "default_notification_strategy": {
                "task_state_update": "failure_only"
            }
        },
        "tat-internal": {
            "type": "tat",
            "config": {
                "username": "foo",
                "password": "very-secret",
                "url": "http://localhost:9999/tat",
                "topic": "utask.notifications"
            },
            "default_notification_strategy": {
                "task_state_update": "silent",
                "task_validation": "always"
            },
            "template_notification_strategies": {
                "task_state_update": [
                    {
                        "templates": ["hello-world", "hello-world-2"],
                        "notification_strategy": "always"
                    }
                ]
            }
        },
        "slack-webhook": {
            "type": "slack",
            "config": {
                "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
            },
            "default_notification_strategy": {
                "task_state_update": "failure_only"
            },
        },
        "webhook-example.org": {
            "type": "webhook",
            "config": {
                "webhook_url": "https://example.org/webhook/XXXXXXXXXXXXXXXXXXXX",
                "username": "foo",
                "password": "very-secret",
                "headers": {
                    "X-Specific-Header": "foobar"
                }
            }
        }
    },
    // notify_actions specifies a notification config for existing events in µTask
    // existing events are:
    // - task_state_update: fired every time a task's state changes
    // - task_validation: fired every time a new task is created and requires a human validation
    // - task_step_update: fired every time a step's state changes
    "notify_actions": {
        "task_state_update": {
            "disabled": false, // set to true to avoid sending out notification
            "notify_backends": ["tat-internal", "slack-webhook"] // choose among the named configs in notify_config, leave empty to broadcast on any notification backend
        },
        "task_validation": {
            "disabled": false, // set to true to avoid sending out notification
            "notify_backends": ["slack-webhook"] // choose among the named configs in notify_config, leave empty to broadcast on any notification backend
        },
        "task_step_update": {
            "disabled": true // set to true to avoid sending out notification
        }
    },
    // database_config holds configuration to fine-tune DB connection
    "database_config": {
        "max_open_conns": 50, // default 50
        "max_idle_conns": 30, // default 30
        "conn_max_lifetime": 60, // default 60, unit: seconds
        "config_name": "database" // configuration entry where connection info can be found, default "database"
    },
    // concealed_secrets allows you to render some configstore items inaccessible to the task engine
    "concealed_secrets": ["database", "encryption-key", "utask-cfg"],
    // resource_limits allows you to define named resources and allocate a maximum number of concurrent actions on them (see Authoring task templates in /README.md)
    "resource_limits": {
        "openstack": 15,
        "socket": 1024,
        "fork": 50,
        "url:example.org": 10
    },
    // max_concurrent_executions defines a global maximum of concurrent tasks running at any given time
    // default value: 100; 0 will stop all tasks processing; -1 to indicate no limit
    "max_concurrent_executions": 100,
    // max_concurrent_executions_from_crashed defines a maximum of concurrent tasks from a crashed instance running at any given time
    // default value: 20; 0 will stop all tasks processing; -1 to indicate no limit
    "max_concurrent_executions_from_crashed": 20,
    // delay_between_crashed_tasks_resolution defines a wait duration between two tasks from a crashed instance will be schedule in the current uTask instance
    // default 1, unit: seconds
    "delay_between_crashed_tasks_resolution": 1,
    // base_url defines the base URL for the µTask UI. It's used for determining the public URL of a task, for notification purposes. dashboard_path_prefix will be appended to this URL.
    "base_url": "https://utask.example.org",
    // dashboard_path_prefix defines the path prefix for the dashboard UI. Should be used if the uTask instance is hosted with a ProxyPass, on a custom path
    // default: empty, no prefix
    "dashboard_path_prefix": "/my-utask-instance",
    // dashboard_api_path_prefix defines the path prefix for the uTask API. Should be used if the uTask instance is hosted with a ProxyPass, on a custom path.
    // dashboard_api_path_prefix will be used by Dashboard UI to contact the uTask API
    // default: empty, no prefix
    "dashboard_api_path_prefix": "/my-utask-instance",
    // dashboard_sentry_dsn defines the Sentry DSN for the Dashboard UI. Used to retrieve Javascript execution errors inside a Sentry instance.
    // default: empty, no SENTRY_DSN
    "dashboard_sentry_dsn": "",
    // steps_compression_algorithm defines the compression algorithm to use to compress the steps data in database.
    // default: empty, no compression. Available compression algorithms: gzip
    "steps_compression_algorithm": "",
    // server_options holds configuration to fine-tune DB connection
    "server_options": {
        // max_body_bytes defines the maximum size that will be read when sending a body to the uTask server.
        // value can't be smaller than 1KB (1024), and can't be bigger than 10MB (10*1024*1024)
        // default: 262144 (256KB), unit: byte
        "max_body_bytes": 262144
    }
}
```

## Optional items

### Basic auth

`basic-auth` key is only for development purposes: this allows the graphical dashboard on your browser to provide you a basic authentication scheme for calls made to the backend. It consists of a map of usernames and passwords.

```js
{
    "admin": "1234"
}
```


### Groups auth

`groups-auth` key is only for development purposes: it is used to declare the users of each group. 
It consists of a map of group names and slices of usernames.

```json
{
    "administrators": ["admin"],
    "maintainers": ["admin"]
}
```
