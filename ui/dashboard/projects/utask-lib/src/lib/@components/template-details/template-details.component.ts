import { Component, OnInit, Input } from '@angular/core';
import JSToYaml from 'convert-yaml';
import Template from '../../@models/template.model';
import { ApiService } from '../../@services/api.service';

@Component({
    selector: 'lib-utask-template-details',
    templateUrl: 'template-details.html',
    styleUrls: ['template-details.sass'],
})
export class TemplateDetailsComponent implements OnInit {
    @Input() templateName: string;
    error: any = null;
    loading = true;
    template: Template;
    text: string;

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
