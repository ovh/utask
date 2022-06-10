import exp = require("constants");
import * as vscode from "vscode";
import { TEMPLATE_FILTERS } from "./consts";

export function isUtaskTemplateDocument(document?: vscode.TextDocument): boolean {
  if (document?.languageId !== "yaml") {
    return false;
  }

  if (document.isUntitled) {
    return false;
  }

  for (let expr of TEMPLATE_FILTERS) {
    if (document.uri.fsPath.match(expr)) {
      return true;
    }
  }

  return false;
}
