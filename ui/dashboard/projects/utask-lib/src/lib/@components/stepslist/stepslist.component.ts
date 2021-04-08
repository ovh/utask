import { Component, Input, OnChanges, Output, SimpleChanges, EventEmitter } from '@angular/core';
import map from 'lodash-es/map';
import uniq from 'lodash-es/uniq';
import omit from 'lodash-es/omit';
import compact from 'lodash-es/compact';
import { WorkflowService } from '../../@services/workflow.service';
import { ModalApiYamlComponent } from '../../@modals/modal-api-yaml/modal-api-yaml.component';
import JSToYaml from 'convert-yaml';
import { EditorOptions } from 'ng-zorro-antd/code-editor';
import { NzModalService } from 'ng-zorro-antd/modal';
import { ModalEditResolutionStepStateComponent } from '../../@modals/modal-edit-resolution-step-state/modal-edit-resolution-step-state.component';
import { NzNotificationService } from 'ng-zorro-antd/notification';
import Step from '../../@models/step.model';
import { ModalApiYamlEditComponent } from '../../@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ApiService } from '../../@services/api.service';

@Component({
    selector: 'lib-utask-steps-list',
    templateUrl: 'stepslist.html',
    styleUrls: ['stepslist.sass'],
})
export class StepsListComponent implements OnChanges {
    @Input() resolution: any;
    @Input() selectedStep: string;
    @Output() stepChanged = new EventEmitter<Step>();
    displayDetails: { [key: string]: boolean } = {};
    filter: any = {
        tags: []
    };
    editorConfigPayload: EditorOptions = {
        readOnly: true,
        language: 'json',
    };
    editorConfigError: EditorOptions = {
        readOnly: true,
        language: 'json',
    };
    editorConfigChildren: EditorOptions = {
        readOnly: true,
        language: 'json',
    };
    filteredStepNames: string[];
    states: any = null;
    JSON = JSON;
    presentStates: string[] = [];
    defaultState;

    constructor(
        private _modal: NzModalService,
        private _workflowService: WorkflowService,
        private _notif: NzNotificationService,
        private _api: ApiService
    ) {
        this.defaultState = this._workflowService.defaultState;
        this.states = this._workflowService.getMapStates();
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes.resolution && this.resolution.steps) {
            this.filterSteps();
            this.setPresentStates();
        } else if (changes.selectedStep) {
            this.filter.tags = this.filter.tags.filter((tag: string) => {
                return !tag.startsWith('Step:');
            });
            if (this.selectedStep) {
                this.displayDetails[this.selectedStep] = true;
                this.filter.tags.push(`Step:${this.selectedStep}`);
                this.filterSteps();
            }
            this.filterSteps();
        }
    }

    previewStepDetails(step: any) {
        this._modal.create({
            nzTitle: `Step - ${step.name}`,
            nzContent: ModalApiYamlComponent,
            nzWidth: '80%',
            nzComponentParams: {
                apiCall: () => {
                    return new Promise((resolve) => {
                        JSToYaml.spacingStart = ' '.repeat(0);
                        JSToYaml.spacing = ' '.repeat(4);
                        resolve(JSToYaml.stringify(step).value);
                    });
                }
            }
        });
    }

    updateStepState(step: Step) {
        this._modal.create({
            nzTitle: `Edit ${step.name} state`,
            nzContent: ModalEditResolutionStepStateComponent,
            nzWidth: '80%',
            nzComponentParams: {
                step,
                resolution: this.resolution,
            },
            nzOnOk: async (data) => {
                this._notif.info('', `The step state has been edited to ${data.result}.`);
                step.state = data.result;
                this.stepChanged.emit(step);
            }
        });
    }

    updateStep(step: Step) {
        this._modal.create({
            nzTitle: 'Request preview',
            nzContent: ModalApiYamlEditComponent,
            nzWidth: '80%',
            nzComponentParams: {
                apiCall: () => this._api.resolution.getStep(this.resolution.id, step.name).toPromise().then((d: any) => {
                    JSToYaml.spacingStart = ' '.repeat(0);
                    JSToYaml.spacing = ' '.repeat(4);
                    return JSToYaml.stringify(
                        omit(d, ['state', 'children_steps', 'children_steps_map', 'output', 'metadatas', 'tags', 'children', 'error', 'try_count', 'last_time', 'item'])
                    ).value
                }).catch(err => {
                    throw err;
                }),
                apiCallSubmit: (data: any) => this._api.resolution.updateStepAsYaml(this.resolution.id, step.name, data).toPromise()
            },
            nzOnOk: (data: ModalApiYamlEditComponent) => {
                this._notif.info('', `The step has been edited.`);
                this.stepChanged.emit(data.result);
            }
        });
    }

    setPresentStates() {
        this.presentStates = [];
        Object.keys(this.resolution.steps).forEach((key: string) => {
            this.presentStates.push(`State:${this.resolution.steps[key].state}`);
        });
        this.presentStates = uniq(this.presentStates);
    }

    getIcon(state: string) {
        return this._workflowService.getState(state).icon;
    }

    filterSteps() {
        const statuses = [];
        const words = [];
        let step = '';
        this.filter.tags.forEach((s: string) => {
            if (s.startsWith('State:')) {
                statuses.push(s.split(':')[1]);
            } else if (s.startsWith('Step:')) {
                step = s.split(':')[1];
            } else {
                words.push(s);
            }
        });

        this.filteredStepNames = compact(map(this.resolution.steps, (i: any, k: string) => {
            if (!this.filter.tags.length) {
                return k;
            }
            let isValid = true;
            if (statuses.length && statuses.indexOf(i.state) === -1) {
                isValid = false;
            }

            if (step && step !== k) {
                isValid = false;
            }

            words.forEach((w: string) => {
                if (k.toLowerCase().indexOf(w.toLowerCase()) === -1) {
                    isValid = false;
                }
            });

            if (isValid) {
                return k;
            }
            return null;
        }));
    }
}
