import { Injectable } from '@angular/core';
import { Router } from '@angular/router';
import { catchError } from 'rxjs/operators';
import { Resolve } from '@angular/router';
import { ApiService } from 'projects/utask-lib/src/lib/@services/api.service';

@Injectable()
export class MetaResolve implements Resolve<any> {
    constructor(
        private _api: ApiService,
        private _router: Router
    ) { }

    resolve() {
        return this._api.meta.get().pipe(
            catchError(err => {
                this._router.navigate(['/error']);
                throw err;
            })
        );
    }
}