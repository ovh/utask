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
                <utask-loader *ngIf="loaders.main"></utask-loader>
                <lib-utask-error-message [data]="errors.main" *ngIf="errors.main && !loaders.main"></lib-utask-error-message>
                <lib-utask-editor *ngIf="!loaders.main && !errors.main" [value]="text" [errors]="errors" [config]="config" (update)="text = $event;"></lib-utask-editor>

                <lib-utask-error-message [data]="errors.submit" *ngIf="errors.submit && !loaders.submit"></lib-utask-error-message>
            </div>   
            <footer class="modal-footer">
                <button type="button" class="btn btn-success" (click)="submit();" [disabled]="loaders.main || loaders.submit || errors.main">
                    Update
                </button>
            </footer>
        </div>
  `
})
export class ModalApiYamlEditComponent implements OnInit {
    @Input() public title: string;
    @Input() apiCall: any;
    @Input() apiCallSubmit: any;
    public text: string;
    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};
    public config: EditorConfig = {
        readonly: false,
        mode: 'ace/mode/yaml',
        theme: 'ace/theme/monokai',
        wordwrap: true,
        maxLines: 40,
    };

    constructor(public activeModal: NgbActiveModal) {
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
            this.activeModal.close(data);
        }).catch((err: any) => {
            this.errors.submit = err;
        }).finally(() => {
            this.loaders.submit = false;
        });
    }
}
