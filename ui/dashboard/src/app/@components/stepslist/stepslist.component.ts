import { Component, Input, Output, OnChanges, SimpleChanges, EventEmitter } from '@angular/core';
import * as _ from 'lodash';
import EditorConfig from 'src/app/@models/editorconfig.model';
import WorkflowHelper from '../../@services/workflowhelper.service';

import { ModalYamlPreviewComponent } from '../../@modals/modal-yaml-preview/modal-yaml-preview.component';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';

@Component({
    selector: 'steps-list',
    templateUrl: 'stepslist.html',
})
export class StepsListComponent implements OnChanges {
    @Input() steps: any[];
    @Input() selectedStep: string;
    // @Output() public select: EventEmitter<any> = new EventEmitter();
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
    filteredStepNames: [];
    states: any = null;
    JSON = JSON;
    presentStates: string[] = [];

    constructor(private modalService: NgbModal) {
        this.states = WorkflowHelper.getMapStates();
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes.steps && this.steps) {
            this.filterSteps();
            this.setPresentStates();
        } else if (changes.selectedStep) {
            _.remove(this.filter.tags, (tag: string) => {
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
      const previewModal = this.modalService.open(ModalYamlPreviewComponent, {
        size: 'xl'
      });
      previewModal.componentInstance.value = step;
      previewModal.componentInstance.title = `Step - ${step.name}`;
      previewModal.result.catch((err) => {
        console.log(err);
      });
    }

    setPresentStates() {
        this.presentStates = [];
        Object.keys(this.steps).forEach((key: string) => {
            this.presentStates.push(`State:${this.steps[key].state}`);
        });
        this.presentStates = _.uniq(this.presentStates);
    }

    getIcon(state: string) {
        return WorkflowHelper.getState(state).icon;
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

        this.filteredStepNames = _.compact(_.map(this.steps, (i: any, k: string) => {
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
