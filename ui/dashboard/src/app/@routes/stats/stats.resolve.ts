import { Injectable } from '@angular/core';
import { Resolve, Router, ActivatedRouteSnapshot } from '@angular/router';
import { ApiService } from 'utask-lib';
import { Stats } from 'fs';

@Injectable()
export class StatsResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService, private router: Router) {
        this.api = api;
    }

    resolve(route: ActivatedRouteSnapshot) {
        return this.api.stats.get().toPromise().then((res: any) => {
            return res as Stats;
        }).catch(() => {
            return {};
        });
    }
}