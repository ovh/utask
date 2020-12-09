import { Component, OnInit, Input } from '@angular/core';
import { NzModalRef } from 'ng-zorro-antd/modal';

@Component({
    selector: 'lib-utask-modal-yaml-preview',
    template: `
        <div>
            <utask-loader *ngIf="loaders.main"></utask-loader>
            <lib-utask-error-message [data]="errors.main" *ngIf="errors.main && !loaders.main"></lib-utask-error-message>
            <lib-utask-editor *ngIf="!loaders.main && !errors.main" [(ngModel)]="text" ngDefaultControl [ngModelOptions]="{standalone: true}" [config]="{ language: 'yaml', readOnly: false, wordWrap: 'on' }"></lib-utask-editor>
            <lib-utask-error-message [data]="errors.submit" *ngIf="errors.submit && !loaders.submit"></lib-utask-error-message>
        </div>
        <div *nzModalFooter>
            <button type="button" nz-button (click)="modal.triggerCancel()">Close</button>
            <button type="button" nz-button (click)="submit();" [disabled]="loaders.main || loaders.submit || errors.main">Update</button>
        </div>
  `,
    styleUrls: ['./modal-api-yaml-edit.sass'],
})
export class ModalApiYamlEditComponent implements OnInit {
    @Input() apiCall: any;
    @Input() apiCallSubmit: any;
    public text: string;
    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};
    result: any;

    constructor(public modal: NzModalRef) {
    }

    ngOnInit() {
        this.loaders.main = true;
        this.apiCall().then((data) => {
            this.text = data;
        }).catch((err: any) => {
            this.errors.main = err;
        }).finally(() => {
            this.loaders.main = false;
        });
    }

    submit() {
        this.loaders.submit = true;
        this.apiCallSubmit(this.text).then((data) => {
            this.errors.submit = null;
            this.result = data;
            this.modal.triggerOk();
        }).catch((err: any) => {
            this.errors.submit = err;
        }).finally(() => {
            this.loaders.submit = false;
        });
    }
}


