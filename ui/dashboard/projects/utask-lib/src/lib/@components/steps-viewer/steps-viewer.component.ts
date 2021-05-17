import {
    Component,
    ViewChild,
    AfterViewInit,
    ComponentRef,
    ComponentFactoryResolver,
    ViewContainerRef,
    OnChanges,
    Input,
    Output,
    EventEmitter,
    SimpleChanges
} from '@angular/core';
import * as d3 from 'd3';
import dagreD3 from 'dagre-d3';
import keys from 'lodash-es/keys';
import difference from 'lodash-es/difference';
import Resolution from '../../@models/resolution.model';
import Step from '../../@models/step.model';
import { StepNodeComponent } from '../step-node/step-node.component';
import isEqual from 'lodash-es/isEqual';

interface Edge {
    v: string;
    w: string;
}

interface Graph {
    _isDirected: boolean;
    _isMultigraph: boolean;
    _isCompound: boolean;
    _label?: string;
    _defaultNodeLabelFn?: any;
    _defaultEdgeLabelFn?: any;
    _nodes: any;
    _parent?: any;
    _children?: any;
    _in: any;
    _preds: any;
    _out: any;
    _sucs: any;
    _edgeObjs: any;
    _edgeLabels: any;
    _nodeCount: number;
    _edgeCount: number;
    isDirected: () => boolean;
    isMultigraph: () => boolean;
    isCompound: () => number;
    setGraph: (label: string) => Graph;
    graph: () => string;
    setDefaultNodeLabel: (newDefault: any) => Graph;
    nodeCount: () => number;
    nodes: () => string[];
    sources: () => any;
    sinks: () => any;
    setNodes: (vs: string[], value: any) => Graph;
    setNode: (v: string, value: any) => Graph;
    node: (v: string) => any;
    hasNode: (v: string) => boolean;
    removeNode: (v: string) => Graph;
    setParent: (v, parent: any) => Graph;
    _removeFromParentsChildList: (v: string) => void;
    parent: (v: string) => any;
    children: (v: string) => any[];
    predecessors: (v: string) => any;
    successors: (v: string) => any;
    neighbors: (v: string) => any;
    isLeaf: (v: string) => boolean;
    filterNodes: (filter: any) => any;
    setDefaultEdgeLabel: (newDefault: any) => Graph;
    edgeCount: () => number;
    edges: () => Edge[];
    setPath: (vs, value) => Graph;
    setEdge: (arg1, arg2, arg3?, arg4?) => Graph;
    edge: (v, w, name) => any;
    hasEdge: (v, w, name?) => boolean;
    removeEdge: (v, w, name?) => Graph;
    inEdges: (v, u) => any;
    outEdges: (v, w) => any;
    nodeEdges: (v, w) => any;
}

@Component({
    selector: 'lib-utask-steps-viewer',
    templateUrl: './steps-viewer.html',
    styleUrls: ['./steps-viewer.sass']
})
export class StepsViewerComponent implements AfterViewInit, OnChanges {
    @ViewChild('svg', { read: ViewContainerRef }) svg: ViewContainerRef;
    item = {
        g: null,
        render: null,
        inner: null,
        svg: null,
        zoom: null,
    };
    @Input() resolution: Resolution;
    selectedStep: string;
    @Output() public select: EventEmitter<any> = new EventEmitter();

    nodesComponent = new Map<string, ComponentRef<StepNodeComponent>>();

    constructor(
        private componentFactoryResolver: ComponentFactoryResolver
    ) { }

    ngAfterViewInit() {
        // setTimeout: To let the parent div to set his height
        setTimeout(() => {
            this.item.svg = d3.select(this.svg.element.nativeElement);
            this.item.inner = this.item.svg.append('g');
            this.item.zoom = d3.zoom().filter(() => {
                return d3.event.ctrlKey;
            }).on('zoom', () => {
                this.item.inner.attr('transform', d3.event.transform);
            });
            this.item.svg.call(this.item.zoom);
            this.item.render = new dagreD3.render();
            this.item.g = new dagreD3.graphlib.Graph();
            this.item.g.setGraph({
                nodesep: 70,
                ranksep: 50,
                rankdir: 'TB',
                marginx: 20,
                marginy: 20
            });
            this.item.g.graph().transition = (selection) => {
                return selection.transition().duration(500);
            };
            this.draw(false);
        }, 1)
    }

