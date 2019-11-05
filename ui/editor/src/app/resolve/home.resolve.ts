import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ActivatedRouteSnapshot, RouterStateSnapshot, Resolve } from '@angular/router';

export class HomeObject {
  fetchTeam(id: string) {
    return 'RESOLVE ==> :) ' + id;
  }
}

@Injectable()
export class HomeResolve implements Resolve<string> {
  constructor(private homeObject: HomeObject) {
    
  }

  resolve(
    route: ActivatedRouteSnapshot,
    state: RouterStateSnapshot
  ): Observable<any>|Promise<any>|any {
    console.log("__________________XXXXXXXXXXXXXX_______________");
    //return this.homeObject.fetchTeam(route.params.id);
    return this.homeObject.fetchTeam(route.params.id);
  }
}