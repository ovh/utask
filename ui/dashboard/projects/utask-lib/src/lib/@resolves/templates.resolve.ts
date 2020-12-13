import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { Resolve } from '@angular/router';
import { HttpHeaders } from '@angular/common/http';
import { ApiService, ParamsListTemplates, UTaskLibOptions } from '../@services/api.service';
import Template from '../@models/template.model';

@Injectable()
export class TemplatesResolve implements Resolve<any> {
    constructor(
        private _api: ApiService,
        private _router: Router,
        private _options: UTaskLibOptions
    ) { }

    hasLast(headers: HttpHeaders, pagination: any) {
        const link = headers.get('link');
        if (link) {
            const match = link.match(/last=([^&;\s>]+)/);
            if (match) {
                pagination.last = match[1];
                return true;
            } else {
                pagination.last = '';
                return false;
            }
        } else {
            pagination.last = '';
            return false;
        }
    }

    resolve() {
        return new Promise((resolve, reject) => {
            const pagination: ParamsListTemplates = {
                page_size: 1000,
                last: ''
            };
            const load = (p: any, items: Template[] = []) => {
                return this._api.template.list(pagination).toPromise().then(data => {
                    items = items.concat(data.body as Template[]);
                    if (this.hasLast(data.headers, p)) {
                        return load(p, items);
                    } else {
                        return items;
                    }
                }).catch((err) => {
                    throw err;
                });
            };
            load(pagination).then((templates: Template[]) => {
                resolve(templates);
            }).catch((err) => {
                this._router.navigate([this._options.uiBaseUrl + '/error']);
                reject(err);
            });
        });
    }
}
