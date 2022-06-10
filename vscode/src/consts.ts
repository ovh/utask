// URL of the json-schema validating µTask templates for the specific tag
export const TEMPLATE_SCHEMA_URL =
  "https://raw.githubusercontent.com/ovh/utask/_VERSION_/hack/template-schema.json";

// URL of the json-schema validating µTask functions for the specific tag
export const FUNCTION_SCHEMA_URL =
  "https://raw.githubusercontent.com/ovh/utask/_VERSION_/hack/function-schema.json";

// List of expressions identifying template folders
export const TEMPLATE_FILTERS = [
  /.+\/templates(-.+)?\/.+\.yaml$/,
];

// List of expressions identifying function folders
export const FUNCTION_FILTERS = [
  /.+\/functions(-.+)?\/.+\.yaml$/,
];
