import { Component, OnInit, Input, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';
import { forkJoin, throwError } from 'rxjs';
import { catchError, finalize } from 'rxjs/operators';
import Template from '../../@models/template.model';
import { ApiService } from '../../@services/api.service';

@Component({
    selector: 'lib-utask-template-details',
    templateUrl: 'template-details.html',
    styleUrls: ['template-details.sass'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class TemplateDetailsComponent implements OnInit {
    @Input() templateName: string;

    error: any;
    loading = true;
    template: Template;
    templateYAML: string = '';

    constructor(
        private _api: ApiService,
        private _cd: ChangeDetectorRef
    ) { }

    ngOnInit() {
        this.loading = true;
        this._cd.markForCheck();
        forkJoin({
            jsonValue: this._api.template.get(this.templateName),
            yamlValue: this._api.template.getYAML(this.templateName)
        })
            .pipe(
                catchError(err => {
                    this.error = err;
                    return throwError(err);
                }),
                finalize(() => {
                    this.loading = false;
                    this._cd.markForCheck();
                })
            )
            .subscribe(data => {
                this.template = data.jsonValue;
                this.templateYAML = data.yamlValue;
            });
    }
}
