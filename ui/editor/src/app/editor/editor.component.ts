import { Observable } from 'rxjs';
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import * as _ from 'lodash';

import * as brace from 'brace';
import 'brace/mode/yaml';
import 'brace/theme/monokai';

import {JSON2YAML} from '../services/json2yaml.service';
import {TemplateYamlHelper} from '../services/templateyamlhelper.service';
import jsYaml from 'js-yaml';

import StepsConfig from '../steps.config';
import Editor from '../models/editor.model';
import { isNgTemplate } from '@angular/compiler';
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

  constructor(private JSON2YAML: JSON2YAML, private TemplateYamlHelper: TemplateYamlHelper) {
    this.JSON2YAML.setSpacing(0, 4);
    this.editor = {
      valid: false,
      text: this.JSON2YAML.stringify(StepsConfig.initValue),
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
    // this.tryConvertToObject();
  }

  ngOnInit() {
    console.log("INIT editor");
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

    // this.editor.snippets = [];
    // JSON2YAML.setSpacing(this.editor.minimumSpacing + this.editor.spacing, this.editor.spacing);
    // Object.keys(stepsConfig.types).forEach((key: string) => {
    //   this.editor.snippets.push({
    //     name: `Add '${key}' step`,
    //     score: 499,
    //     meta: 'Snippet',
    //     snippet: JSON2YAML.stringify({ '${1:step_name}': stepsConfig.types[key].snippet })
    //   });
    // });
    // JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
    langTools.setCompleters([]);
    langTools.addCompleter({
      getCompletions: (editor, session, pos, prefix, callback) => {
        const path = this.TemplateYamlHelper.getPath(this.editor.text, pos.row).join('.');
        let arr = [];

        const currentText = this.editor.text.split('\n')[pos.row];
        const isEmptyRow = currentText.trim() === '';
        const countSpacing = currentText.match(/^\s*/)[0].length;
        const isEndOfLine = currentText.length === pos.column;
        console.log(path, currentText, pos, countSpacing, isEmptyRow, isEndOfLine);

        if (isEndOfLine || isEmptyRow) {

          Object.keys(stepsConfig.types).forEach((key: string) => {
            // this.editor.snippets.push({
            //   name: `Add '${key}' step`,
            //   score: 499,
            //   meta: 'Snippet',
            //   snippet: JSON2YAML.stringify({ '${1:step_name}': stepsConfig.types[key].snippet })
            // });
            // // this.editor.snippets.forEach((s: any) => {
            let s = {
              name: `Add '${key}' step`,
              score: 499,
              meta: 'Snippet',
              snippet: ''
            };
            // Fin de ligne donc passage à la ligne
            if (!isEmptyRow) {
              this.JSON2YAML.setSpacing(this.editor.minimumSpacing + this.editor.spacing, this.editor.spacing);
              s.snippet = `\n${this.JSON2YAML.stringify({ '${1:step_name}': stepsConfig.types[key].snippet })}\n`;
              this.JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
            } else {
              // EMPTY LINE
              console.log(this.editor.minimumSpacing, this.editor.spacing, countSpacing)
              const tmpSpacing = _.max([this.editor.minimumSpacing + this.editor.spacing - countSpacing, 0]);
              console.log(tmpSpacing);
              this.JSON2YAML.setSpacing(tmpSpacing, this.editor.spacing);
              s.snippet = `${this.JSON2YAML.stringify({ '${1:step_name}': stepsConfig.types[key].snippet })}`;
              this.JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
            }
            console.log(s);
            arr.push(s);
            // // });
          });

        }

        if ((isEmptyRow || isEndOfLine) && path.match(/^steps\.[a-zA-Z0-9\-\_\s]+\.action\.configuration/)) {
          // JSON2YAML.setSpacing(this.editor.minimumSpacing + this.editor.spacing * 4, this.editor.spacing);
          if (!isEmptyRow) {
            this.JSON2YAML.setSpacing(this.editor.spacing, this.editor.spacing);
          } else {
            let startedSpacing = (this.editor.minimumSpacing + this.editor.spacing * 4) - currentText.length;
            if (startedSpacing < 0) {
              startedSpacing = 0;
            }
            this.JSON2YAML.setSpacing(startedSpacing, this.editor.spacing);
          }
          // ['GET', 'POST', 'PUT', 'DELETE', 'PATCH'].forEach((m: string) => {
          arr.push({
            name: `Add method`,
            score: 501,
            meta: 'Snippet',
            snippet: (isEmptyRow ? '' : '\n') + this.JSON2YAML.stringify({ method: `\${1:}` }) + (isEndOfLine ? '' : '\n')
          });
          // });
          this.JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
        }
        if (path.match(/^steps\.[a-zA-Z0-9\-\_\s]+\.action\.configuration\.method$/)) {
          arr.push({
            name: 'POST', value: 'POST', score: 502, meta: 'Method'
          });
          arr.push({
            name: 'GET', value: 'GET', score: 502, meta: 'Method'
          });
          arr.push({
            name: 'PATCH', value: 'PATCH', score: 502, meta: 'Method'
          });
          arr.push({
            name: 'DELETE', value: 'DELETE', score: 502, meta: 'Method'
          });
          arr.push({
            name: 'PUT', value: 'PUT', score: 502, meta: 'Method'
          });
        }
        if ((isEmptyRow || isEndOfLine) && path.match(/^steps\.[a-zA-Z0-9\-\_\s]+\.dependencies$/)) {
          const stepsName = this.TemplateYamlHelper.getStepsName(this.editor.text);
          stepsName.map((key: string) => {
            arr.push({ name: key, value: key, score: 500, meta: 'Step' });
          });
        }
        callback(null, arr);

        // const currentText = this.editor.text.split('\n')[pos.row];
        // if (YamlHelper.getKey(currentText, 4) === 'method') {
        //   callback(null, [{
        //     name: 'POST', value: 'POST', score: 501, meta: 'Method'
        //   },
        //   {
        //     name: 'GET', value: 'GET', score: 501, meta: 'Method'
        //   },
        //   {
        //     name: 'PUT', value: 'PUT', score: 501, meta: 'Method'
        //   },
        //   {
        //     name: 'DELETE', value: 'DELETE', score: 501, meta: 'Method'
        //   },
        //   {
        //     name: 'PATCH', value: 'PATCH', score: 501, meta: 'Method'
        //   }]);
        // } else {
        //   const path = YamlHelper.getPath(this.editor.text, pos.row).join('.');
        //   const arr = _.clone(this.editor.snippets);
        //   if (path.match(/^steps\.[a-zA-Z0-9\-\_\s]+\.dependencies$/)) {
        //     const stepsName = YamlHelper.getStepsName(this.editor.text);
        //     stepsName.map((key: string) => {
        //       arr.push({ name: key, value: key, score: 500, meta: 'Step' });
        //     });
        //   }
        //   callback(null, arr);
        // }
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
    const min = _.min(text.split('\n').map((item) => {
      // Empty line
      if (item.trim() === '') {
        return null;
      }

      // Start & End Document (https://fr.wikipedia.org/wiki/YAML)
      const isEndOrStartFile = item.match(/^\s*(\.{3}|\-{3})/);
      if (isEndOrStartFile) {
        return null;
      }

      const isComment = item.match(/^\s*#/);
      if (isComment) {
        return null;
      }

      const match = item.match(/^\s*/);
      if (match) {
        return match[0].length;
      }

      return 0;
    })) || 0;
    return min;
  }

  getSpacing(text: string, minSpacing: number) {
    const spacing = _.min(text.split('\n').map((item) => {
      // Empty line
      if (item.trim() === '') {
        return null;
      }

      // Start & End Document (https://fr.wikipedia.org/wiki/YAML)
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

  randomStepName(length) {
    let result = '';
    const characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
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
      console.log(err);
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
      this.JSON2YAML.setSpacing(this.editor.minimumSpacing + this.editor.spacing, this.editor.spacing);
      str = this.JSON2YAML.stringify(obj);
      this.JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
    } else {
      this.JSON2YAML.setSpacing(this.editor.minimumSpacing, this.editor.spacing);
      str = this.JSON2YAML.stringify(obj);
    }
    return str.split('\n');
  }

  injectText(from: string[], text: string[], position) {
    text.forEach((s: string) => {
      from.splice(position, 0, s);
      position++;
    });
  }

  addStep(type: string) {
    this.clearMarkers();

    const textEditor: string[] = this.editor.text.split('\n');
    const indexSteps: number = this.TemplateYamlHelper.getStepsRow(this.editor.text);//this.whereIsSteps(textEditor);
    const obj = indexSteps === -1 ? { steps: {} } : {};
    const randomName = `Step${this.randomStepName(16)}`;
    if (indexSteps === -1) {
      obj.steps[randomName] = this.types[type].value;
    } else {
      obj[randomName] = this.types[type].value;
    }

    const textToInject = this.toYaml(obj, indexSteps > -1);
    this.injectText(textEditor, textToInject, indexSteps + 1);

    this.setText(textEditor.join('\n'), indexSteps, textToInject.length);
  }

  setText(text: string, stepPosition: number, linesNumber: number) {
    this.editor.ace.setValue(text);
    // this.editor.ace.moveCursorTo(stepPosition, 0, true);
    this.editor.ace.gotoLine(stepPosition, 0, true);
    this.editor.ace.clearSelection();
    const range = new Range(stepPosition, 0, stepPosition + linesNumber, 0);
    const markerId: number = this.editor.ace.session.addMarker(range, 'ace_active-highlight', 'fullLine', true);
    this.editor.markerIds.push(markerId);
    setTimeout(() => {
      this.editor.ace.session.removeMarker(markerId);
    }, 850);
    this.tryConvertToObject();
  }
}
