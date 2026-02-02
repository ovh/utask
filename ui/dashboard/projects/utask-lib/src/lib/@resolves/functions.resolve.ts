import { Injectable } from '@angular/core';
import { Router } from '@angular/router';

import { HttpHeaders } from '@angular/common/http';
import { ApiService, ParamsListFunctions, UTaskLibOptions } from '../@services/api.service';
import { Function } from '../@models/function.model';

@Injectable()
export class FunctionsResolve  {
    constructor(
        private _api: ApiService,
        private _router: Router,
        private _options: UTaskLibOptions
    ) { }

    hasLast(headers: HttpHeaders): string {
        const link = headers.get('link');
        if (!link) {
            return null;
        }
        const match = link.match(/last=([^&;\s>]+)/);
        if (!match) {
            return null;
        }
        return match[1];
    }

    async resolve() {
        const pagination: ParamsListFunctions = {
            page_size: 1000
        };

        // Load first page
        let items: Array<Function>;
        let res = await this._api.function.list(pagination).toPromise().catch((err) => {
            this._router.navigate([this._options.uiBaseUrl + '/error']);
            throw err;
        });
        pagination.last = this.hasLast(res.headers);
        items = res.body;

        // Load more page if needed
        while (pagination.last) {
            res = await this._api.function.list(pagination).toPromise().catch((err) => {
                this._router.navigate([this._options.uiBaseUrl + '/error']);
                throw err;
            });
            pagination.last = this.hasLast(res.headers);
            items.push(...res.body);
        }

        return items;
    }
}
