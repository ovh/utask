import { Component, OnInit, Input } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import EditorConfig from '../../@models/editorconfig.model';

@Component({
    selector: 'app-modal-yaml-preview',
    template: `
        <div>
            <div class="modal-header">
                <h4 class="modal-title" id="modal-basic-title">
                    {{title}}
                </h4>
                <button type="button" class="close" aria-label="Close" (click)="activeModal.dismiss('Cross click')">
                <span aria-hidden="true">&times;</span>
                </button>
            </div>
            <div class="modal-body">
                <utask-loader *ngIf="loading"></utask-loader>
                <lib-utask-error-message [data]="error" *ngIf="error && !loading"></lib-utask-error-message>
                <lib-utask-editor *ngIf="!loading" [value]="text" [config]="config"></lib-utask-editor>
            </div>   
        </div>
  `
})
export class ModalApiYamlComponent implements OnInit {
    @Input() public title: string;
    @Input() apiCall: any;
    public text: string;
    public config: EditorConfig = {
        readonly: true,
        mode: 'ace/mode/yaml',
        theme: 'ace/theme/monokai',
        wordwrap: true,
        maxLines: 40,
    };
    loading = false;
    error = null;

    constructor(public activeModal: NgbActiveModal) {
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
