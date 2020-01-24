import { timer } from 'rxjs';
import { Injectable } from '@angular/core';
import * as _ from 'lodash';
import dagreD3 from 'dagre-d3';
import WorkflowHelper from '../@services/workflowhelper.service';
import $ from 'jquery';
const d3 = require('d3');
import tippy from 'tippy.js';

@Injectable()
export class GraphService {
    constructor() {
    }

    initZoom(dagreGraphWidth: number, svgWidth: number, dagreGraphHeight: number, svgHeight: number, d3Zoom: any, d3Svg: any, ratioZoom: number) {
        const scaleWidth = svgWidth / dagreGraphWidth * ratioZoom;
        const scaleHeight = svgHeight / dagreGraphHeight * ratioZoom;
        const scale = _.min([scaleWidth, scaleHeight]);
        const w = scale * dagreGraphWidth;
        const h = scale * dagreGraphHeight;
        d3Svg.call(
            d3Zoom.transform,
            d3.zoomIdentity.translate(
                (svgWidth - w) / 2,
                (svgHeight- h) / 2,
            ).scale(scale)
        );
    }

    drawSvg(steps: any[], templateHasBeenExecuted: boolean, svgNativeElement: any) {
        const dagreGraph = new dagreD3.graphlib.Graph().setGraph({});
        this.drawNodesAndEdges(steps, dagreGraph, templateHasBeenExecuted);
        const renderGraph = new dagreD3.render();
        this.addShapesToGraph(renderGraph);
        const d3Svg = d3.select(svgNativeElement);
        const innerSVG = d3Svg.select('g');
        renderGraph(innerSVG, dagreGraph);
        const d3Zoom = d3.zoom().on('zoom', () => {
            innerSVG.attr('transform', d3.event.transform);
        });
        d3Svg.call(d3Zoom);
        let radioZoom = 0.9;
        if (steps.length < 5) {
            radioZoom = 0.3;
        } else if (steps.length < 10) {
            radioZoom = 0.6;
        }
        this.initZoom(
            dagreGraph.graph().width,
            svgNativeElement.width.animVal.value,
            dagreGraph.graph().height,
            svgNativeElement.height.animVal.value,
            d3Zoom,
            d3Svg,
            radioZoom
        );
        innerSVG.selectAll('g.node')
            .attr('title', (stepName: string) => dagreGraph.node(stepName).tooltip)
            .each(function () {
                tippy(this, {
                    content: $(this).attr('title'),
                });
            });
        return innerSVG;
    }

    generateSteps(item) {
        const steps = [];
        if (
            _.get(item, 'steps', null) &&
            _.isObjectLike(steps)
        ) {
            _.each(_.get(item, 'steps', null), (data: any, key: string) => {
                steps.push({ key, data });
            });
            return steps;
        } else {
            return [];
        }
    }

    checkSteps(steps: any[]): any {
        const response = {
            valid: true,
            errorMessage: ''
        };
        if (steps.length === 0) {
            return {
                valid: false,
                errorMessage: 'The steps list is empty'
            };
        }
        const keys = steps.map((s: any) => s.key);

        steps.forEach((step: any) => {
            const dependencies = _.get(step.data, 'dependencies', []);
            if (dependencies && dependencies.length) {
                dependencies.forEach((dep: any) => {
                    let d = dep.split(':')[0];
                    if (keys.indexOf(d) === -1) {
                        response.errorMessage = `Step '${step.key}' have a dependency to '${d}' which don't exist`;
                        response.valid = false;
                        return response;
                    }
                });
            }
        });
        return response;
    }

    selectNode(inner, nodeHtmlElement: any, stepName: string) {
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

        $(nodeHtmlElement).addClass('node_selected');

        inner.selectAll('g.edgeLabel').each(function (label) {
            if (stepName !== label.v && stepName !== label.w) {
                $(this).addClass('label_unselected');
            }
        });
    }

    clearNodesSelection(svg: HTMLElement) {
        ['node_selected', 'node_unselected', 'edge_unselected', 'label_unselected'].forEach((className: string) => {
            while (svg.getElementsByClassName(className).length) {
                const node = svg.getElementsByClassName(className)[0];
                node.classList.remove(className);
            }
        });
    }

