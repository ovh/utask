import { Component, Input, OnChanges } from '@angular/core';
import * as _ from "lodash";

@Component({
    selector: 'error-message',
    template: `
        <div class="alert alert-danger" role="alert">{{text}}</div>
    `,
})
export class ErrorMessageComponent implements OnChanges {
    @Input()
    data: any;
    text = '';

    constructor() {
    }

    ngOnChanges() {
        if (_.isString(this.data)) {
            this.text = this.data;
        } else {
            this.text = _.get(this.data, 'error.error', (_.get(this.data, 'error', 'An error just occured, please retry')));
        }
    }
}
