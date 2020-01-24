import { Component, OnInit, Input, Output, OnChanges, SimpleChanges, EventEmitter, ElementRef, ViewChild, AfterViewInit } from '@angular/core';
import * as _ from 'lodash';
import Template from 'src/app/@models/template.model';
import { ApiService } from '../../@services/api.service';
import EditorConfig from 'src/app/@models/editorconfig.model';
import JSToYaml from 'convert-yaml';

@Component({
    selector: 'template-details',
    templateUrl: 'template-details.html',
})
export class TemplateDetailsComponent implements OnInit {
    @Input() templateName: string;
    error: any = null;
    loading = true;
    template: Template;
    text: string;
    // steps: any[];
    public config: EditorConfig = {
        readonly: true,
        mode: 'ace/mode/yaml',
        theme: 'ace/theme/monokai',
        wordwrap: true,
        maxLines: 50,
    };

    constructor(private api: ApiService) {
    }

    // generateSteps(item) {
    //     const steps = [];
    //     if (
    //       _.get(item, 'steps', null) &&
    //       _.isObjectLike(steps)
    //     ) {
    //       _.each(_.get(item, 'steps', null), (data: any, key: string) => {
    //         steps.push({ key, data });
    //       });
    //       return steps;
    //     } else {
    //       return [];
    //     }
    //   }

    ngOnInit() {
        this.loading = true;
        this.api.getTemplate(this.templateName).toPromise().then((data) => {
            this.template = data as Template;
            JSToYaml.spacingStart = ' '.repeat(0);
            JSToYaml.spacing = ' '.repeat(4);
            this.text = JSToYaml.stringify(this.template).value;
            // this.steps = this.generateSteps(this.template);
        }).catch((err: any) => {
            this.error = err;
        }).finally(() => {
            this.loading = false;
        });
    }
}

/*
public config: EditorConfig = {
    readonly: true,
    mode: 'ace/mode/yaml',
    theme: 'ace/theme/monokai',
    wordwrap: true
  };

  getText(obj: any) {
    JSON2YAML.setSpacing(0, 4);
    return JSON2YAML.stringify(obj);
  }

  generateSteps(item: any) {
    if (!item) {
      return [];
    }
    const steps = [];
    if (
      _.get(this, 'resolution.steps', null) &&
      _.isObjectLike(item)
    ) {
      _.each(item, (data: any, key: string) => {
        steps.push({ key, data });
      });
      return steps;
    }
    return [];
  }
*/