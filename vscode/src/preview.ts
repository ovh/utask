import * as vscode from "vscode";
import * as nls from "vscode-nls";
import * as uri from "vscode-uri";

import { log } from "./log";
import { isUtaskTemplateDocument } from "./utils";

const localize = nls.loadMessageBundle();

export function init(context: vscode.ExtensionContext) {
  const preview = new UtaskPreview(context);

  context.subscriptions.push(
    vscode.commands.registerCommand("utask.preview.show", () => {
      if (
        isUtaskTemplateDocument(vscode.window.activeTextEditor?.document) &&
        vscode.window.activeTextEditor?.document.uri
      ) {
        preview.load(vscode.window.activeTextEditor?.document.uri);
      }
    })
  );

  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor(editor => {
      if (editor?.document && isUtaskTemplateDocument(editor.document)) {
        vscode.commands.executeCommand('setContext', 'isUtaskTemplate', true);
      } else {
        vscode.commands.executeCommand('setContext', 'isUtaskTemplate', false);
      }
    })
  );
}

class UtaskPreview extends vscode.Disposable {
  private static viewType = "utask.preview";

  private _panel?: vscode.WebviewPanel;
  private _resource?: vscode.Uri;

  constructor(private _context: vscode.ExtensionContext) {
    super(() => {
      this.dispose();
    });

    _context.subscriptions.push(
      vscode.window.onDidChangeActiveTextEditor(editor => {
        if (this._panel && editor && isUtaskTemplateDocument(editor.document)) {
          this.load(editor.document.uri);
        }
      })
    );

    _context.subscriptions.push(
      vscode.workspace.onDidSaveTextDocument(document => {
        if (document.uri === this._resource) {
          this.refresh();
        }
      })
    );
  }

  public load(resource: vscode.Uri) {
    this._resource = resource;

    if (!this._panel) {
      this._panel = vscode.window.createWebviewPanel(
        UtaskPreview.viewType,
        "µTask Preview",
        vscode.ViewColumn.Two,
        {
          enableScripts: true,
          localResourceRoots: [
            vscode.Uri.joinPath(this._context.extensionUri, 'dist-web'),
          ]
        }
      );

      this._panel.onDidDispose(() => {
        this._panel = undefined;
      });

      this._panel.webview.onDidReceiveMessage((msg: { type: string; value?: any }) => {
        switch (msg.type) {
          case 'initialized':
            this.refresh();
            break;

          default:
            log(`Unknown message type: ${msg.type}`);
        }
      });

      const resourceLabel = uri.Utils.basename(this._resource);
      this._panel.title = localize(
        "utask.preview.title",
        "Preview {0}",
        resourceLabel
      );
      this._panel.webview.html = this.getHtmlContent();
    } else {
      this.refresh();
    }
  }

  public refresh() {
    if (this._panel && this._resource) {
      vscode.workspace.openTextDocument(this._resource).then(document => {
        this._panel?.webview.postMessage({
          type: 'refresh',
          value: document.getText(),
        });
      });
    }
  }

  private getHtmlContent() {
    if (!this._panel) {
      return '';
    }

    const stylesUri = this._panel.webview.asWebviewUri(
      vscode.Uri.joinPath(this._context.extensionUri, "dist-web", "styles.css")
    );

    const scriptRuntimeUri = this._panel.webview.asWebviewUri(
      vscode.Uri.joinPath(this._context.extensionUri, "dist-web", "runtime.js")
    );

    const scriptPolyfillsUri = this._panel.webview.asWebviewUri(
      vscode.Uri.joinPath(this._context.extensionUri, "dist-web", "polyfills.js")
    );

    const scriptVendorUri = this._panel.webview.asWebviewUri(
      vscode.Uri.joinPath(this._context.extensionUri, "dist-web",
        "vendor.js")
    );

    const scriptMainUri = this._panel.webview.asWebviewUri(
      vscode.Uri.joinPath(this._context.extensionUri, "dist-web",
        "main.js")
    );

    const baseUri = this._panel.webview.asWebviewUri(vscode.Uri.joinPath(
      this._context.extensionUri, 'dist-web')
    ).toString().replace('%22', '');

    return `<!doctype html>
      <html lang="en">
      <head>
        <meta charset="utf-8">
        <title>Utask.Preview</title>
        <base href="${baseUri}/">
        <meta name="viewport" content="width=device-width, initial-scale=1">
        <link rel="stylesheet" href="${stylesUri}"></head>
      <body>
        <app-root></app-root>
        <script src="${scriptRuntimeUri}" type="module"></script>
        <script src="${scriptPolyfillsUri}" type="module"></script>
        <script src="${scriptVendorUri}" type="module"></script>
        <script src="${scriptMainUri}" type="module"></script>
      </body>
      </html>`;
  }
}
