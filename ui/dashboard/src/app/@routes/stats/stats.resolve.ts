import { Injectable } from '@angular/core';
import { Resolve, Router, ActivatedRouteSnapshot } from '@angular/router';
import { Stats } from 'fs';
import { ApiService } from 'projects/utask-lib/src/lib/@services/api.service';

@Injectable()
export class StatsResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService) {
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