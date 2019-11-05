# µTask, the Lighweight Automation Engine

[![Go Report Card](https://goreportcard.com/badge/github.com/ovh/utask)](https://goreportcard.com/report/github.com/ovh/utask)
[![Maintenance](https://img.shields.io/maintenance/yes/2019.svg)]()
![GitHub last commit](https://img.shields.io/github/last-commit/ovh/utask)
[![GoDoc](https://godoc.org/github.com/ovh/utask?status.svg)](https://godoc.org/github.com/ovh/utask)
![GitHub stars](https://img.shields.io/github/stars/ovh/utask?style=social)
<!-- [![Coverage Status](https://coveralls.io/repos/github/ovh/utask/badge.svg?branch=master)](https://coveralls.io/github/ovh/utask?branch=master) -->
<!-- [![Github All Releases](https://img.shields.io/github/downloads/ovh/utask/total.svg)](https://github.com/ovh/utask/releases) -->
<!-- [![Release](https://badge.fury.io/gh/ovh/utask.svg)](https://github.com/ovh/utask/releases) -->
 
µTask is an automation engine built for the cloud. It is:
- simple to operate: only a postgres DB is required
- secure: all data is encrypted, only visible to authorized users
- extensible: you can develop custom actions in golang

µTask allows you to model business processes in a **declarative yaml format**. Describe a set of inputs and a graph of actions and their inter-dependencies: µTask will asynchronously handle the execution of each action, working its way around transient errors and keeping a trace of all intermediary states until completion.

## Contents

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

### Example

Here's a (contrived) example of a task template, showcasing many of its capabilities. A description of each property is provided below. 

The `hello-world-now` template takes in a `language` input parameter, which admits two possible values, and adopts its default value if no input is provided. The first step of the task is an external API call to retrieve the current UTC time. A second step waits for completion of the first step, then prints out a message conditioned by the input parameter. A final result is built from the output of both steps and shown to the task's requester.

```yaml
name:             hello-world-now
description:      Say hello to the world, now!
long_description: This task prints out a greeting to the entire world, after retrieving the current UTC time from an external API
doc_link:         https://en.wikipedia.org/wiki/%22Hello,_World!%22_program

title_format:     Say hello in {{.input.language}}
result_format:
  echo_message: '{{.step.sayHello.output.message}}'
  echo_when:    '{{.step.sayHello.output.when}}'

allowed_resolver_usernames:   []
allow_all_resolver_usernames: true
auto_runnable: true
blocked:       false
hidden:        false

variables:
- name: english-message
  value: Hello World!
- name: spanish-message
  expression: |-
    // a short javascript snippet
    var h = 'Hola';
    var m = 'mundo';
    h + ' ' + m + '!';

inputs:
- name: language
  description: The language in which you wish to greet the world
  legal_values: [english, spanish]
  optional: true
  default: english

steps:
  getTime:
    description: Get UTC time
    action:
      type: http
      configuration:
        url: http://worldclockapi.com/api/json/utc/now
        method: GET
  sayHello:
    description: Echo a greeting in your language of choice
    dependencies: [getTime]
    action:
      type: echo
      configuration:
        output:
          message: >-
            {{if (eq .input.language `english`)}}{{eval `english-message`}}
            {{else if (eq .input.language `spanish`)}}{{eval `spanish-message`}}{{end}}
          when: '{{.step.getTime.output.currentDateTime}}'
```

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

#### Basic Step Properties 

- `name`: a unique identifier
- `description`: a human readable sentence to convey the step's intent
- `action`: the step's workload, defined by a `type` and a `configuration` (see Builtin actions below)
- `dependencies`: a list of step names on which this step waits before running

#### Advanced Step Properties 

- `retry_pattern`: (seconds|minutes|hours) define on what temporal order of magnitude the re-runs of this step should be spread


### Builtin actions

#### echo

#### http

#### ssh

#### subtask

#### notify

#### apiovh
 
## Extending µTask with plugins <a name="plugins"></a>

### Action Plugins

### Init Plugins
 
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
 
 * Contribute: [https://github.com/ovh/utask/blob/master/CONTRIBUTING.md](https://github.com/ovh/utask/blob/master/CONTRIBUTING.md)
 * Report bugs: [https://github.com/ovh/utask/issues](https://github.com/ovh/utask/issues)
 * Get latest version: [https://github.com/ovh/utask/releases](https://github.com/ovh/utask/releases)
 * License: [https://github.com/ovh/utask/blob/master/LICENSE](https://github.com/ovh/utask/blob/master/LICENSE)