import { Injectable } from '@angular/core';
import { Observable, empty } from 'rxjs';
import { Router } from '@angular/router';
import { catchError } from 'rxjs/operators';
import { /*ActivatedRouteSnapshot, RouterStateSnapshot, */Resolve } from '@angular/router';
import { ApiService } from '../@services/api.service';

@Injectable()
export class MetaResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService, private router: Router) {
        this.api = api;
    }

    resolve() {
        // route: ActivatedRouteSnapshot,
        // state: RouterStateSnapshot
        // ): Observable<any> | Promise<any> | any {
        // // // // return this.homeObject.fetchTeam(route.params.id);
        return this.api.getMeta().pipe(
            catchError(err => {
                this.router.navigate(['/error']);
                return empty();
            })
        );
    }
}