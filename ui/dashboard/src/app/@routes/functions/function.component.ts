import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import Function from 'utask-lib/@models/function.model';

@Component({
  template: `
    <div class="main">
      <header>
        <h1>Function - {{functionName}}</h1>
      </header>
      <section>
        <utask-editor [value]="JSON.stringify(function, null, 4)" [config]="{readonly: true}"></utask-editor>
      </section>
    </div>
  `,
})
export class FunctionComponent implements OnInit {
  functionName: string;
  function: Function;
  JSON = JSON;

  constructor(private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.functionName = params.functionName;
      this.function = _.find(this.route.parent.snapshot.data.functions, { name: this.functionName });
    });
  }
}