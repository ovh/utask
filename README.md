# µTask, the Lightweight Automation Engine

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

µTask allows you to model business processes in a **declarative yaml format**. Describe a set of inputs and a graph of actions and their inter-dependencies: µTask will asynchronously handle the execution of each action, working its way around transient errors and keeping an encrypted, auditable trace of all intermediary states until completion.

<img src="./assets/img/utask.png" width="50%" align="right">

## Table of contents

- [Real-world examples](#examples)
- [Quick Start](#quickstart)
- [Operating in production](#operating)
- [Configuration](#configuration)
- [Authoring Task Templates](#templates)
- [Extending µTask with plugins](#plugins)
- [Contributing](#contributing)

## Real-world examples <a name="examples"></a>

Here are a few real-world examples that can be implemented with µTask:

### Kubernetes ingress TLS certificate provisioning

A new ingress is created on the production kubernetes cluster. A hook triggers a µTask template that:
- generates a private key
- requests a new certificate
- meets the certificate issuer's challenges
- commits the resulting certificate back to the cluster

### New team member bootstrap

A new member joins the team. The team leader starts a task specifying the new member's name, that:
- asks the new team member to generate an SSH key pair and copy the public key in a µTask-generated form
- registers the public SSH key centrally
- creates accounts on internal services (code repository, CI/CD, internal PaaS, ...) for the new team member
- triggers another task to spawn a development VM
- sends a welcome email full of GIFs

### Payments API asynchronous processing

The payments API receives a request that requires an asynchronous antifraud check. It spawns a task on its companion µTask instance that:
- calls a first risk-assessing API which returns a number
- if the risk is low, the task succeeds immediately
- otherwise it calls a SaaS antifraud solution API which returns a score
- if the score is good, the task succeeds
- if the score is very bad, the task fails
- if it is in between, it triggers a human investigation step where an operator can enter a score in a µTask-generated form
- when it is done, the task sends an event to the payments API to notify of the result

The payments API keeps a reference to the running workflow via its task ID. Operators of the payments API can follow the state of current tasks by requesting the µTask instance directly. Depending on the payments API implementation, it may allow its callers to follow a task's state.

## Quick start <a name="quickstart"></a>

### Running with docker-compose

Download our latest install script, setup your environment and launch your own local instance of µTask. 

```bash
mkdir utask && cd utask
wget https://github.com/ovh/utask/releases/latest/download/install-utask.sh
sh install-utask.sh
docker-compose up
```

All the configuration for the application is found in the environment variables in docker-compose.yaml. You'll see that basic auth is setup for user `admin` with password `1234`. Try logging in with this user on the graphical dashboard: [http://localhost:8081/ui/dashboard](http://localhost:8081/ui/dashboard).

You can also explore the API schema: [http://localhost:8081/unsecured/spec.json](http://localhost:8081/unsecured/spec.json).

Request a new task:
![](./assets/img/utask_new_task.png)

Get an overview of all tasks:
![](./assets/img/utask_dashboard.png)

Get a detailed view of a running task:
![](./assets/img/utask_running.png)

Browse available task templates:
![](./assets/img/utask_templates.png)

### Running with your own postgres service

Alternatively, you can clone this repository and build the µTask binary:

```bash
make all
```

## Operating in production <a name="operating"></a>

The folder you created in the previous step is meant to become a git repo where you version your own task templates and plugins. Re-download and run the latest install script to bump your version of µTask.

You'll deploy your version of µTask by building a docker image based on the official µTask image, which will include your extensions. See the Dockerfile generated during installation.

### Architecture

µTask is designed to run a task scheduler and perform the task workloads within a single runtime: work is not delegated to external agents. Multiple instances of the application will coordinate around a single postgres database: each will be able to determine independently which tasks are available. When an instance of µTask decides to execute a task, it will take hold of that task to avoid collisions, then release it at the end of an execution cycle. 

A task will keep running as long as its steps are successfully executed. If a task's execution is interrupted before completion, it will become available to be re-collected by one of the active instances of µTask. That means that execution might start in one instance and resume on a different one.

### Maintenance procedures

#### Key rotation

1. Generate a new key with [symmecrypt](https://github.com/ovh/symmecrypt), with the 'storage' label.
2. Add it to your configuration items. The library will take all keys into account and use the latest possible key, falling back to older keys when finding older data.
3. Set your API in maintenance mode (env var or command line arg, see config below): all write actions will be refused when you reboot the API.
4. Reboot API.
5. Make a POST request on the /key-rotate endpoint of the API.
6. All data will be encrypted with the latest key, you can delete older keys.
7. De-activate maintenance mode.
8. Reboot API.

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

Extending this basic authentication mechanism is possible by developing an "init" plugin, as described [below](#plugins).

## Authoring Task Templates <a name="templates"></a>

Checkout the [µTask examples directory](./examples).

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
- `.step.[STEP_NAME].children`: the collection of results from a 'foreach' step
- `.step.[STEP_NAME].error`: error message from a failed step
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

Several conditions can be specified, the first one to evaluate as `true` is applied. A condition is composed of:
- a `type` (skip or check) 
- a list of `if` assertions (`value`, `operator`, `expected`) which all have to be true (AND on the collection), 
- a `then` object to impact the state of steps (`this` refers to the current step)
- an optional `message` to convey the intention of the condition, making it easier to inspect tasks

Here's an example of a `skip` condition. The value of an input is evaluated to determine the result: if the value of `runType` is `dry`, the `createUser` step will not be executed, its state will be set directly to DONE.
```yaml
inputs:
- name: runType
  description: Run this task with/without side effects
  legal_values: [dry, wet]
steps:
  createUser:
    description: Create new user
    action:
      ... etc...
    conditions:
    - type: skip
      if:
      - value: '{{.input.runType}}'
        operator: EQ
        expected: dry
      then:
        this: DONE
      message: Dry run, skip user creation
```

Here's an example of a `check` condition. Here the return of an http call is inspected: a 404 status will put the step in a custom NOT_FOUND state. The default behavior would be to consider any 4xx status as a client error, which blocks execution of the task. The check condition allows you to consider this situation as normal, and proceed with other steps that take the NOT_FOUND state into account (creating the missing resource, for instance).

```yaml
steps:
  getUser:
    description: Get user
    custom_states: [NOT_FOUND]
    action: 
      type: http
      configuration:
        url: http://example.org/user/{{.input.id}}
        method: GET
    conditions:
    - type: check
      if:
      - value: '{{.step.getUser.metadata.HTTPStatus}}'
        operator: EQ
        expected: '404'
      then:
        this: NOT_FOUND
      message: User {{.input.id}} not found
```

#### Basic Step Properties 

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

Browse [builtin actions](./pkg/plugins/builtin)

|Plugin name|Description|Link  
|---|---|---
|**`echo`** | Print out a pre-determined result | [Access plugin doc](./pkg/plugins/builtin/echo/README.md)
|**`http`** | Make an http request | [Access plugin doc](./pkg/plugins/builtin/http/README.md)
|**`subtask`** | Spawn a new task on µTask | [Access plugin doc](./pkg/plugins/builtin/subtask/README.md)
|**`notify`**  | Dispatch a notification over a registered channel | [Access plugin doc](./pkg/plugins/builtin/notify/README.md)
|**`apiovh`**  | Make a signed call on OVH's public API (requires credentials retrieved from configstore, containing the fields `endpoint`, `appKey`, `appSecret`, `consumerKey`, more info [here](https://docs.ovh.com/gb/en/customer/first-steps-with-ovh-api/)) | [Access plugin doc](./pkg/plugins/builtin/apiovh/README.md)
|**`ssh`**     | Connect to a remote system and run commands on it | [Access plugin doc](./pkg/plugins/builtin/ssh/README.md)
|**`email`**   | Send an email | [Access plugin doc](./pkg/plugins/builtin/email/README.md)
|**`ping`**    | Send a ping to an hostname *Warn: This plugin will keep running until the count is done* | [Access plugin doc](./pkg/plugins/builtin/ping/README.md)

 
#### Loops

A step can be configured to take a json-formatted collection as input, in its `foreach` property. It will be executed once for each element in the collection, and its result will be a collection of each iteration. This scheme makes it possible to chain several steps with the `foreach` property.

For the following step definition (note json-format of `foreach`):
```yaml
steps:
  prefixStrings:
    description: Process a collection of strings, adding a prefix
    foreach: '[{"id":"a"},{"id":"b"},{"id":"c"}]'  
    action:
      type: echo
      configuration:
        output:
          prefixed: pre-{{.iterator.id}}
```

The following output can be expected to be accessible at `{{.step.prefixStrings.children}}`
```js
[{
  "prefixed": "pre-a"
},{
  "prefixed": "pre-b"
},{
  "prefixed": "pre-c"
}]
```

This output can be then passed to another step in json format:
```yaml
foreach: '{{.step.prefixStrings.children | jsonmarshal}}'
```

## Extending µTask with plugins <a name="plugins"></a>

µTask is extensible with [golang plugins](https://golang.org/pkg/plugin/) compiled in *.so format. Two kinds of plugins exist:
- action plugins, that you can re-use in your task templates to implement steps
- init plugins, a way to customize the authentication mechanism of the API, and to draw data from different providers of the configstore library

The installation script for utask creates a folder structure that will automatically package and build your code in a docker image, with your plugins ready to be loaded by the main binary at boot time. Create a separate folder for each of your plugins, within either the `plugins` or the `init` folders.

### Action Plugins

Action plugins allow you to extend the kind of work that can be performed during a task. An action plugin has a name, that will be referred to as the action `type` in a template. It declares a configuration structure, a validation function for the data received from the template as configuration, and an execution function which performs an action based on valid configuration.

Create a new folder within the `plugins` folder of your utask repo. There, develop a `main` package that exposes a `Plugin` variable that implements the `TaskPlugin` defined in the `plugins` package: 

```golang
type TaskPlugin interface {
	ValidConfig(baseConfig json.RawMessage, config json.RawMessage) error
	Exec(stepName string, baseConfig json.RawMessage, config json.RawMessage, ctx interface{}) (interface{}, interface{}, error)
	Context(stepName string) interface{}
	PluginName() string
	PluginVersion() string
	MetadataSchema() json.RawMessage
}
```

The `taskplugin` [package](./pkg/plugins/taskplugin/taskplugin.go) provides helper functions to build a Plugin: 

```golang
package main

import (
	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

var (
	Plugin = taskplugin.New("my-plugin", "v0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}))
)

type Config struct { ... }

func validConfig(config interface{}) (err error) {
  cfg := config.(*Config)
  ...
  return
}

func exec(stepName string, config interface{}, ctx interface{}) (output interface{}, metadata interface{}, err error) {
  cfg := config.(*Config)
  ...
  return
}
```

### Init Plugins

Init plugins allow you to customize your instance of µtask by giving you access to its underlying configuration store and its API server. 

Create a new folder within the `init` folder of your utask repo. There, develop a `main` package that exposes a `Plugin` variable that implements the `InitializerPlugin` defined in the `plugins` package: 

```golang
type Service struct {
	Store  *configstore.Store
	Server *api.Server
}

type InitializerPlugin interface {
	Init(service *Service) error // access configstore and server to customize µTask
	Description() string         // describe what the initialization plugin does
}
```

As of version `v1.0.0`, this is meant to give you access to two features:
- `service.Store` exposes the `RegisterProvider(name string, f configstore.Provider)` method that allow you to plug different data sources for you configuration, which are not available by default in the main runtime
- `service.Server` exposes the `WithAuth(authProvider func(*http.Request) (string, error))` method, where you can provide a custom source of authentication and authorization based on the incoming http requests

If you develop more than one initialization plugin, they will all be loaded in alphabetical order. You might want to provide a default initialization, plus more specific behaviour under certain scenarios.

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
