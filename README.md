# µTask, the Lighweight Automation Engine

[![Build Status](https://travis-ci.org/ovh/utask.svg?branch=master)](https://travis-ci.org/ovh/utask)
[![Go Report Card](https://goreportcard.com/badge/github.com/ovh/utask)](https://goreportcard.com/report/github.com/ovh/utask)
[![Coverage Status](https://coveralls.io/repos/github/ovh/utask/badge.svg?branch=master)](https://coveralls.io/github/ovh/utask?branch=master)
[![GoDoc](https://godoc.org/github.com/ovh/utask?status.svg)](https://godoc.org/github.com/ovh/utask)
[![GitHub stars](https://img.shields.io/github/stars/ovh/utask)](https://github.com/ovh/utask/stargazers)
![GitHub last commit](https://img.shields.io/github/last-commit/ovh/utask)
[![GitHub license](https://img.shields.io/github/license/ovh/utask)](https://github.com/ovh/utask/blob/master/LICENSE)
 
µTask is an automation engine built for the cloud. It is:
- **simple to operate**: only a postgres DB is required
- **secure**: all data is encrypted, only visible to authorized users
- **extensible**: you can develop custom actions in golang

µTask allows you to model business processes in a **declarative yaml format**. Describe a set of inputs and a graph of actions and their inter-dependencies: µTask will asynchronously handle the execution of each action, working its way around transient errors and keeping a trace of all intermediary states until completion.

## Table of contents

- [Quick Start](#quickstart)
- [Configuration](#configuration)
- [Authoring Task Templates](#templates)
- [Extending µTask with plugins](#plugins)
- [Contributing](#contributing)

## Quick start <a name="quickstart"></a>

Build standalone µTask binary:

```bash
$ make all
```

Boot µTask along with its postgres database:

```bash
$ docker-compose up
```

Go to `http://localhost:8081/ui/dashboard` on your browser, or explore the API schema on `http://localhost:8081/unsecured/spec.json`.

Request a new task:
![](./assets/img/utask_new_task.png)

Get an overview of all tasks:
![](./assets/img/utask_dashboard.png)

Get a detailed view of a running task:
![](./assets/img/utask_running.png)

Browse available task templates:
![](./assets/img/utask_templates.png)

## Configuration 🔨 <a name="configuration"></a>

### Command line args 

The µTask binary accepts the following arguments as binary args or env var.
All are optional and have a default value:
- `init-path`: the directory from where initialization plugins (see "Developing plugins") are loaded in *.so form (default: `./init`)
- `plugins-path`: the directory from where action plugins (see "Developing plugins") are loaded in *.so form (default: `./plugins`)
- `templates-path`: the directory where yaml-formatted task templates are loaded from (default: `./templates`)
- `region`: an arbitrary identifier, to aggregate a running group of µTask instances (commonly containers), and differentiate them from another group, in a separate region (default: `default`)
- `http-port`: the port on which the HTTP API listents (default: `8081`)
- `debug`: a boolean flag to activate verbose logs (default: `false`)
- `maintenance-mode`: a boolean to switch API to maintenance mode (default: `false`)

### Config keys and files

Checkout the [µTask config keys and files README](./config/README.md).

### Authentication

The vanilla version of µTask doesn't handle authentication by itself, it is meant to be placed behind a reverse proxy that provides a username through the "x-remote-user" http header. A username found there will be trusted as is, and used for authorization purposes (admin actions, task resolution, etc...).

For development purposes, an optional `basic-auth` configstore item can be provided to define a mapping of usernames and passwords. This is not meant for use in production.

Extending this basic authentication mechanism is possible by developing an "init" plugin, as described below.

## Examples

Checkout the [µTask examples directory](./examples).

## Authoring Task Templates <a name="templates"></a>

A process that can be executed by µTask is modelled as a `task template`: it is written in yaml format and describes a sequence of steps, their interdepencies, and additional conditions and constraints to control the flow of execution.

The user that creates a task is called `requester`, and the user that executes it is called `resolver`. Both can be the same user in some scenarios.

A user can be allowed to resolve a task in three ways:
- the user is included in the global configuration's list of `admin_usernames`
- the user is included in the global configuration's list of `resolver_usernames`
- the user is included in the task's template list of `allowed_resolver_usernames`

### Value Templating

µTask uses the go [templating engine](https://golang.org/pkg/text/template/) in order to introduce dynamic values during a task's execution. As you'll see in the example template below, template handles can be used to access values from different sources. Here's a summary of how you can access values through template handles:
- `.input.[INPUT_NAME]`: the value of an input provided by the task's requester
- `.resolver_input.[INPUT_NAME]`: the value of an input provided by the task's resolver
- `.step.[STEP_NAME].output.foo`: field `foo` from the output of a named step
- `.step.[STEP_NAME].metadata.HTTPStatus`: field `HTTPStatus` from the metadata of a named step
- `.config.[CONFIG_ITEM].bar`: field `bar` from a config item (configstore, see above)
- `.iterator.foo`: field `foo` from the iterator in a loop (see `foreach` steps below)

### Basic properties

- `name`: a short unique human-readable identifier
- `description`: sentence-long description of intent
- `long_description`: paragraph-long basic documentation
- `doc_link`: URL for external documentation about the task
- `title_format`: templateable text, generates a title for a task based on this template
- `result_format`: templateable map, used to generate a final result object from data collected during execution

### Advanced properties

- `allowed_resolver_usernames`: a list of usernames with the right to resolve a task based on this template
- `allow_all_resolver_usernames`: boolean (default: false): when true, any user can execute a task based on this template
- `auto_runnable`; boolean (default: false): when true, the task will be executed directly after being created, IF the requester is an accepted resolver or `allow_all_resolver_usernames` is true
- `blocked`: boolean (default: false): no tasks can be created from this template
- `hidden`: boolean (default: false): the template is not listed on the API, it is concealed to regular users
- `retry_max`: int (default: 100): maximum amount of consecutive executions of a task based on this template, before being blocked for manual review 

### Inputs

When creating a new task, a requester needs to provide parameters described as a list of objects under the `inputs` property of a template. Additional parameters can be requested from a task's resolver user: those are represented under the `resolver_inputs` property of a template.

An input's definition allows to define validation constraints on the values provided for that input. See example template above.

#### Input properties

- `name`: unique name, used to access the value provided by the task's requester
- `description`: human readable description of the input, meant to give context to the task's requester
- `regex`: (optional) a regular expression that the provided value must match
- `legal_values`: (optional) a list of possible values accepted for this input
- `collection`: boolean (default: false) a list of values is accepted, instead of a single value
- `type`: (string|number|bool) (default: string) the type of data accepted
- `optional`: boolean (default: false) the input can be left empty
- `default`: (optional) a value assigned to the input if left empty

### Variables

A template variable is a named holder of either:
-  a fixed value
-  a JS expression evaluated on the fly. 

See the example template above to see variables in action. The expression in a variable can contain template handles to introduce values dynamically (from executed steps, for instance), like a step's configuration.

### Steps

A step is the smallest unit of work that can be performed within a task. At is's heart, a step defines an **action**: several types of actions are available, and each type requires a different configuration, provided as part of the step definition. The state of a step will change during a task's resolution process, and determine which steps become elligible for execution. Custom states can be defined for a step, to fine-tune execution flow (see below).

A sequence of ordered steps constitutes the entire workload of a task. Steps are ordered by declaring **dependencies** between each other. A step declares its dependencies as a list of step names on which it waits, meaning that a step's execution will be on hold until its dependencies have been resolved. A dependency can be qualified with a step's state. For example, `step2` can declare a dependency on `step1` in the following ways:
- `step1`: wait for `step1` to be in state `DONE`
- `step1:PRUNE`: wait for `step1` to be in state `PRUNE`
- `step1:ANY`: wait for `step1` to be in any "final" state, ie. it cannot keep running

The flow of this sequence can further be controlled with **conditions** on the steps: a condition is a clause that can be run before or after the step's action. A condition can either be used:
- to skip a step altogether
- to analyze its outcome and override the engine's default behaviour

TODO step.conditions examples

#### Basic Properties 

- `name`: a unique identifier
- `description`: a human readable sentence to convey the step's intent
- `dependencies`: a list of step names on which this step waits before running
- `retry_pattern`: (seconds|minutes|hours) define on what temporal order of magnitude the re-runs of this step should be spread 

#### Action

The `action` field of a step defines the actual workload to be performed. It consists of at least a `type` chosen among the registered action plugins, and a `configuration` fitting that plugin. See below for a detailed description of builtin plugins. For information on how to develop your own action plugins, refer to [this section](#plugins).

When an `action`'s configuration is repeated across several steps, it can be factored by defining `base_configurations` at the root of the template. For example:

```yaml
base_configurations:
  postMessage:
    method: POST
    url: http://message.board/new
``` 

This base configuration can then be leveraged by any step wanting to post a message, with different bodies:

```yaml
steps:
  sayHello:
    description: Say hello on the message board
    action:
      type: http
      base_configuration: postMessage
      configuration:
        body: Hello
  sayGoodbye:
    description: Say goodbye on the message board
    dependencies: [sayHello]
    action:
      type: http
      base_configuration: postMessage
      configuration:
        body: Goodbye
```

These two step definitions are the equivalent of: 

```yaml
steps:
  sayHello:
    description: Say hello on the message board
    action:
      type: http
      configuration:
        body: Hello
        method: POST
        url: http://message.board/new
  sayGoodbye:
    description: Say goodbye on the message board
    dependencies: [sayHello]
    action:
      type: http
      configuration:
        body: Goodbye
        method: POST
        url: http://message.board/new
```

The output of an action can be enriched by means of a `base_output`. For example, in a template with an input field named `id`, value `1234` and a call to a service which returns the following payload:

```js
{
  "name": "username"
}
```

The following action definition:

```yaml
steps:
  getUser:
    description: Prefix an ID received as input, return both
    action:
      type: http
      base_output:
        id: "{{.input.id}}"
      configuration:
        method: GET
        url: http://directory/user/{{.input.id}}
```

Will render the following output, a combination of the action's raw output and the base_output:

```js
{
  "id": "1234",
  "name": "username"
}
```

#### Builtin actions

Plugin name|Description|Configuration  
---|---|---
**`echo`** | Print out a pre-determined result | `output`: an object with the complete output of the step
           || `metadata`: an object containing the metadata returned by the step
           || `error_message`: for testing purposes, an error message to simulate execution failure
           || `error_type`: (client/server) for testing purposes: `client` error blocks execution, `server` lets the step be retried
**`http`** | Make an http request | `url`: destination for the http call, including host, path and query params
           || `method`: http method (GET/POST/PUT/DELETE)
           || `body`: a string representing the payload to be sent with the request
           || `headers`: a list of headers, represented as objects composed of `name` amd `value`
**`subtask`** | Spawn a new task on µTask | `template`: the name of a task template, as accepted through µTask's  API
              || `inputs`: a map of named values, as accepted on µTask's API
**`notify`**  | Dispatch a notification over a registered channel | `message`: the main payload of the notification
              || `fields`: a collection of extra fields to annotate the message
              || `backends`: a collection of the backends over which the message will be dispatched (values accepted: named backends as configured in [`utask-cfg`](./config/README.md))
**`apiovh`**  | Make a signed call on OVH's public API (requires credentials retrieved from configstore, containing the fields `endpoint`, `appKey`, `appSecret`, `consumerKey`, more info [here](https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/))| `path`: http route + query params
              || `method`: http method (GET/POST/PUT/DELETE)
              || `body`: a string representing the payload to be sent with the request
**`ssh`**     | Connect remotely to a system and run commands on it| TODO
 
#### Loops

A step can be configured to take a json-formatted collection as input, in its `foreach` property. It will be executed once for each element in the collection, and its result will be a collection of each iteration. This scheme makes it possible to chain several steps with the `foreach` property.

TODO provide an example

## Extending µTask with plugins <a name="plugins"></a>

TODO

### Action Plugins

TODO

### Init Plugins

TODO
 
## Contributing <a name="contributing"></a>
 
### Backend

In order to iterate on feature development, run the utask server plus a backing postgres DB by invoking `make run-test-stack-docker` in a terminal. Use SIGINT (`Ctrl+C`) to reboot the server, and SIGQUIT (`Ctrl+4`) to teardown the server and its DB.

In a separate terminal, rebuild (`make re`) each time you want to iterate on a code patch, then reboot the server in the terminal where it is running.

### Frontend

µTask serves two graphical interfaces: one for general use of the tool (`dashboard`), the other one for task template authoring (`editor`). They're found in the `ui` folder and each have their own Makefile for development purposes. 

Run `make dev` to launch a live-reloading on your machine. The editor is a standalone GUI, while the dashboard needs a backing µTask api (see above to run a server).

### Run the tests

Run all test suites against an ephemeral postgres DB:

```bash
$ make test-docker
```
	
### Get in touch

You've developed a new cool feature ? Fixed an annoying bug ? We'll be happy
to hear from you! Take a look at [CONTRIBUTING.md](https://github.com/ovh/utask/blob/master/CONTRIBUTING.md)

 
## Related links
 
 * Contribute: [CONTRIBUTING.md](CONTRIBUTING.md)
 * Report bugs: [https://github.com/ovh/utask/issues](https://github.com/ovh/utask/issues)
 * Get latest version: [https://github.com/ovh/utask/releases](https://github.com/ovh/utask/releases)
 * License: [LICENSE](LICENSE)
