import { HttpErrorResponse } from '@angular/common/http';
import { ChangeDetectionStrategy, ChangeDetectorRef, Component, Input, OnChanges } from '@angular/core';
import get from 'lodash-es/get';
import isString from 'lodash-es/isString';

@Component({
    selector: 'lib-utask-error-message',
    template: `
        <nz-alert nzType="error" [nzMessage]="text"></nz-alert>
    `,
    changeDetection: ChangeDetectionStrategy.OnPush
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
        } else if (this.data.error && this.data.error.error) {
            this.text = this.data.error.error;
        } else if (this.data.message) {
            this.text = this.data.message;
        }
        this._cd.markForCheck();
    }
}
