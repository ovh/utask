import { Component, OnInit, Input, Output, OnChanges, SimpleChanges, EventEmitter } from '@angular/core';
import * as _ from 'lodash';

import {WorkflowHelper} from '../services/workflowhelper.service';

const d3 = require('d3');
// import d3 from 'd3';
import dagreD3 from 'dagre-d3';
import $ from 'jquery';

@Component({
    selector: 'steps-viewer',
    templateUrl: 'stepsviewer.html',
})
export class StepsViewerComponent implements OnChanges {
    @Input() steps: any[];
    @Output() public select: EventEmitter<any> = new EventEmitter();
    @Input() selectParam;

    constructor(private WorkflowHelper: WorkflowHelper) {

    }

    error: any = null;
    graph: any = {
        dimensions: null
    };
    done = false;
    selectedNode: any;

    ngOnChanges(changes: SimpleChanges) {
        if (changes.steps) {
            this.init();
            this.unselectAll();
            this.selectedNode = null;
            this.select.emit(null);
        }
    }

    init() {
        try {
            if (this.checkData()) {
                const dagreGraph = new dagreD3.graphlib.Graph().setGraph({});
                this.createNodesAndEdges(dagreGraph);
                const renderGraph = new dagreD3.render();
                this.applyShapes(renderGraph);
                const svg = d3.select('svg');
                const innerSVG = svg.select('g');
                const zoom = d3.zoom().on('zoom', () => {
                    innerSVG.attr('transform', d3.event.transform);
                });
                svg.call(zoom);
                renderGraph(innerSVG, dagreGraph);
                const self = this;
                innerSVG.selectAll('g.node')
                    .attr('title', v => dagreGraph.node(v).tooltip)
                    .on('click', function (v: any) {
                        self.doSelection(innerSVG, this, v);
                    })
                // .each(function () {
                //     $(this).tipsy();
                // });
                this.graph.dimensions = {
                    width: dagreGraph.graph().width,
                    height: dagreGraph.graph().height
                };
                this.center(svg, this.graph.dimensions, zoom);
                this.error = null;
            }
        } catch (exc) {
            this.error = 'An error occured, the template is invalid';
        }
    }

    unselectAll() {
        $('.node_selected').removeClass('node_selected');
        $('.node_unselected').removeClass('node_unselected');
        $('.edge_unselected').removeClass('edge_unselected');
        $('.label_unselected').removeClass('label_unselected');
    }

    doSelection(inner, nodeHtmlELement, stepName) {
        this.unselectAll();
        if (this.selectedNode === stepName) {
            this.selectedNode = null;
        } else {
            this.selectedNode = stepName;
            const list = [];
            inner.selectAll('g.edgePath').each(function (edge) {
                if (stepName !== edge.v && stepName !== edge.w) {
                    $(this).addClass('edge_unselected');
                } else {
                    if (list.indexOf(edge.v) === -1) {
                        list.push(edge.v);
                    }
                    if (list.indexOf(edge.w) === -1) {
                        list.push(edge.w);
                    }
                }
            });

            inner.selectAll('g.node').each(function (node) {
                if (list.indexOf(node) === -1 && stepName !== node) {
                    $(this).addClass('node_unselected');
                }
            });

            $(nodeHtmlELement).addClass('node_selected');

            inner.selectAll('g.edgeLabel').each(function (label) {
                if (stepName !== label.v && stepName !== label.w) {
                    $(this).addClass('label_unselected');
                }
            });
        }
        this.select.emit(this.selectedNode);
    }

    checkData() {
        let response = true;
        const keys = this.steps.map((s: any) => s.key);
        if (!_.isArray(this.steps)) {
            this.error = 'The steps list is not an array';
            return false;
        } else if (this.steps.length === 0) {
            this.error = 'The steps list is empty or invalid';
            return false;
        } else if (_.uniq(keys).length !== this.steps.length) {
            this.error = 'Duplicate steps name';
            return false;
        }

        this.steps.forEach((step: any) => {
            const dependencies = _.get(step.data, 'dependencies', []);
            if (dependencies && dependencies.length) {
                dependencies.forEach((dep: any) => {
                    let d = dep.split(':')[0];
                    if (keys.indexOf(d) === -1) {
                        this.error = `Step '${step.key}' have a dependency to '${d}' which don't exist`;
                        response = false;
                    }
                });
            }
        });
        return response;
    }

