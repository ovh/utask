import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router, NavigationEnd, Event } from '@angular/router';
import get from 'lodash-es/get';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';

@Component({
  templateUrl: './base.html',
})
export class BaseComponent implements OnInit {
  meta: Meta;
  constructor(private activedRoute: ActivatedRoute, private router: Router) {
    this.router.events.subscribe((event: Event) => {
      if (event instanceof NavigationEnd) {
        window.scroll(0, 0);
        // Navigation Service - Title & history
        let route = this.activedRoute;
        while (route.firstChild) {
          route = route.firstChild;
        }
        route.data.subscribe((values) => {
          // Title
          if (typeof values.title === 'string') {
            document.title = values.title;
          } else if (values.title) {
            let title = '';
            const args = values.title.args.map((s: string) => {
              return get(values, s);
            });
            title = this.format(values.title.value, ...args);
            document.title = title;
          }
        });
      }
    });
  }

  ngOnInit() {
    this.meta = this.activedRoute.snapshot.data.meta;
  }

  private format(str: string, ...args) {
    return str.replace(/{(\d+)}/g, function (match, number) {
      return typeof args[number] != 'undefined'
        ? args[number]
        : match
        ;
    });
  }
}
