import * as _ from 'lodash';
import { Injectable } from '@angular/core';

@Injectable()
export class JSON2YAML {
  spacing: string;
  spacingStart: string;

  constructor() {
    this.spacing = _.repeat(' ', 4);
    this.spacingStart = _.repeat(' ', 8);
  }

  setSpacing(startCountSpacing: number, countSpacing: number) {
    this.spacing = _.repeat(' ', countSpacing);
    this.spacingStart = _.repeat(' ', startCountSpacing);
  }

  getType(obj) {
    const type = typeof obj;
    if (obj instanceof Array) {
      return 'array';
    } else if (type === 'string' && obj.indexOf('\n') > -1) {
      return 'string_multiline';
    } else if (type === 'string' && obj.indexOf('\n') === -1) {
      return 'string';
    } else if (type === 'boolean') {
      return 'boolean';
    } else if (type === 'number') {
      return 'number';
    } else if (type === 'undefined' || obj === null) {
      return 'null';
    } else {
      return 'hash';
    }
  }

  convert(obj: any, ret: any, pos: number = 0) {
    const type = this.getType(obj);
    switch (type) {
      case 'array':
        this.convertArray(obj, ret);
        break;
      case 'hash':
        this.convertHash(obj, ret);
        break;
      case 'string':
        this.convertString(obj, ret);
        break;
      case 'string_multiline':
        this.convertStringMultiline(obj, ret);
        break;
      case 'null':
        ret.push('null');
        break;
      case 'number':
        ret.push(obj.toString());
        break;
      case 'boolean':
        ret.push(obj ? 'true' : 'false');
        break;
    }
  }

  convertArray(obj, ret) {
    if (obj.length === 0) {
      ret.push('[]');
    }
    obj.forEach((o: any) => {
      const recurse = [];
      this.convert(o, recurse);
      recurse.forEach((item: any, index: number) => {
        if (!index) {
          ret.push('- ');
        }
        ret.push(this.spacing + item);
      });
    });
  }

  convertHash(obj, ret) {
    for (const k of Object.keys(obj)) {
      const recurse = [];
      if (obj.hasOwnProperty(k)) {
        const ele = obj[k];
        this.convert(ele, recurse);
        const type = this.getType(ele);
        if (['string', 'null', 'number', 'boolean'].indexOf(type) > -1) {
          ret.push(`${this.normalizeString(k)}: ${recurse[0]}`);
        } else if (type === 'string_multiline') {
          ret.push(`${this.normalizeString(k)}: ${recurse[0]}`);
        } else {
          ret.push(this.normalizeString(k) + ': ');
          recurse.forEach((d: string) => {
            ret.push(`${this.spacing}${d}`);
          });
        }
      }
    }
  }

  normalizeString(str) {
    return `"${str}"`;
  }

  convertString(obj, ret) {
    if (obj.match(/\s/)) {
      ret.push(this.normalizeString(obj));
    } else {
      ret.push(`${obj}`);
    }
  }

  convertStringMultiline(obj, ret) {
    let str = '|-\n';
    const arrObj = obj.split('\n');
    arrObj.forEach((s, i) => {
      str += this.spacingStart + this.spacing + s + ((arrObj.length - 1 !== i) ? '\n' : '');
    });
    ret.push(str);
  }

  stringify(obj: any) {
    if (typeof obj === 'string') {
      obj = JSON.parse(obj);
    }

    const ret = [];
    this.convert(obj, ret);
    return this.spacingStart + ret.join(`\n${this.spacingStart}`);
  }
};