import { Component, ViewChild, ElementRef, AfterViewInit, Input, OnChanges } from '@angular/core';
import Chart from 'chart.js';

class VMTaskState {
    label: string;
    color: string;
    value: number;
    key: string;
}

@Component({
    selector: 'app-chart-task-states',
    template: `
        <div><canvas #chart></canvas></div>
    `,
    styleUrls: ['./chart-task-states.sass'],
})
export class ChartTaskStatesComponent implements AfterViewInit, OnChanges {
    @ViewChild('chart', null) chart: ElementRef;
    @Input() data: VMTaskState[];

    ngAfterViewInit() {
        this.generateChart();
    }

    ngOnChanges() {
        if (this.chart) {
            this.generateChart();
        }
    }

    generateChart() {
        new Chart(this.chart.nativeElement, {
            type: "doughnut",
            data: {
                datasets: [
                    {
                        data: this.data.map(d => d.value),
                        backgroundColor: this.data.map(d => d.color),
                        labels: this.data.map(d => d.label),
                    }
                ]
            },
            options: {
                tooltips: {
                    callbacks: {
                        label: (tooltip, data) => {
                            return `${data.datasets[0].data[tooltip.index]} tasks ${data.datasets[0].labels[tooltip.index]}`;
                        }
                    },
                    enabled: true
                },
                responsive: true,
                maintainAspectRatio: false,
                animation: {
                    duration: 0
                },
                hover: {
                    animationDuration: 0, // duration of animations when hovering an item
                },
                responsiveAnimationDuration: 0, // animation duration after a resize
            }
        });
    }
};