import StepState from '../@models/stepstate.model';
import { Injectable } from '@angular/core';

@Injectable()
export class WorkflowService {
    public shapes: any[] = [{
        pattern: 'shape_white_striped',
        stroke: '#c9c9c9',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 13.5,
            height: 13.5,
            patternTransform: 'rotate(45)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 0,
                y2: 100,
                strokeWidth: 100,
                stroke: '#f0f0f0'
            }, {
                x1: 0,
                y: 0,
                x2: 0,
                y2: 13.5,
                strokeWidth: 15,
                stroke: '#333',
                opacity: 0.08
            }]
        }
    }, {
        pattern: 'shape_white',
        stroke: '#c9c9c9',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 1,
            height: 1,
            patternTransform: 'rotate(45)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 0,
                y2: 1,
                strokeWidth: 10,
                stroke: '#f0f0f0'
            }]
        }
    }, {
        pattern: 'shape_blue',
        stroke: '#2aa0f0',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 1,
            height: 1,
            patternTransform: 'rotate(45)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 0,
                y2: 1,
                strokeWidth: 10,
                stroke: '#32acff'
            }]
        }
    }, {
        pattern: 'shape_red',
        stroke: '#aa3818',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 10,
            height: 10,
            patternTransform: 'rotate(45)',
            lines: [{
                x1: 8,
                y: 8,
                x2: 8,
                y2: 8,
                strokeWidth: 12,
                stroke: '#b04020'
            }, {
                x1: 0,
                y: 0,
                x2: 10,
                y2: 10,
                strokeWidth: 20,
                stroke: '#b04020',
                opacity: 0.9
            }]
        }
    }, {
        pattern: 'shape_green',
        stroke: '#99d250',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 1,
            height: 1,
            patternTransform: 'rotate(45)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 10,
                y2: 10,
                strokeWidth: 10,
                stroke: '#a0da57'
            }]
        }
    }, {
        pattern: 'shape_orange',
        stroke: '#f09000',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 9.5,
            height: 9.5,
            patternTransform: 'rotate(120)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 0,
                y2: 13.5,
                strokeWidth: 50,
                stroke: '#f09000',
                opacity: 0.85
            }, {
                x1: 0,
                y: 0,
                x2: 0,
                y2: 6.5,
                strokeWidth: 10,
                stroke: '#fab000',
                opacity: 0.95
            }, {
                x1: 0,
                y: 0,
                x2: 0,
                y2: 4.5,
                strokeWidth: 30,
                stroke: '#f09000',
                opacity: 0.95
            }]
        }
    }, {
        pattern: 'shape_black',
        stroke: '#333',
        value: {
            patternUnits: 'userSpaceOnUse',
            width: 9.5,
            height: 9.5,
            patternTransform: 'rotate(120)',
            lines: [{
                x1: 0,
                y: 0,
                x2: 0,
                y2: 13.5,
                strokeWidth: 50,
                stroke: '#333',
                opacity: 0.85
            }]
        }
    }];
    public states: StepState[] = [{
        key: 'TO_RETRY',
        color: '#f09000',
        fontColor: 'white',
        shape: 'shape_orange',
        isFinal: false,
        icon: 'history',
        error: false
    }, {
        key: 'RUNNING',
        color: '#32acff',
        fontColor: 'white',
        shape: 'shape_blue',
        isFinal: false,
        icon: 'sync',
        error: false
    }, {
        key: 'TODO',
        color: '#AAA',
        shape: 'shape_white',
        fontColor: 'black',
        isFinal: false,
        icon: 'hourglass',
        error: false
    }, {
        key: 'EXPANDED',
        color: '#32acff',
        shape: 'shape_blue',
        fontColor: 'white',
        isFinal: false,
        icon: 'clock-circle',
        error: false
    }, {
        key: 'CLIENT_ERROR',
        color: '#b04020',
        shape: 'shape_red',
        fontColor: 'white',
        isFinal: false,
        icon: 'close-circle',
        error: true
    }, {
        key: 'DONE',
        color: '#a0da57',
        shape: 'shape_green',
        fontColor: 'white',
        isFinal: true,
        icon: 'check-circle',
        error: false
    }, {
        key: 'PRUNE',
        color: '#DDD',
        shape: 'shape_white_striped',
        fontColor: 'grey',
        isFinal: true,
        icon: 'stop',
        error: false
    }, {
        key: 'SERVER_ERROR',
        color: '#b04020',
        shape: 'shape_red',
        fontColor: 'white',
        isFinal: false,
        icon: 'close-circle',
        error: true
    }, {
        key: 'FATAL_ERROR',
        color: '#b04020',
        shape: 'shape_red',
        fontColor: 'white',
        isFinal: true,
        icon: 'close-circle',
        error: true
    }];
    public defaultState: any = {
        color: '#ff9803',
        shape: 'shape_orange',
        fontColor: 'white',
        isFinal: true,
        icon: 'check-circle',
        error: false
    };

    getState(key: string): StepState {
        const state = this.states.find(s => s.key === key);
        return state || {
            ...this.defaultState,
            key
        };
    }

    getArrayStates() {
        return this.states;
    }

    getMapStates() {
        const states: any = {};
        this.states.forEach((s: any) => {
            states[s.key] = s;
        });
        return states;
    }
}
