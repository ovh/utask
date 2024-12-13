# `batch` Plugin

This plugin creates a batch of tasks based on the same template and waits for it to complete. It acts like the `subtask` combined with a `foreach`, but doesn't modify the resolution by adding new steps dynamically. As it makes less calls to the underlying database, this plugin is suited for large batches of tasks, where the `subtask` / `foreach` combination would usually struggle, especially by bloating the database.
Tasks belonging to the same batch share a common `BatchID` as well as a tag holding their parent's ID.

##### Remarks:
Like the subtask plugin, it's unadvised to have a step based on the batch plugin running alongside other steps in a template. If these other steps take time to return a result, the batch plugin may miss the wake up call from its children tasks.
The output of child tasks is not made available in this plugin's output. This feature will come later.

## Configuration

| Fields               | Description                                                                                                       |
|----------------------|-------------------------------------------------------------------------------------------------------------------|
| `template_name`      | the name of a task template, as accepted through µTask's API                                                      |
| `inputs`             | a list of mapped key/value, as accepted on µTask's API. Each element represents the input of an individual task   |
| `json_inputs`        | same as `inputs`, but as a JSON string. If specified, it overrides `inputs`                                       |
| `common_inputs`       | a map of named values, as accepted on µTask's API, given to all task in the batch by combining it with each input |
| `common_json_inputs`  | same as `common_inputs` but as a JSON string. If specified, it overrides `common_inputs`                             |
| `tags`               | a map of named strings added as tags when creating child tasks                                                    |
| `sub_batch_size`     | the number tasks to create and run at once, as a string. `0` for infinity (i.e.: all tasks are created at once and waited for) (default). Higher values reduce the amount of calls made to the database, but increase sensitivity to database unavailability (if a task creation fails, the whole sub batch must be created again) |
| `comment`            | a string set as `comment` when creating child tasks                                                               |
| `resolver_usernames` | a string containing a JSON array of additional resolver users for child tasks                                     |
| `resolver_groups`    | a string containing a JSON array of additional resolver groups for child tasks                                    |
| `watcher_usernames`  | a string containing a JSON array of additional watcher users for child tasks                                      |
| `watcher_groups`     | a string containing a JSON array of additional watcher groups for child tasks                                     |

## Example

An action of type `batch` requires the following kind of configuration:

```yaml
action:
  type: batch
  configuration:
    # [Required]
    # A template that must already be registered on this instance of µTask
    template_name: some-task-template
    # Valid inputs, as defined by the referred template, here requiring 3 inputs: foo, otherFoo and fooCommon
    inputs:
        - foo: bar-1
          otherFoo: otherBar-1
        - foo: bar-2
          otherFoo: otherBar-1
        - foo: bar-3
          otherFoo: otherBar-3
    # [Optional]
    common_inputs:
        fooCommon: barCommon
    # Some tags added to all child tasks
    tags:
        fooTag: value-of-foo-tag
        barTag: value-of-bar-tag
    # The amount of tasks to run at once
    sub_batch_size: "2"
    # A list of users which are authorized to resolve this specific task
    resolver_usernames: '["authorizedUser"]'
    resolver_groups: '["authorizedGroup"]'
    watcher_usernames: '["authorizedUser"]'
    watcher_groups: '["authorizedGroup"]'
```

## Requirements

None.

## Return

### Output

None.

### Metadata

| Name                 | Description                               |
|----------------------|-------------------------------------------|
| `batch_id`           | The public identifier of the batch        |
| `remaining_tasks`    | How many tasks still need to complete     |
| `tasks_started`       | How many tasks were started so far        |
