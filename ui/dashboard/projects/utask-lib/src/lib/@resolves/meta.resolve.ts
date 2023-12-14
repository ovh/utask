import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { catchError } from 'rxjs/operators';

import { ApiService, UTaskLibOptions } from '../@services/api.service';

@Injectable()
export class MetaResolve  {
    constructor(
        private _api: ApiService,
        private _router: Router,
        private _options: UTaskLibOptions
    ) { }

    resolve() {
        return this._api.meta.get().pipe(
            catchError(err => {
                this._router.navigate([this._options.uiBaseUrl + '/error']);
                throw err;
            })
        );
    }
}