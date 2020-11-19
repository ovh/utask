import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  template: `
    <div class="main">
      <header>
        <h1>Function - {{functionName}}</h1>
      </header>
      <section>
        <lib-utask-editor [value]="JSON.stringify(function, null, 4)" [config]="{readonly: true}"></lib-utask-editor>
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
      this.function = this.route.parent.snapshot.data.functions.find(f => f.name === this.functionName);
    });
  }
}