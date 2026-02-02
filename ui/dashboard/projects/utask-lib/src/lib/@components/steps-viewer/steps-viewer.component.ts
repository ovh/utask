import {
    Component,
    ViewChild,
    OnChanges,
    Input,
    Output,
    EventEmitter,
    ChangeDetectorRef,
    ChangeDetectionStrategy
} from '@angular/core';
import keys from 'lodash-es/keys';
import { Resolution } from '../../@models/resolution.model';
import { NzGraphComponent, NzGraphData, NzGraphLayoutConfig, NzGraphZoomDirective, NzRankDirection } from "ng-zorro-antd/graph";

@Component({
    selector: 'lib-utask-steps-viewer',
    templateUrl: './steps-viewer.html',
    styleUrls: ['./steps-viewer.sass'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class StepsViewerComponent implements OnChanges {
    // Inputs & Outputs
    @Input() resolution: Resolution;
    @Output() public select: EventEmitter<any> = new EventEmitter();

    selectedStep: string;

    // Graph
    graphData: NzGraphData;
    rankDirection: NzRankDirection = 'TB';
    @ViewChild(NzGraphComponent, { static: true }) nzGraphComponent!: NzGraphComponent;
    @ViewChild(NzGraphZoomDirective, { static: true }) zoomController!: NzGraphZoomDirective;
    layoutConfig: NzGraphLayoutConfig = {
        defaultNode: {
            width: 350,
            height: 50,
        }
    };
    zoom = 0.5;

    constructor(
        private _cd: ChangeDetectorRef
    ) { }

    ngOnChanges() {
        this.clear();
        this.draw();
    }

    clear() {
        this.graphData = new NzGraphData({
            nodes: [],
            edges: []
        });
    }

    graphInitialized(_ele: NzGraphComponent): void {
        this.fit();
    }

    fit() {
        this.zoomController?.fitCenter();
        this._cd.markForCheck();
    }

    selectNode(step: any) {
        this.selectedStep = this.selectedStep != step.name ? step.name : '';
        this.select.emit(this.selectedStep);
        this.draw();
    }

    draw() {
        let nodes = [];
        let edges = [];
        keys(this.resolution.steps).forEach((key) => {
            let step = this.resolution.steps[key];
            let classSelected = 'STANDARD';
            if (this.selectedStep) {
                if (this.selectedStep === key) {
                    classSelected = 'SELECTION';
                } else if (
                    (step.dependencies ?? []).map(d => d.split(':')[0]).indexOf(this.selectedStep) > -1 ||
                    (this.resolution.steps[this.selectedStep].dependencies ?? []).map(d => d.split(':')[0]).indexOf(key) > -1
                ) {
                    classSelected = 'SELECTED';
                } else {
                    classSelected = 'HIDDEN';
                }
            }

            nodes.push({
                id: key,
                label: key,
                selected: false,
                step,
                class: classSelected,
                width: 350
            });
        });

        keys(this.resolution.steps).forEach((key) => {
            let step = this.resolution.steps[key];
            (step.dependencies ?? []).forEach(d => {
                let depArray = d.split(':');
                let depName = depArray[0];
                let depState = this.resolution.steps[depName].state;
                let depCondition = depArray.length > 1 ? depArray[1] : '';
                let classSelected = 'STANDARD';
                if (this.selectedStep) {
                    if (
                        this.selectedStep === key ||
                        this.selectedStep === depName
                    ) {
                        classSelected = 'SELECTED';
                    } else {
                        classSelected = 'HIDDEN';
                    }
                }

                edges.push({
                    v: depName,
                    w: key,
                    markerEnd: (depCondition === 'ANY' && classSelected !== 'HIDDEN') ? 'url(#edge)' : 'url(#edge-end-arrow)',
                    classes: `${classSelected} ${depCondition} ${depState}`,
                    label: (depCondition === 'ANY' && classSelected !== 'HIDDEN') ? 'WAIT' : ''
                });
            });
        });

        this.graphData = new NzGraphData({
            nodes,
            edges,
        });
        this._cd.markForCheck();
    }
}
