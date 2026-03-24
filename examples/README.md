# µTask Examples

This directory contains example templates and plugins to help you get started with µTask.

## Table of Contents

- [Getting Started](#getting-started)
- [Template Examples](#template-examples)
  - [Hello World Now](#hello-world-now)
- [Plugin Examples](#plugin-examples)
- [How to Use These Examples](#how-to-use-these-examples)
- [Creating Your Own Templates](#creating-your-own-templates)

## Getting Started

If you're new to µTask, we recommend starting with the [hello-world-now](templates/hello-world-now.yaml) template. It demonstrates:

- Basic template structure
- Input parameters with validation
- Variable definition (both static and JavaScript expressions)
- HTTP API calls
- Step dependencies
- Conditional output formatting
- JSON schema validation

## Template Examples

### Hello World Now

**File:** [templates/hello-world-now.yaml](templates/hello-world-now.yaml)

A introductory template that greets the world with the current UTC time.

#### What it does

1. Takes a `language` input parameter (english or spanish)
2. Calls an external API to get the current UTC time
3. Generates a greeting message based on the selected language
4. Returns the greeting along with the timestamp

#### Key Concepts Demonstrated

| Feature | Implementation |
|---------|---------------|
| **Input with default value** | `language` parameter defaults to "english" |
| **Legal values validation** | Only "english" and "spanish" are accepted |
| **Static variables** | `english-message` defined with fixed value |
| **JavaScript expressions** | `spanish-message` computed with JS |
| **HTTP action** | `getTime` step fetches from worldclockapi.com |
| **Step dependencies** | `sayHello` waits for `getTime` to complete |
| **Conditional templating** | Message output depends on input language |
| **JSON schema validation** | Validates API response structure |
| **Idempotent steps** | `getTime` marked as safe to retry |
| **Resource limiting** | Uses "worldclockapi" resource for rate limiting |

#### Running the Example

```bash
# Start your µTask instance
docker-compose up

# Access the dashboard at http://localhost:8081/ui/dashboard
# Login with admin / 1234

# Create a new task with the "hello-world-now" template
# Try changing the language input between "english" and "spanish"
```

#### Expected Output

When run with `language=english`:
```json
{
  "echo_message": "Hello World!",
  "echo_when": "2024-01-15T10:30Z"
}
```

When run with `language=spanish`:
```json
{
  "echo_message": "Hola mundo!",
  "echo_when": "2024-01-15T10:30Z"
}
```

## Plugin Examples

### Init Plugin

**Directory:** [plugins/init](plugins/init/)

An example initialization plugin that demonstrates how to configure custom authentication when µTask boots up.

#### What it does

- Initializes at µTask startup
- Configures a custom authentication provider
- Accesses the configstore for configuration data

#### Key Concepts Demonstrated

- Implementing the `InitializerPlugin` interface
- Accessing the µTask server for customization
- Registering custom authentication providers

#### Files

- `main.go` - Plugin implementation

## How to Use These Examples

### 1. Copy and Modify

The easiest way to create your own templates is to copy an example and modify it:

```bash
# Copy the hello world example
cp examples/templates/hello-world-now.yaml my-template.yaml

# Edit to match your needs
nano my-template.yaml
```

### 2. Test in Dashboard

1. Place your template in the `templates` directory
2. Restart µTask (or wait for auto-reload)
3. Open the dashboard and create a task with your template
4. Inspect the execution flow and outputs

### 3. Validate Your Template

Use the JSON schema for validation in your IDE:

- Schema location: `hack/template-schema.json`
- VSCode extension: `ovh.vscode-utask`

## Creating Your Own Templates

### Basic Structure

```yaml
name: my-template                    # Unique identifier
description: Short description       # One-line summary
long_description: |                  # Detailed documentation
  Multi-line description of what
  this template does and when to use it.

title_format: "Task for {{.input.name}}"  # Dynamic task title

inputs:                              # Input parameters
  - name: name
    description: Name to process
    type: string

steps:                               # Execution steps
  step1:
    description: First step
    action:
      type: echo
      configuration:
        output: "Hello {{.input.name}}"
```

### Best Practices

1. **Always include `description` and `long_description`** - Helps users understand your template
2. **Use meaningful step names** - Makes debugging and logs easier to read
3. **Mark idempotent steps** - Set `idempotent: true` for steps that can safely be retried
4. **Define custom states when needed** - For handling expected non-success outcomes
5. **Add JSON schema validation** - Catches data issues early
6. **Document expected inputs/outputs** - Use comments in your YAML

### Common Patterns

#### Pattern 1: HTTP API Call with Error Handling

```yaml
steps:
  apiCall:
    description: Call external API
    custom_states: [NOT_FOUND]
    action:
      type: http
      configuration:
        url: "https://api.example.com/resource/{{.input.id}}"
        method: GET
    conditions:
      - type: check
        if:
          - value: '{{.step.apiCall.metadata.HTTPStatus}}'
            operator: EQ
            expected: "404"
        then:
          this: NOT_FOUND
        message: Resource not found
```

#### Pattern 2: Conditional Step Execution

```yaml
steps:
  checkExists:
    # ... step definition ...

  createResource:
    dependencies: ["checkExists:NOT_FOUND"]
    description: Create if doesn't exist
    action:
      type: http
      configuration:
        url: "https://api.example.com/resource"
        method: POST
```

#### Pattern 3: Loop Over Collection

```yaml
steps:
  processItems:
    description: Process each item
    foreach: '[{"id":"1"},{"id":"2"},{"id":"3"}]'
    action:
      type: echo
      configuration:
        output:
          processed: "Item {{.iterator.id}}"
```

## Need Help?

- **Documentation**: See the main [README.md](../README.md)
- **Contributing**: See [CONTRIBUTING.md](../CONTRIBUTING.md)
- **Issues**: Report bugs at [GitHub Issues](https://github.com/ovh/utask/issues)

---

*These examples are maintained by the µTask community. Contributions are welcome!*
