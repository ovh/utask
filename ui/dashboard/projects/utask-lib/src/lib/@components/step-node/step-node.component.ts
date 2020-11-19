import { Component, Input, OnInit, ElementRef, ViewChild, AfterViewInit, Output, EventEmitter, OnChanges } from "@angular/core";
import tippy from 'tippy.js';
import Step from "../../@models/step.model";

@Component({
    selector: 'utask-step-node',
    template: `
        <div #div class="step step-{{styleClass}}" (click)="clickNode()">
            <div class="title">{{key}}</div>
        </div>
    `,
    styleUrls: ['./step-node.sass']
})
export class StepNodeComponent implements OnInit, AfterViewInit, OnChanges {
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

    ngAfterViewInit() {
        tippy(this.div.nativeElement, {
            content: `${this.key} - ${this.step.state}<br/>${this.step.description}`,
            allowHTML: true,
            animation: 'scale'
        });
    }

    clickNode() {
        this.click.emit(this.step.name);
    }
}
