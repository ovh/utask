import { Component, OnInit, inject } from '@angular/core';
import { NZ_MODAL_DATA, NzModalRef } from 'ng-zorro-antd/modal';

interface IModalData {
    apiCall: any
}

@Component({
    selector: 'lib-utask-modal-yaml-preview',
    template: `
        <div>
            <utask-loader *ngIf="loading"></utask-loader>
            <lib-utask-error-message [data]="error" *ngIf="error && !loading"></lib-utask-error-message>
            <lib-utask-editor class="editor" *ngIf="!loading" [ngModel]="text" ngDefaultControl [ngModelOptions]="{standalone: true}" [config]="{ language: 'yaml', readOnly: true, wordWrap: 'on' }">
            </lib-utask-editor>
        </div>
        <div *nzModalFooter>
            <button type="button" nz-button (click)="modal.close()">Close</button>
        </div>
  `,
    styleUrls: ['./modal-api-yaml.sass'],
    standalone: false
})
export class ModalApiYamlComponent implements OnInit {
    public text: string;
    loading = false;
    error = null;

    readonly nzModalData: IModalData = inject(NZ_MODAL_DATA);

    constructor(public modal: NzModalRef) {
    }

    ngOnInit() {
        this.loading = true;
        this.nzModalData.apiCall().then((data) => {
            this.text = data;
        }).catch((err: any) => {
            this.error = err;
        }).finally(() => {
            this.loading = false;
        });
    }
}