    addShapesToGraph(render: any) {
        WorkflowHelper.shapes.forEach(shape => {
            render.shapes()[shape.pattern] = function (
                parent: any,
                bbox: any,
                node: any
            ) {
                const w = bbox.width;
                const h = bbox.height;
                const points = [
                    // Bottom left
                    { x: 0, y: 0 },
                    // Bottom right
                    { x: w, y: 0 },
                    // Top Right
                    { x: w, y: -h },
                    // Top Left
                    { x: 0, y: -h }
                ];
                const shapeSvg = parent
                    .insert('polygon', ':first-child')
                    .attr('points', points.map(d => `${d.x},${d.y}`).join(' '))
                    .attr('transform', `translate(${-w / 2},${h / 2})`)
                    .attr('fill', `url(#${shape.pattern})`)
                    .attr('stroke', shape.stroke)
                    .attr('stroke-width', 0)
                    .attr(' :opacity', '1');
                node.intersect = (point: any) => {
                    return dagreD3.intersect.polygon(node, points, point);
                };
                return shapeSvg;
            };
        });
    }

    generateDagreNode(step: any) {
        return {
            shape: step.data.state
                ? WorkflowHelper.getState(step.data.state).shape
                : 'shape_black',
            label: _.replace(step.key, /[A-Z]{1,}/g, s => ` ${s}`),
            labelStyle: `
                    font-size:18px;font-weight:400;text-transform:capitalize;fill:${
                WorkflowHelper.getState(step.data.state).fontColor
                };`,
            tooltip: `
                    <h4 class='cp_h4'>${step.key}${
                step.data.state ? ' : ' : ''
                }${step.data.state ? step.data.state : ''}</h4>
                    <p>${step.data.description}</p>
                    `
        };
    }

    drawNodesAndEdges(steps: any[], dagreGraph: any, templateHasBeenExecuted: boolean) {
        steps.forEach((step: any) => {
            dagreGraph.setNode(step.key, this.generateDagreNode(step));

            (step.data.dependencies || []).forEach((d: any) => {
                const dependencyName = d.split(':')[0];
                const dependencyCondition = d.indexOf(':') > -1 ? d.split(':')[1] : 'DONE';
                if (templateHasBeenExecuted) {
                    const dependencyState = _.get(_.find(steps, { key: dependencyName }), 'data.state');
                    dagreGraph.setEdge(
                        dependencyName,
                        step.key,
                        this.generateDagreArrow(dependencyCondition, dependencyState, step)
                    );
                } else {
                    dagreGraph.setEdge(
                        dependencyName,
                        step.key,
                        this.generateDagreArrowForTemplate(dependencyCondition)
                    );
                }
            });
        });
    }

    generateDagreArrow(dependencyCondition, dependencyState, stepState: string) {
        const arrow: any = {
            rx: 5,
            ry: 5,
            style: 'stroke: black;fill:transparent;stroke-width: 2px;'
        };
        if (dependencyCondition === 'ANY') {
            arrow.label = 'WAIT';
            arrow.arrowhead = 'undirected';
            arrow.labelStyle = 'fill:purple;font-size:18px;';
            arrow.style += 'stroke: #ad0067;';
        } else {
            arrow.arrowhead = 'vee';
            let colorArrow = '';
            if (['TO_RETRY', 'RUNNING', 'TODO', 'EXPANDED', 'CLIENT_ERROR'].indexOf(dependencyState) > -1) {
                colorArrow = WorkflowHelper.getState(stepState).color;
            } else {
                colorArrow = WorkflowHelper.getState(dependencyState).color;
            }
            arrow.style += `stroke: ${colorArrow};fill:transparent;`;
            arrow.arrowheadStyle = `fill: ${colorArrow};stroke: ${colorArrow};`;
        }
        return arrow;
    }

    generateDagreArrowForTemplate(dependencyCondition: string) {
        const arrow: any = {
            rx: 5,
            ry: 5,
            style: 'stroke: black;fill:transparent;',
            labelStyle: 'fill: #333;'
        };
        if (dependencyCondition !== 'ANY') {
            arrow.label = dependencyCondition;
            arrow.style += ';stroke-width: 2px; stroke-dasharray: 2, 2;';
        }
        return arrow;
    }
}
