# CHANGELOG

A detailed changelog is available in our [GitHub release page](https://github.com/ovh/utask/releases "GitHub release page").

We will list in this document any breaking changes between versions, that require modifications of the task templates you maintain.

## Breaking changes

### v1.5.0 (2020-04-09)
##### Templating
- `jsonmarshal` function has been __deleted__. It was obsolete since the `sprig` package used by utask provides many JSON related functions (e.g. `toJson`).
- `jsonfield` function has been __deleted__. It was obsolete since a pipeline such as ``{{ field `foo` | toJson }}`` achieves the same thing.

##### API model
- ``/resolution``: `resolution.payload` (deprecated, duplicate of / replaced by `resolution.output` since 1.0) has finally been __deleted__.

### v1.4.0 (2020-03-31)
##### API routes
- `GET /resolution` API route has been __deleted__. We decided that resolutions should always be accessed through tasks

##### Configuration
- `resolver_usernames` configuration field __deleted__. Global resolver is not a feature we want to support, and its usage was misguided.

##### SQL
- `001_resolver_watcher_usernames_indexes.sql` migration file should be applied while upgrading. This migration is non-breaking but will boost performance on big instances for tasks listing.

### v1.3.0 (2020-03-17)
##### HTTP plugin
- `timeout_seconds` configuration field __deleted__. It has been replaced by the field `timeout`, using the Golang time.Duration format (5s, 1m, ...), for consistency with other plugins and to express timeout durations inferior than 1 second. #86
- `parameters` configuration field __deleted__. It has been replaced by the field `query_parameters`: the name is now clearer, and the object format is similar to the `headers` configuration field. #86
- `deny_redirects` configuration field __deleted__. It has been replaced by the field `follow_redirect` : the new default behavior will be to never follow redirections, unless specified by this configuration field. #86

##### SSH plugin
- `allow_exit_non_zero` configuration field __deleted__. This field has no strict replacement, as its behavior was not in the uTask philosophy. The field `exit_codes_unrecoverable` has been introduced to catch some exit codes as `CLIENT_ERROR` if the error is not recoverable (to either halt the execution or change to a custom status via check conditions). #85
- `exit_status` metadata field __renamed__ `exit_code`. This field's name is now consistent between the ssh and script plugins. #85

##### script plugin
- `allow_exit_non_zero` configuration field __deleted__. This field has no strict replacement, as its behavior was not in the uTask philosophy. The field `exit_codes_unrecoverable` has been introduced to catch some exit codes as `CLIENT_ERROR` if the error is not recoverable (to either halt the execution or change to a custom status via check conditions). #87
- `last_line_not_json` configuration field __deleted__. It has been replaced by the field `output_mode` which supports more options, and can be configured to the value `disabled` to reproduce the same behavior. #87
