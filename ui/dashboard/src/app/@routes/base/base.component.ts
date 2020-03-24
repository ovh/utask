import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router, NavigationStart, NavigationEnd, NavigationError, Event } from '@angular/router';
import { Subscription, observable } from 'rxjs';
import * as _ from 'lodash';

@Component({
  templateUrl: './base.html',
})
export class BaseComponent implements OnInit {
  meta: any;
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
              return _.get(values, s);
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
