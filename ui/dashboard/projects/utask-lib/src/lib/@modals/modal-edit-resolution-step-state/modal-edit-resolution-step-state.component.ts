import { Component, OnInit, Input } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { NzModalRef } from 'ng-zorro-antd/modal';
import Resolution from '../../@models/resolution.model';
import Step from '../../@models/step.model';
import { ApiService } from '../../@services/api.service';
import orderBy from 'lodash-es/orderBy';

@Component({
    selector: 'lib-utask-modal-edit-resolution-step-state',
    templateUrl: './modal-edit-resolution-step-state.html',
    styleUrls: ['./modal-edit-resolution-step-state.sass']
})
export class ModalEditResolutionStepStateComponent implements OnInit {
    @Input('step') step: Step;
    @Input('resolution') resolution: Resolution;
    modalForm: FormGroup;
    states: string[];
    result: string;

    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};

    constructor(public modal: NzModalRef, private api: ApiService, private fb: FormBuilder) {
    }

    ngOnInit() {
        this.modalForm = this.fb.group(
            {
                stepState: new FormControl(this.step.state, [Validators.required]),
            },
        );
        this.states = orderBy([...['ANY', 'TODO', 'RUNNING', 'DONE', 'CLIENT_ERROR', 'SERVER_ERROR', 'FATAL_ERROR', 'CRASHED', 'PRUNE', 'TO_RETRY', 'RETRY_NOW', 'AFTERRUN_ERROR'], ...this.step.custom_states ?? []], s => s);
    }

    submit() {
        this.loaders.submit = true;
        this.api.resolution.updateStepState(this.resolution.id, this.step.name, this.modalForm.value.stepState).toPromise()
            .then(data => {
                this.errors.submit = null;
                this.result = this.modalForm.value.stepState;
                this.modal.triggerOk();
            }).catch((err: any) => {
                this.errors.submit = err;
            }).finally(() => {
                this.loaders.submit = false;
            });
    }
}
