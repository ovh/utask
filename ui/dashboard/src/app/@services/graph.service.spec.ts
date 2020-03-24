import { TestBed, async, inject } from '@angular/core/testing';
import { GraphService } from './graph.service';
const d3 = require('d3');
import dagreD3 from 'dagre-d3';
import $ from 'jquery';
import { Component, ViewChild, ElementRef, AfterViewInit } from '@angular/core';

@Component({
    selector: 'app-test',
    template: `<svg width="1000px" height="1000px" #svg><g></g></svg>`,
})
class TestComponent implements AfterViewInit {
    @ViewChild('svg') svg: ElementRef;

    ngAfterViewInit() {
    }
}

describe('GraphService', () => {
    let graphService: GraphService;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
            ],
            declarations: [
                TestComponent,
            ],
            providers: [
                GraphService
            ],
        }).compileComponents();
        graphService = TestBed.get(GraphService);
    }));

    it('Injection graph service', () => {
        inject([GraphService], (injectedService: GraphService) => {
            expect(injectedService).toBe(graphService);
        });
    });

    it(`Generate steps`, () => {
        var steps = graphService.generateSteps({ steps: null });
        expect(steps).toEqual([]);

        var steps = graphService.generateSteps({
            steps: {
                step1: 1,
                step2: 2
            }
        });
        expect(steps).toEqual([{
            key: 'step1',
            data: 1
        }, {
            key: 'step2',
            data: 2
        }]);
    });

    it(`Check Steps`, () => {
        var response = graphService.checkSteps([{
            key: 'step1',
            data: 1
        }, {
            key: 'step2',
            data: 2
        }]);
        expect(response.valid).toEqual(true);
        expect(response.errorMessage).toEqual('');

        var response = graphService.checkSteps([]);
        expect(response.valid).toEqual(false);
        expect(response.errorMessage).toEqual('The steps list is empty');

        var response = graphService.checkSteps([{
            key: 'step1',
            data: {
                dependencies: ['step3']
            }
        }, {
            key: 'step2',
            data: 2
        }]);
        expect(response.valid).toEqual(false);
        expect(response.errorMessage).toEqual(`Step 'step1' have a dependency to 'step3' which don't exist`);
    });

    it('Clear Nodes Selection', () => {
        var svg = document.createElement("svg");
        svg.innerHTML = '<g class="node_selected"></g><g class="node_selected"></g>';
        graphService.clearNodesSelection(svg);
        expect(svg.innerHTML).toEqual('<g class=""></g><g class=""></g>');

        svg = document.createElement("svg");
        svg.innerHTML = '<g class="node_unselected"></g><g class="edge_unselected"></g>';
        graphService.clearNodesSelection(svg);
        expect(svg.innerHTML).toEqual('<g class=""></g><g class=""></g>');

        svg = document.createElement("svg");
        svg.innerHTML = '<g class="label_unselected"></g><g class="label_unselected"></g>';
        graphService.clearNodesSelection(svg);
        expect(svg.innerHTML).toEqual('<g class=""></g><g class=""></g>');
    });

    it('Init Zoom', () => {
        var svg = document.createElement('svg');
        svg.appendChild(document.createElement('g'));
        const d3SVG = d3.select(svg);
        const myG = d3SVG.select('g');

        const d3Zoom = d3.zoom().on('zoom', () => {
            myG.attr('transform', d3.event.transform);
        });

        graphService.initZoom(10, 100, 5, 100, d3Zoom, d3SVG);
        expect(svg.outerHTML).toEqual('<svg><g transform="translate(5,27.5) scale(9)"></g></svg>');

        graphService.initZoom(10, 45, 10, 45, d3Zoom, d3SVG);
        expect(svg.outerHTML).toEqual('<svg><g transform="translate(2.25,2.25) scale(4.05)"></g></svg>');
    });

    it('drawNodesAndEdges', () => {
        const dagreGraph = new dagreD3.graphlib.Graph().setGraph({});
        const steps = [{
            key: "Step1",
            data: {
                dependencies: ["Step2"],
                description: "Step 1 - description"
            }
        }, {
            key: "Step2",
            data: {
                dependencies: [],
                description: "Step 2 - description"
            }
        }, {
            key: "Step3",
            data: {
                dependencies: ["Step2"],
                description: "Step 3 - description"
            }
        }, {
            key: "Step4",
            data: {
                dependencies: ["Step1"],
                description: "Step 4 - description"
            }
        }];
        graphService.drawNodesAndEdges(steps, dagreGraph, true);
        expect(Object.keys(dagreGraph._nodes).length).toEqual(steps.length);
        expect(Object.keys(dagreGraph._edgeObjs).length).toEqual(3);
    });

    it('draw SVG', () => {
        const fixture = TestBed.createComponent(TestComponent);
        fixture.detectChanges();

        graphService.drawSvg(
            [{
                key: "Step1",
                data: {
                    dependencies: ["Step2"],
                    description: "Step 1 - description"
                }
            }, {
                key: "Step2",
                data: {
                    dependencies: [],
                    description: "Step 2 - description"
                }
            }, {
                key: "Step3",
                data: {
                    dependencies: ["Step2"],
                    description: "Step 3 - description"
                }
            }, {
                key: "Step4",
                data: {
                    dependencies: ["Step1", "Step3", "Step4"],
                    description: "Step 4 - description"
                }
            }],
            true,
            fixture.componentInstance.svg.nativeElement
        );

        const htmlGenerated = fixture.componentInstance.svg.nativeElement.outerHTML;

        expect(htmlGenerated.match(/<g class="node"/g).length).toEqual(4);
        expect(htmlGenerated.match(/<g class="edgePath"/g).length).toEqual(5);
    });
});
