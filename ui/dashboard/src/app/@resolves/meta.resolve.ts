import { Injectable } from '@angular/core';
import { Observable, empty } from 'rxjs';
import { Router } from '@angular/router';
import { catchError } from 'rxjs/operators';
import { /*ActivatedRouteSnapshot, RouterStateSnapshot, */Resolve } from '@angular/router';
import { ApiService } from 'utask-lib';

@Injectable()
export class MetaResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService, private router: Router) {
        this.api = api;
    }

    resolve() {
        return this.api.meta.get().pipe(
            catchError(err => {
                this.router.navigate(['/error']);
                return empty();
            })
        );
    }
}