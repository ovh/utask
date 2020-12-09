import { Component, Input, OnChanges } from '@angular/core';

export class VMTaskState {
    label: string;
    color: string;
    value: number;
    key: string;
}

@Component({
    selector: 'lib-utask-chart-task-states',
    templateUrl: './chart-task-states.html',
    styleUrls: ['./chart-task-states.sass']
})
export class ChartTaskStatesComponent implements OnChanges {
    @Input() data: VMTaskState[];

    dataset: any[];
    view: any[] = [700, 400];
    colorScheme = {
        domain: []
    };

    constructor() { }

    ngOnChanges() {
        this.generateChart();
    }

    generateChart() {
        this.colorScheme.domain = this.data.map(d => d.color);
        this.dataset = this.data.map(d => { return { name: d.label, value: d.value }; });
    }
};