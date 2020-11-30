import { Component, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { ActivatedRoute, Router, NavigationEnd } from '@angular/router';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';
import { filter, map, mergeMap } from 'rxjs/operators';

@Component({
  templateUrl: './base.html',
  styleUrls: ['./base.scss'],
})
export class BaseComponent implements OnInit {
  meta: Meta;

  constructor(
    private _activatedRoute: ActivatedRoute,
    private _router: Router,
    private _titleService: Title
  ) {
    this._router.events
      .pipe(filter(event => event instanceof NavigationEnd))
      .pipe(map(() => this._activatedRoute))
      .pipe(map((route) => {
        while (route.firstChild) {
            route = route.firstChild;
        }
        return route;
    }))
      .pipe(filter(route => route.outlet === 'primary'))
      .pipe(mergeMap(route => route.data))
      .subscribe((routeData) => {
        this._titleService.setTitle((routeData.title ? routeData.title + ' - ' : '') + routeData.meta.application_name);
      });
  }

  ngOnInit() {
    this.meta = this._activatedRoute.snapshot.data.meta;
  }
}
