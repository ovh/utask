import { Component, OnInit } from '@angular/core';
import { Title } from '@angular/platform-browser';
import { ActivatedRoute, Router, NavigationEnd } from '@angular/router';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';
import { filter, map, mergeMap } from 'rxjs/operators';
import { ThemeService } from 'src/app/@services/theme.service';

@Component({
    standalone: false,
  templateUrl: './base.html',
  styleUrls: ['./base.scss']
})
export class BaseComponent implements OnInit {
  meta: Meta;
  darkThemeActive: boolean;

  constructor(
    private _activatedRoute: ActivatedRoute,
    private _router: Router,
    private _titleService: Title,
    private _themeService: ThemeService
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
    const theme = this._themeService.getTheme();
    this.darkThemeActive = theme && theme === 'dark';
  }

  switchDarkTheme(active: boolean): void {
    if (active) {
      this._themeService.changeTheme('dark');
    } else {
      this._themeService.changeTheme('default');
    }
  }
}
