import * as vscode from "vscode";
import { FUNCTION_SCHEMA_URL, TEMPLATE_SCHEMA_URL } from "./consts";
import { init as initLogger, log } from "./log";
import { init as initPreview } from "./preview";

export function activate(context: vscode.ExtensionContext) {
  initLogger();

  const yaml = vscode.workspace.getConfiguration("yaml");
  const yamlSchemas = yaml.schemas || {};
  const schemaSettings = vscode.workspace.getConfiguration("utask.schema");
  const schemaRegex = new RegExp('^https://raw\.githubusercontent\.com/ovh/utask/(master|v[0-9]+\.[0-9]+\.[0-9]+)/hack/(function|template)-schema\.json$');

  // Compute version
  const schemaVersion = (schemaSettings.version === undefined || schemaSettings.version === 'latest') ? 'master' : `v${schemaSettings.version}.0`;
  const functionSchema = FUNCTION_SCHEMA_URL.replace('_VERSION_', schemaVersion);
  const templateSchema = TEMPLATE_SCHEMA_URL.replace('_VERSION_', schemaVersion);

  // Target settings
  const newYamlSchemas = Object.assign(Object.keys(yamlSchemas).reduce((result, path) => {
    if (schemaRegex.exec(path) && path !== functionSchema && path !== templateSchema) {
      return result;
    }
    return Object.assign(result, {
      [path]: yamlSchemas[path]
    })
  }, {}), {
    [functionSchema]: ["**/functions/*.yaml", "**/functions-*/*.yaml"],
    [templateSchema]: ["**/templates/*.yaml", "**/templates-*/*.yaml"]
  });

  yaml.update('schemas', newYamlSchemas, true).then(() => {
    log(`YAML schema updated for version ${schemaVersion}`);
  });

  initPreview(context);
}
