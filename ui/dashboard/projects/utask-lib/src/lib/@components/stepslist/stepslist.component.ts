import { Component, Input, OnChanges, SimpleChanges } from '@angular/core';
import remove from 'lodash-es/remove';
import map from 'lodash-es/map';
import uniq from 'lodash-es/uniq';
import compact from 'lodash-es/compact';
import { WorkflowService } from '../../@services/workflow.service';
import { ModalApiYamlComponent } from '../../@modals/modal-api-yaml/modal-api-yaml.component';
import JSToYaml from 'convert-yaml';
import { EditorOptions } from 'ng-zorro-antd/code-editor';
import { NzModalService } from 'ng-zorro-antd/modal';

@Component({
    selector: 'lib-utask-steps-list',
    templateUrl: 'stepslist.html',
    styleUrls: ['stepslist.sass'],
})
export class StepsListComponent implements OnChanges {
    @Input() resolution: any;
    @Input() selectedStep: string;
    displayDetails: { [key: string]: boolean } = {};
    filter: any = {
        tags: []
    };
    editorConfigPayload: EditorOptions = {
        readOnly: true,
        wordWrap: 'on',
    };
    editorConfigError: EditorOptions = {
        readOnly: true,
        wordWrap: 'on',
    };
    editorConfigChildren: EditorOptions = {
        readOnly: true,
        wordWrap: 'on',
    };
    filteredStepNames: string[];
    states: any = null;
    JSON = JSON;
    presentStates: string[] = [];
    defaultState;

    constructor(
        private _modal: NzModalService,
        private _workflowService: WorkflowService
    ) {
        this.defaultState = this._workflowService.defaultState;
        this.states = this._workflowService.getMapStates();
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes.resolution && this.resolution.steps) {
            this.filterSteps();
            this.setPresentStates();
        } else if (changes.selectedStep) {
            remove(this.filter.tags, (tag: string) => {
                return tag.startsWith('Step:');
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
        const modal = this._modal.create({
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
