import { Injectable } from '@angular/core';
import { Resolve, Router, ActivatedRouteSnapshot } from '@angular/router';
import { ApiService } from 'src/app/@services/api.service';
import { Stats } from 'src/app/@models/stats.model';

@Injectable()
export class StatsResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService, private router: Router) {
        this.api = api;
    }

    resolve(route: ActivatedRouteSnapshot) {
        return this.api.getStats().toPromise().then((res: any) => {
            return res as Stats;
        }).catch(() => {
            return {};
        });
    }
}