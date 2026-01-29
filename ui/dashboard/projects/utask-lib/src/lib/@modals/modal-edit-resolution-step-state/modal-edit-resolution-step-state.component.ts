import { Component, OnInit, inject } from '@angular/core';
import { FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { NZ_MODAL_DATA, NzModalRef } from 'ng-zorro-antd/modal';
import Resolution from '../../@models/resolution.model';
import Step from '../../@models/step.model';
import { ApiService } from '../../@services/api.service';
import orderBy from 'lodash-es/orderBy';

interface IModalData {
    step: Step;
    resolution: Resolution;
}

@Component({
    selector: 'lib-utask-modal-edit-resolution-step-state',
    templateUrl: './modal-edit-resolution-step-state.html',
    styleUrls: ['./modal-edit-resolution-step-state.sass'],
    standalone: false
})
export class ModalEditResolutionStepStateComponent implements OnInit {
    modalForm: FormGroup;
    states: string[];
    result: string;

    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};

    readonly nzModalData: IModalData = inject(NZ_MODAL_DATA);

    constructor(public modal: NzModalRef, private api: ApiService, private fb: FormBuilder) {
    }

    ngOnInit() {
        this.modalForm = this.fb.group(
            {
                stepState: new FormControl(this.nzModalData.step.state, [Validators.required]),
            },
        );
        this.states = orderBy([...['ANY', 'TODO', 'RUNNING', 'DONE', 'CLIENT_ERROR', 'SERVER_ERROR', 'FATAL_ERROR', 'CRASHED', 'PRUNE', 'TO_RETRY', 'RETRY_NOW', 'AFTERRUN_ERROR'], ...this.nzModalData.step.custom_states ?? []], s => s);
    }

    submit() {
        this.loaders.submit = true;
        this.api.resolution.updateStepState(this.nzModalData.resolution.id, this.nzModalData.step.name, this.modalForm.value.stepState).toPromise()
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
