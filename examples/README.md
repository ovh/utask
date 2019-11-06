# Examples

## Template

You can find template examples under [template directory](template/).

To begin, we introduce the [hello-world-now.yaml](template/hello-world-now.yaml). Here's a (contrived) example of a task template, showcasing many of its capabilities. A description of each property is provided below.

The `hello-world-now` template takes in a `language` input parameter, which admits two possible values, and adopts its default value if no input is provided. The first step of the task is an external API call to retrieve the current UTC time. A second step waits for completion of the first step, then prints out a message conditioned by the input parameter. A final result is built from the output of both steps and shown to the task's requester.

## Plugin

🚧 Under construction 🚧