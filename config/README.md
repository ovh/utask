# µTask config keys and files 🔨

## Configstore

For more complex data, µTask takes its configuration from the [configstore](https://github.com/ovh/configstore) library.

The source of configuration can be set through the `CONFIGURATION_FROM` environment variable. Read `docker-compose.yaml` file to see an example of configuration coming from the environment.

Configuration is stored in `items` with text content, each found under a `key`.

## Mandatory items

### Datasbase

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
    // resolver_usernames is a list of usernames with the privilege to resolve any task
    "resolver_usernames": ["user1"],
    // completed_task_expiration is a textual representation of how long a task is kept in DB after its completion
    "completed_task_expiration": "720h", // default == 720h == 30 days
    // notify_config contains a map of named notification configurations, composed of a type and config data, 
    // implemented notifiers include: 
    // - tat (github.com/ovh/tat)
    // - slack webhook (https://api.slack.com/messaging/webhooks)
    "notify_config": {
        "tat-internal": {
            "type": "tat",
            "config": {
                "username": "foo",
                "password": "very-secret",
                "url": "http://localhost:9999/tat",
                "topic": "utask.notifications"
            }
        },
        "slack-webhook": {
            "type": "slack",
            "config": {
                "webhook_url": "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
            }
        }
    },
    // notify_actions specifies a notification config for existing events in µTask
    // existing events are:
    // - task_state_action (fired every time a task's state changes)
    "notify_actions": {
        "task_state_action": {
            "disabled": false, // set to true to avoid sending out notification
            "notify_backends": ["tat-internal", "slack-webhook"] // choose among the named configs in notify_config, leave empty to broadcast on any notification backend
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
   // resource_limits allows you to define named resources and allocate a maximum number of concurrent actions on them (see Authoring task templates below)
   "resource_limits": {
       "openstack": 15,
   },
   // max_concurrent_executions defines a global maximum of concurrent actions running at any given time
   // if none provided, no upper bound is enforced
   "max_concurrent_executions": 100
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