import { Component, OnInit, inject } from '@angular/core';
import { NZ_MODAL_DATA, NzModalRef } from 'ng-zorro-antd/modal';

interface IModalData {
    apiCall: any
    apiCallSubmit: any
}

@Component({
    selector: 'lib-utask-modal-yaml-preview',
    template: `
        <div>
            <utask-loader *ngIf="loaders.main"></utask-loader>
            <lib-utask-error-message [data]="errors.main" *ngIf="errors.main && !loaders.main"></lib-utask-error-message>
            <lib-utask-input-editor [config]="{ language: 'yaml', wordWrap: 'on' }" [(ngModel)]="text"></lib-utask-input-editor>
            <lib-utask-error-message [data]="errors.submit" *ngIf="errors.submit && !loaders.submit"></lib-utask-error-message>
        </div>
        <div *nzModalFooter>
            <button type="button" nz-button (click)="modal.triggerCancel()">Close</button>
            <button type="button" nz-button (click)="submit();" [disabled]="loaders.main || loaders.submit || errors.main">Update</button>
        </div>
  `,
    styleUrls: ['./modal-api-yaml-edit.sass'],
    standalone: false
})
export class ModalApiYamlEditComponent implements OnInit {
    public text: string;
    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};
    result: any;

    readonly nzModalData: IModalData = inject(NZ_MODAL_DATA);

    constructor(
        public modal: NzModalRef
    ) { }

    ngOnInit() {
        this.loaders.main = true;
        this.nzModalData.apiCall().then((data) => {
            this.text = data;
        }).catch((err: any) => {
            this.errors.main = err;
        }).finally(() => {
            this.loaders.main = false;
        });
    }

    submit() {
        this.loaders.submit = true;
        this.nzModalData.apiCallSubmit(this.text).then((data) => {
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


