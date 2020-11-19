import { Component, Input, Output, OnChanges, SimpleChanges, EventEmitter } from '@angular/core';
import remove from 'lodash-es/remove';
import map from 'lodash-es/map';
import uniq from 'lodash-es/uniq';
import compact from 'lodash-es/compact';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { WorkflowService } from '../../@services/workflow.service';
import EditorConfig from '../../@models/editorconfig.model';
import { ModalApiYamlComponent } from '../../@modals/modal-api-yaml/modal-api-yaml.component';
import JSToYaml from 'convert-yaml';

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
    editorConfigPayload: EditorConfig = {
        readonly: true,
        maxLines: 10,
    };
    editorConfigError: EditorConfig = {
        readonly: true,
        maxLines: 10,
    };
    editorConfigChildren: EditorConfig = {
        readonly: true,
        maxLines: 20,
    };
    filteredStepNames: string[];
    states: any = null;
    JSON = JSON;
    presentStates: string[] = [];
    defaultState;

    constructor(private modalService: NgbModal, private workflowService: WorkflowService) {
        this.defaultState = this.workflowService.defaultState;
        this.states = this.workflowService.getMapStates();
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
        const previewModal = this.modalService.open(ModalApiYamlComponent, {
            size: 'xl'
        });
        previewModal.componentInstance.title = `Step - ${step.name}`;
        previewModal.componentInstance.apiCall = () => {
            return new Promise((resolve) => {
                JSToYaml.spacingStart = ' '.repeat(0);
                JSToYaml.spacing = ' '.repeat(4);
                let text = JSToYaml.stringify(step).value;
                resolve(text);
            });
        };
    }

    setPresentStates() {
        this.presentStates = [];
        Object.keys(this.resolution.steps).forEach((key: string) => {
            this.presentStates.push(`State:${this.resolution.steps[key].state}`);
        });
        this.presentStates = uniq(this.presentStates);
    }

    getIcon(state: string) {
        return this.workflowService.getState(state).icon;
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
