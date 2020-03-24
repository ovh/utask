import { Component, OnInit } from '@angular/core';
import * as _ from 'lodash';

import * as brace from 'brace';
import 'brace/mode/yaml';
import 'brace/theme/monokai';

import JSToYaml from 'convert-yaml';
import { TemplateYamlHelper } from '../services/templateyamlhelper.service';
import jsYaml from 'js-yaml';

import StepsConfig from '../steps.config';
import Editor from '../models/editor.model';
import stepsConfig from '../steps.config';

const Range = brace.acequire('ace/range').Range;
require('brace/ext/language_tools');
const langTools = brace.acequire('ace/ext/language_tools');

@Component({
  templateUrl: 'editor.html',
})
export class EditorComponent implements OnInit {
  editor: Editor;
  steps: any[] = [];
  objectKeys = Object.keys;
  types: any = StepsConfig.types;

  constructor(private TemplateYamlHelper: TemplateYamlHelper) {
    JSToYaml.spacing = ' '.repeat(4);
    JSToYaml.spacingStart = '';
    this.editor = {
      valid: false,
      text: JSToYaml.stringify(StepsConfig.initValue).value,
      value: null,
      ace: null,
      type: 'yaml',
      mode: 'ace/mode/yaml',
      theme: 'ace/theme/monokai',
      minimumSpacing: 0,
      spacing: 4,
      markerIds: [],
      selectedMarkerIds: [],
      changedDelay: +new Date(),
      snippets: [],
      error: null
    };
  }

  ngOnInit() {
    this.initEditor();
    this.tryConvertToObject();
  }

  selectStep(stepName: string) {
    if (!stepName) {
      this.clearSelectedMarkers();
      return;
    }

    const stepPosition = this.TemplateYamlHelper.getStepRow(this.editor.text, stepName);
    const stepEndPosition = this.TemplateYamlHelper.getEndStep(this.editor.text, stepPosition);

    this.editor.ace.clearSelection();
    this.editor.ace.gotoLine(stepPosition, 0, true);
    const range = new Range(stepPosition, 0, stepEndPosition, 0);
    const markerId: number = this.editor.ace.session.addMarker(range, 'ace_active-selected', 'row', true);
    this.clearSelectedMarkers();
    this.editor.selectedMarkerIds.push(markerId);
  }

  initEditor() {
    this.editor.ace = brace.edit('ace-editor');
    this.editor.ace.getSession().setMode(this.editor.mode);

    this.editor.ace.setOptions({
      enableBasicAutocompletion: true,
      enableSnippets: true,
      enableLiveAutocompletion: false,
      tabSize: 2
    });

    langTools.setCompleters([]);
    langTools.addCompleter({
      getCompletions: (editor, session, pos, prefix, callback) => {
        const path = this.TemplateYamlHelper.getPath(this.editor.text, pos.row).join('.');
        let arr = [];

        const currentText = this.editor.text.split('\n')[pos.row];
        const isEmptyRow = currentText.trim() === '';
        const countSpacing = currentText.match(/^\s*/)[0].length;
        if (path === 'steps' || countSpacing <= (this.editor.minimumSpacing + this.editor.spacing)) {
          Object.keys(stepsConfig.types).forEach((key: string) => {
            JSToYaml.spacingStart = ' '.repeat(_.max([this.editor.minimumSpacing + this.editor.spacing - countSpacing, 0]));
            const snippet = {
              name: `Add '${key}' step`,
              score: 499,
              meta: 'Snippet',
              snippet: `${isEmptyRow ? '' : '\n'}${JSToYaml.stringify({ '${1:step_name}': stepsConfig.types[key].snippet }).value}\n`
            };
            JSToYaml.spacingStart = ' '.repeat(this.editor.minimumSpacing);
            arr.push(snippet);
          });
        } else if (path.match(/^steps\.[a-zA-Z0-9\-\_\s]+\.dependencies$/)) {
          const stepsName = this.TemplateYamlHelper.getStepsName(this.editor.text);
          stepsName.map((key: string) => {
            arr.push({ name: key, value: key, score: 500, meta: 'Step' });
          });
        }

        callback(null, arr);
      }
    });

    this.editor.ace.setTheme(this.editor.theme);
    this.editor.ace.setValue(this.editor.text);
    this.editor.ace.clearSelection();
    this.editor.ace.getSession().on('change', () => {
      this.editor.changedDelay = +new Date();
      const lastChanged = this.editor.changedDelay;
      setTimeout(() => {
        if (lastChanged === this.editor.changedDelay) {
          this.editor.text = this.editor.ace.getValue();
          this.editor.minimumSpacing = this.getMinimumSpacing(this.editor.text);
          this.editor.spacing = this.getSpacing(this.editor.text, this.editor.minimumSpacing);
          this.tryConvertToObject();
        }
      }, 250);
    });
  }

  askImport() {
    const element = document.createElement('input');
    element.setAttribute('type', 'file');
    element.setAttribute(
      'href',
      `data:application/json;charset=utf-8,${encodeURIComponent(this.editor.ace.getValue())}`
    );
    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    element.onchange = (event) => {
      const file = _.get(event, 'target.files[0]');
      if (file) {
        const reader = new FileReader();
        reader.readAsText(file, 'UTF-8');
        reader.onload = (e: any) => {
          this.editor.text = e.target.result;
          this.initEditor();
        };
        reader.onerror = (err: any) => {
          console.log('ERROR: ', err);
        };
      }
    };
    document.body.removeChild(element);
  }

