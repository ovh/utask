import * as _ from 'lodash';
import { Injectable } from '@angular/core';

@Injectable()
export class TemplateYamlHelper {
    minimumSpacing: number;
    spacing: number;

    constructor() {
        this.minimumSpacing = 0;
        this.spacing = 4;
    }

    updateSpacing(min: number, sp: number) {
        this.minimumSpacing = min;
        this.spacing = sp;
    }

    getStepsName(text: string) {
        const textArr = text.split('\n');
        const rep = [];
        // const reg = new RegExp(`^\\s{${this.minimumSpacing + this.spacing}}"{0,1}([a-zA-Z\\-\\_0-9]+)"{0,1}\s*:`);
        const reg = new RegExp(`^\\s{${this.minimumSpacing + this.spacing}}("([a-zA-Z\\-\\_0-9\\s]+)"|([a-zA-Z\\-\\_0-9]+))\\s*:`);

        textArr.forEach((txt: string) => {
            const stepName = txt.match(reg);
            if (stepName) {
                rep.push(stepName[2]);
            }
        });
        return rep;
    }

    isEmptyRow(text: string) {
        return !text.trim();
    }

    getStepName(path: string): string {
        const pathArr = path.split('.');
        if (path.length > 1 && pathArr[0] === 'steps') {
            return pathArr[1];
        }
        return null;
    }

    isSeparation(text: string): boolean {
        return !!text.match(/^\s*(\.{3}|\-{3})/);
    }

    isComment(line: string): boolean {
        return !!line.match(/^\s*#/);
    }

    getStepsRow(text: string, minimumSpacing: number): number {
        const arrayStr = text.split('\n');
        const reg = new RegExp(`^\\s{${minimumSpacing}}"{0,1}steps"{0,1}:`);
        let stepPosition = -1;
        const breakException = {};
        try {
            arrayStr.forEach((str: string, index: number) => {
                if (str.match(reg)) {
                    stepPosition = index;
                    throw breakException;
                }
            });
        } catch (err) {
            if (err !== breakException) {
                throw err;
            }
        }
        return stepPosition;
    }

    getEndStep(text: string, row: number): number {
        const arrayStr = text.split('\n');
        const reg = new RegExp(`^\\s{${this.minimumSpacing},${this.minimumSpacing + this.spacing}}"{0,1}[a-zA-Z\\-\\_0-9]+"{0,1}\s*:`);
        let i = row + 1;
        for (i; i < arrayStr.length; i++) {
            if (arrayStr[i].match(reg)) {
                break;
            }
        }
        return i;
    }

    getStepRow(text: string, stepname: string) {
        const arrayStr = text.split('\n');
        const reg = new RegExp(`^\\s{${this.minimumSpacing + this.spacing}}"{0,1}${stepname}"{0,1}\s*`);
        let stepPosition = -1;
        const breakException = {};
        try {
            arrayStr.forEach((str: string, index: number) => {
                if (str.match(reg)) {
                    stepPosition = index;
                    throw breakException;
                }
            });
        } catch (err) {
            if (err !== breakException) {
                throw err;
            }
        }
        return stepPosition;
    }

    getPath(text: string, row: number): string[] {
        const textArr = text.split('\n');
        const rep = [];
        let maxStep = Infinity;
        for (let i = row; i >= 0; i--) {
            const key = this.getKey(textArr[i], maxStep);
            if (key) {
                rep.push(key);
                maxStep = (textArr[i].match(/^(\s*)/g)[0].length - this.minimumSpacing) / this.spacing - 1;
            }
        }
        return _.reverse(rep);
    }

    getKey(text: string, spacingStep: number = Infinity) {
        if (!text) {
            return null;
        }
        const maxStep = this.minimumSpacing + this.spacing * spacingStep;
        const minStep = this.minimumSpacing;
        let reg;
        if (maxStep === Infinity) {
            reg = new RegExp(`^\\s{${minStep},}("([a-zA-Z\\-\\_0-9\\s]+)"|([a-zA-Z\\-\\_0-9]+))\\s*:`);
        } else {
            reg = new RegExp(`^\\s{${minStep},${maxStep}}("([a-zA-Z\\-\\_0-9\\s]+)"|([a-zA-Z\\-\\_0-9]+))\\s*:`);
        }
        const rep = text.match(reg);
        if (rep) {
            return rep[2];
        }
        return null;
    }
};
