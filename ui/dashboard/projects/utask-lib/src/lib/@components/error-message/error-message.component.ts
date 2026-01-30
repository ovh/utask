import { HttpErrorResponse } from '@angular/common/http';
import { ChangeDetectionStrategy, ChangeDetectorRef, Component, Input, OnChanges } from '@angular/core';
import isString from 'lodash-es/isString';
import { parse } from 'yaml'

@Component({
    selector: 'lib-utask-error-message',
    template: `
        <nz-alert nzType="error" [nzMessage]="text"></nz-alert>
    `,
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class ErrorMessageComponent implements OnChanges {
    @Input() data: HttpErrorResponse | string;
    text = '';

    constructor(
        private _cd: ChangeDetectorRef
    ) { }

    ngOnChanges() {
        this.text = 'An error just occured, please retry';
        if (!this.data) {
            this._cd.markForCheck();
            return;
        }
        if (isString(this.data)) {
            this.text = this.data;
        } else {
            this.text = this.data.message || this.text;
            if (this.data.error) {
                let error = this.data.error;
                const contentType = this.data.headers.get("content-type");
                if (typeof error == 'string') {
                    if (contentType.indexOf('application/x-yaml') > -1) {
                        try { error = parse(error) } catch { }
                    } else if (contentType.indexOf('application/json') > -1) {
                        try { error = JSON.parse(error) } catch { }
                    }
                }
                if (error.error) {
                    this.text = `${this.text} - Details: ${error.error}`;
                }
            }
        }
        this._cd.markForCheck();
    }
}
