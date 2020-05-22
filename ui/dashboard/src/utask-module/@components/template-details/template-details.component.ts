import { Component, OnInit, Input, Output, OnChanges, SimpleChanges, EventEmitter, ElementRef, ViewChild, AfterViewInit } from '@angular/core';
import * as _ from 'lodash';
import JSToYaml from 'convert-yaml';
import { ApiService } from 'utask-lib';
import EditorConfig from 'utask-lib/@models/editorconfig.model';
import Template from 'utask-lib/@models/template.model';

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

    ngOnInit() {
        this.loading = true;
        this.api.template.get(this.templateName).toPromise().then((data) => {
            this.template = data as Template;
            JSToYaml.spacingStart = ' '.repeat(0);
            JSToYaml.spacing = ' '.repeat(4);
            this.text = JSToYaml.stringify(this.template).value;
        }).catch((err: any) => {
            this.error = err;
        }).finally(() => {
            this.loading = false;
        });
    }
}