    ngOnChanges(diff: SimpleChanges) {
        if (this.svg) {
            if (diff.resolution.previousValue.id !== diff.resolution.currentValue.id) {
                this.selectedStep = null;
                this.svg.clear();
                this.svg.element.nativeElement.innerHTML = '';
                this.nodesComponent.clear();
                this.item.inner = this.item.svg.append('g');
                this.item.zoom = d3.zoom().filter(() => {
                    return d3.event.ctrlKey;
                }).on('zoom', () => {
                    this.item.inner.attr('transform', d3.event.transform);
                });
                this.item.svg.call(this.item.zoom);
                this.item.render = new dagreD3.render();
                this.item.g = new dagreD3.graphlib.Graph();
                this.item.g.setGraph({
                    nodesep: 70,
                    ranksep: 50,
                    rankdir: 'LR',
                    marginx: 20,
                    marginy: 20
                });
                this.item.g.graph().transition = (selection) => {
                    return selection.transition().duration(500);
                };
                this.draw(false);
            } else if (!isEqual(diff.resolution.previousValue, diff.resolution.currentValue)) {
                this.draw(true);
            }
        }
    }

    generateNodesAndEdges(g: Graph, resolution: Resolution) {
        let oldNodes = g.nodes();
        let newNodes = keys(resolution.steps);
        let nodesToDelete = difference(oldNodes, newNodes);
        nodesToDelete.forEach(key => {
            g.removeNode(key);
        });

        keys(resolution.steps).forEach((key) => {
            let step = resolution.steps[key];
            let componentRef = this.generateComponent(key, step);
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
            g.setNode(key, {
                labelType: 'html',
                lineInterpolate: 'basis',
                class: `${classSelected}`,
                label: () => componentRef.location.nativeElement,
                padding: 0,
                style: 'fill: #FAFAFA; stroke: #F0F0F0; stroke-width: 1px;',
                minWidth: 550,
                background: 'red',
                height: 100,
            });
        });

        let edges = g.edges();
        edges.forEach((edge: Edge) => {
            g.removeEdge(edge.v, edge.w);
        });

        keys(resolution.steps).forEach((key) => {
            let step = resolution.steps[key];
            (step.dependencies ?? []).forEach(d => {
                let depArray = d.split(':');
                let depName = depArray[0];
                let depState = resolution.steps[depName].state;
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
                g.setEdge(depName, key, {
                    label: `&nbsp;&nbsp;${(depCondition === 'ANY' && classSelected !== 'HIDDEN') ? 'WAIT' : ''}&nbsp;&nbsp;`,
                    class: `${classSelected} ${depCondition} ${depState}`,
                    labelType: 'html',
                });
            });
        });
    }

    selectStep(stepName: string) {
        if (this.selectedStep === stepName) {
            this.selectedStep = null;
        } else {
            this.selectedStep = stepName;
        }
        this.select.emit(this.selectedStep);
        this.draw(true);
    }

    createNodeComponent(key: string, step: Step): ComponentRef<StepNodeComponent> {
        const nodeComponentFactory = this.componentFactoryResolver.resolveComponentFactory(StepNodeComponent);
        const componentRef = nodeComponentFactory.create(this.svg.parentInjector);
        componentRef.instance.step = step;
        componentRef.instance.key = key;
        componentRef.instance.click.subscribe(v => this.selectStep(v));
        componentRef.changeDetectorRef.detectChanges();
        return componentRef;
    }

    generateComponent(key: string, step: Step): any {
        let componentRef = this.nodesComponent.get(key);
        if (!componentRef) {
            componentRef = this.createNodeComponent(key, step);
            this.nodesComponent.set(key, componentRef);
        } else {
            componentRef.instance.step = step;
            componentRef.changeDetectorRef.detectChanges();
            componentRef.instance.ngOnChanges();
        }
        this.svg.insert(componentRef.hostView, this.svg.length ? this.svg.length - 1 : 0);
        return componentRef;
    }

    center() {
        var graphWidth = this.item.g.graph().width;
        var graphHeight = this.item.g.graph().height;
        var width = parseInt(this.item.svg.style('width').replace(/px/, ''));
        var height = parseInt(this.item.svg.style('height').replace(/px/, ''));
        let radioZoom = 0.9;
        if (keys(this.resolution.steps).length < 5) {
            radioZoom = 0.6;
        } else if (keys(this.resolution.steps).length < 9) {
            radioZoom = 0.75;
        }
        var zoomScale = Math.min(width / graphWidth, height / graphHeight) * radioZoom;
        var translateX = (width / 2) - ((graphWidth * zoomScale) / 2)
        var translateY = (height / 2) - ((graphHeight * zoomScale) / 2);
        this.item.svg.call(this.item.zoom.transform, d3.zoomIdentity.translate(translateX, translateY).scale(zoomScale));
    }

    draw(isUpdate: boolean) {
        let transformValue;
        if (isUpdate) {
            transformValue = this.item.svg.select('g').attr('transform');
            transformValue = transformValue.split('scale(')[1].split(')')[0];
            this.item.svg.call(this.item.zoom.scaleTo as any, 1);
        }
        this.generateNodesAndEdges(this.item.g, this.resolution);
        this.item.inner.call(this.item.render, this.item.g);
        if (!isUpdate) {
            this.center();
        }
        if (isUpdate) {
            this.item.svg.call(this.item.zoom.scaleTo as any, transformValue);
        }
    }
}