  download() {
    const element = document.createElement('a');
    element.setAttribute(
      'href',
      `data:application/json;charset=utf-8,${encodeURIComponent(this.editor.ace.getValue())}`
    );
    element.setAttribute('download', `utask-template-${+new Date()}.yaml`);
    element.style.display = 'none';
    document.body.appendChild(element);
    element.click();
    document.body.removeChild(element);
  }

  getMinimumSpacing(text: string) {
    const min = _.min(text.split('\n').map((line: string) => {
      if (line.trim() === '') {
        return null;
      }

      const isEndOrStartFile = line.match(/^\s*(\.{3}|\-{3})/);
      if (isEndOrStartFile) {
        return null;
      }

      const isComment = line.match(/^\s*#/);
      if (isComment) {
        return null;
      }

      const match = line.match(/^\s*/);
      if (match) {
        return match[0].length;
      }

      return 0;
    })) || 0;
    return min;
  }

  getSpacing(text: string, minSpacing: number): number {
    const spacing = _.min(text.split('\n').map((item) => {
      if (item.trim() === '') {
        return null;
      }

      const isEndOrStartFile = item.match(/^\s*(\.{3}|\-{3})/);
      if (isEndOrStartFile) {
        return null;
      }

      const isComment = item.match(/^\s*#/);
      if (isComment) {
        return null;
      }

      const match = item.match(/^\s*(-\s|)/);
      if (match) {
        return match[0].length === minSpacing ? null : match[0].length;
      }

      return 0;
    }));
    if (spacing && (spacing - minSpacing) > 0) {
      return (spacing - minSpacing);
    }
    return 4;
  }

  randomStepName(length: number): string {
    let result = '';
    const characters = 'abcdefghijklmnopqrstuvwxyz0123456789';
    const charactersLength = characters.length;
    for (let i = 0; i < length; i++) {
      result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
  }

  generateSteps() {
    const steps = [];
    if (
      _.get(this, 'editor.value.steps', null) &&
      _.isObjectLike(this.editor.value.steps)
    ) {
      _.each(this.editor.value.steps, (data: any, key: string) => {
        steps.push({ key, data });
      });
      this.steps = steps;
    } else {
      this.steps = [];
    }
  }

  tryConvertToObject() {
    try {
      this.editor.value = jsYaml.safeLoad(this.editor.text);
      this.generateSteps();
      this.editor.ace.getSession().clearAnnotations();
      this.editor.error = null;
    } catch (err) {
      if (this.editor.ace) {
        this.editor.ace.getSession().setAnnotations([{
          row: err.mark.line,
          column: 0,
          text: err.message,
          type: "error"
        }]);
      }
      this.editor.error = err;
      this.editor.value = null;
      this.steps = [];
    }
  }

  clearMarkers() {
    this.editor.markerIds.forEach((mId: number) => {
      this.editor.ace.session.removeMarker(mId);
    });
  }

  clearSelectedMarkers() {
    this.editor.selectedMarkerIds.forEach((mId: number) => {
      this.editor.ace.session.removeMarker(mId);
    });
  }

  toYaml(obj: any, hasStep: boolean): string[] {
    let str;
    if (hasStep) {
      JSToYaml.spacingStart = ' '.repeat(this.editor.minimumSpacing + this.editor.spacing);
    }
    str = JSToYaml.stringify(obj).value;
    JSToYaml.spacingStart = ' '.repeat(this.editor.minimumSpacing);
    return str.split('\n');
  }

  injectText(from: string[], text: string[], position) {
    text.forEach((s: string) => {
      from.splice(position, 0, s);
      position++;
    });
  }

  addStep(type: string) {
    this.tryConvertToObject();
    this.clearMarkers();
    const textEditor: string[] = this.editor.text.split('\n');
    const stepsRow: number = this.TemplateYamlHelper.getStepsRow(this.editor.text, this.getMinimumSpacing(this.editor.text));
    const obj = stepsRow === -1 ? { steps: {} } : {};
    const randomName = `Step_${this.randomStepName(16)}`;
    if (stepsRow === -1) {
      obj.steps[randomName] = this.types[type].value;
    } else {
      obj[randomName] = this.types[type].value;
    }

    const textToInject = this.toYaml(obj, stepsRow > -1);
    this.injectText(textEditor, textToInject, stepsRow + 1);
    this.setText(textEditor.join('\n'), stepsRow, textToInject.length);
  }

  setText(text: string, stepPosition: number, linesNumber: number) {
    this.editor.ace.setValue(text);
    this.editor.ace.clearSelection();
    const range = new Range(stepPosition, 0, stepPosition + linesNumber, 0);
    const markerId: number = this.editor.ace.session.addMarker(range, 'ace_active-highlight', 'fullLine', true);
    this.editor.markerIds.push(markerId);
    this.editor.ace.gotoLine(stepPosition, 0, true);
    setTimeout(() => {
      this.editor.ace.session.removeMarker(markerId);
    }, 850);
    this.tryConvertToObject();
  }
}
