import { Component, Input, OnInit, ElementRef, ViewChild, Output, EventEmitter, OnChanges } from '@angular/core';
import Step from '../../@models/step.model';

@Component({
    selector: 'lib-utask-step-node',
    templateUrl: './step-node.html',
    styleUrls: ['./step-node.sass']
})
export class StepNodeComponent implements OnInit, OnChanges {
    @Input() step: Step;
    @Input() key: string;
    @ViewChild('div', { static: false }) div: ElementRef;
    @Output() click = new EventEmitter<string>();

    styleClass: string;

    constructor() {
    }

    ngOnInit() {
        switch (this.step.state) {
            case 'DONE': {
                this.styleClass = 'green';
                break;
            }
            case 'TO_RETRY':
            case 'RUNNING':
            case 'EXPANDED': {
                this.styleClass = 'blue';
                break;
            }
            case 'TODO': {
                this.styleClass = 'grey';
                break;
            }
            case 'PRUNE': {
                this.styleClass = 'prune';
                break;
            }
            case 'CLIENT_ERROR':
            case 'SERVER_ERROR':
            case 'FATAL_ERROR':
                {
                    this.styleClass = 'red';
                    break;
                }
            default: {
                this.styleClass = 'default';
                break;
            }
        };
    }

    ngOnChanges() {
        if (this.div) {
            switch (this.step.state) {
                case 'DONE': {
                    this.styleClass = 'green';
                    break;
                }
                case 'TO_RETRY':
                case 'RUNNING':
                case 'EXPANDED': {
                    this.styleClass = 'blue';
                    break;
                }
                case 'TODO': {
                    this.styleClass = 'grey';
                    break;
                }
                case 'PRUNE': {
                    this.styleClass = 'prune';
                    break;
                }
                case 'CLIENT_ERROR':
                case 'SERVER_ERROR':
                case 'FATAL_ERROR':
                    {
                        this.styleClass = 'red';
                        break;
                    }
                default: {
                    this.styleClass = 'default';
                    break;
                }
            };
        }
    }

    clickNode() {
        this.click.emit(this.step.name);
    }
}
