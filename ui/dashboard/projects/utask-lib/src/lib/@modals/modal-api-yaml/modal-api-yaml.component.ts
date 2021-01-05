import { Component, OnInit, Input } from '@angular/core';
import { NzModalRef } from 'ng-zorro-antd/modal';

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
    styleUrls: ['./modal-api-yaml.sass']
})
export class ModalApiYamlComponent implements OnInit {
    @Input() apiCall: any;
    public text: string;
    loading = false;
    error = null;

    constructor(public modal: NzModalRef) {
    }

    ngOnInit() {
        this.loading = true;
        this.apiCall().then((data) => {
            this.text = data;
        }).catch((err: any) => {
            this.error = err;
        }).finally(() => {
            this.loading = false;
        });
    }
}