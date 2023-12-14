import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot } from '@angular/router';
import { Stats } from '../@models/task.model';
import { ApiService } from '../@services/api.service';

@Injectable()
export class StatsResolve  {
    constructor(
        private _api: ApiService
    ) { }

    resolve(route: ActivatedRouteSnapshot) {
        return this._api.stats.get().toPromise().then((res: any) => {
            return res as Stats;
        }).catch(() => {
            return {};
        });
    }
}