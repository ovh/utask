import { Component, Input, Output, OnChanges, SimpleChanges, EventEmitter, ElementRef, ViewChild, AfterViewInit } from '@angular/core';
import * as _ from 'lodash';

import { GraphService } from 'src/app/@services/graph.service';

@Component({
    selector: 'steps-viewer',
    templateUrl: 'stepsviewer.html',
})
export class StepsViewerComponent implements OnChanges, AfterViewInit{
    @ViewChild('svg', null) svg: ElementRef;
    @Input() item: any;
    // TODO: RENAME DONE OU DELETE IT
    @Input() done: boolean;
    @Output() public select: EventEmitter<any> = new EventEmitter();

    error: any = null;
    selectedNode: any;
    loaded = false;

    constructor(private graphService: GraphService) {
    }

    ngOnChanges(changes: SimpleChanges) {
        if (changes.item && this.loaded) {
            const steps = this.graphService.generateSteps(this.item);
            const dataValid = this.graphService.checkSteps(steps);
            if (dataValid.valid) {
                this.draw(steps);
                this.selectedNode = null;
                this.select.emit('');
            } else {
                this.error = dataValid.errorMessage;
            }
        }
    }

    // SVG will be created
    ngAfterViewInit() {
        const steps = this.graphService.generateSteps(this.item);
        const dataValid = this.graphService.checkSteps(steps);
        if (dataValid.valid) {
            this.draw(steps);
        } else {
            this.error = dataValid.errorMessage;
        }
    }

    draw(steps: any[]) {
        try {
            const self = this;
            const innerSVG = this.graphService.drawSvg(steps, this.done, this.svg.nativeElement);
            innerSVG.selectAll('g.node')
                .on('click', function (stepName: string) {
                    let nodeHtmlElement = this;
                    self.graphService.clearNodesSelection(self.svg.nativeElement);
                    if (self.selectedNode === stepName) {
                        self.selectedNode = null;
                    } else {
                        self.selectedNode = stepName;
                        self.graphService.selectNode(innerSVG, nodeHtmlElement, stepName);
                    }
                    self.select.emit(self.selectedNode);
                });
            this.error = null;
        } catch (exc) {
            console.log(exc);
            this.error = 'An error occured, the template is invalid';
        }
        this.loaded = true;
    }
}