    createNodesAndEdges(g: any) {
        this.steps.forEach((step: any) => {
            g.setNode(step.key, {
                shape: step.data.state
                    ? this.WorkflowHelper.getState(step.data.state).shape
                    : 'shape_black',
                label: _.replace(step.key, /[A-Z]{1,}/g, s => ` ${s}`),
                // labelType: 'html',
                labelStyle: `
                        font-size:18px;font-weight:400;text-transform:capitalize;fill:${
                    this.WorkflowHelper.getState(step.data.state).fontColor
                    };`,
                tooltip: `
                        <h4 class='cp_h4'>${step.key}${
                    step.data.state ? ' : ' : ''
                    }${step.data.state ? step.data.state : ''}</h4>
                        <p>${step.data.description}</p>
                        `
            });
            if (this.done) {
                (step.data.dependencies || []).forEach((d: any) => {
                    let stepName = d;
                    let stepCondition = '';
                    if (d.indexOf(':') > -1) {
                        stepName = d.split(':')[0];
                        stepCondition = d.split(':')[1];
                    }
                    let arrow: any = {};
                    if (stepCondition === 'ANY') {
                        arrow = {
                            rx: 5,
                            ry: 5,
                            label: 'WAIT',
                            labelStyle: 'fill:purple;font-size:18px;',
                            /*
                                              label: '',
                                              labelStyle: 'font-family: 'Font Awesome 5 Free';fill:orange;font-size:18px;',
                                              */
                            arrowhead: 'undirected',
                            style: 'stroke: black;fill:transparent;stroke-width: 2px;'
                        };
                    } else {
                        arrow = {
                            rx: 5,
                            ry: 5,
                            arrowhead: 'vee',
                            style: 'stroke: black;fill:transparent;stroke-width: 2px;'
                        };
                    }

                    if (stepCondition === 'ANY') {
                        arrow.style += 'stroke: #ad0067;';
                    } else {
                        let fromState = this.getStateFromStepname(stepName);
                        let colorArrow = '';
                        if (
                            [
                                'TO_RETRY',
                                'RUNNING',
                                'TODO',
                                'EXPANDED',
                                'CLIENT_ERROR'
                            ].indexOf(fromState) > -1
                        ) {
                            colorArrow = this.WorkflowHelper.getState(step.data.state).color;
                        } else {
                            colorArrow = this.WorkflowHelper.getState(fromState).color;
                        }
                        arrow.style += `stroke: ${colorArrow};fill:transparent;`;
                        arrow.arrowheadStyle = `fill: ${colorArrow};stroke: ${colorArrow};`;
                    }
                    g.setEdge(stepName, step.key, arrow);
                });
            } else {
                (step.data.dependencies || []).forEach((d: any) => {
                    let stepName = d;
                    let stepCondition = '';
                    if (d.indexOf(':') > -1) {
                        stepName = d.split(':')[0];
                        stepCondition = d.split(':')[1];
                    }
                    if (stepCondition === 'ANY') {
                        g.setEdge(stepName, step.key, {
                            rx: 5,
                            ry: 5,
                            style: 'stroke: #333;fill:transparent;',
                            labelStyle: 'fill: #333;'
                        });
                    } else {
                        g.setEdge(stepName, step.key, {
                            label: !this.done ? stepCondition || 'DONE' : ' ',
                            rx: 5,
                            ry: 5,
                            style:
                                'stroke: #333;fill:transparent;stroke-width: 2px; stroke-dasharray: 2, 2;',
                            labelStyle: 'fill: #333;'
                        });
                    }
                });
            }
        });
    }

    getStateFromStepname(stepName: any) {
        return _.get(_.find(this.steps, { key: stepName }), 'data.state');
    }

    applyShapes(render: any) {
        this.WorkflowHelper.shapes.forEach(shape => {
            render.shapes()[shape.pattern] = function (
                parent: any,
                bbox: any,
                node: any
            ) {
                var w = bbox.width;
                var h = bbox.height;
                var points = [
                    // Bottom left
                    { x: 0, y: 0 },
                    // Bottom right
                    { x: w, y: 0 },
                    // Top Right
                    { x: w, y: -h },
                    // Top Left
                    { x: 0, y: -h }
                ];
                var shapeSvg = parent
                    .insert('polygon', ':first-child')
                    .attr('points', points.map(d => `${d.x},${d.y}`).join(' '))
                    .attr('transform', `translate(${-w / 2},${h / 2})`)
                    .attr('fill', `url(#${shape.pattern})`)
                    .attr('stroke', shape.stroke)
                    .attr('stroke-width', 0)
                    .attr(' :opacity', '1');
                node.intersect = function (point: any) {
                    return dagreD3.intersect.polygon(node, points, point);
                };
                return shapeSvg;
            };
        });
    }

    center(svg: any, dimensions: any, zoom: any) {
        let svgDimensions = {
            width: svg._groups[0][0].width.animVal.value,
            height: svg._groups[0][0].height.animVal.value
        };
        let initialScale = 0.6;
        svg.call(
            zoom.transform,
            d3.zoomIdentity
                .translate(
                    -(dimensions.width * initialScale - svgDimensions.width) / 2,
                    -(dimensions.height * initialScale - svgDimensions.height) / 2
                )
                .scale(initialScale)
        );
    }
}
