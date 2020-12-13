import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  template: `
    <div class="main">
      <header>
        <h1>Function - {{functionName}}</h1>
      </header>
      <section>
        <lib-utask-editor [ngModel]="JSON.stringify(function, null, 4)" ngDefaultControl [ngModelOptions]="{standalone: true}" [config]="{readonly: true, wordWrap: 'on'}"></lib-utask-editor>
      </section>
    </div>
  `,
})
export class FunctionComponent implements OnInit {
  functionName: string;
  function: Function;
  JSON = JSON;

  constructor(
    private _route: ActivatedRoute
  ) { }

  ngOnInit() {
    this._route.params.subscribe(params => {
      this.functionName = params.functionName;
      this.function = this._route.parent.snapshot.data.functions.find(f => f.name === this.functionName);
    });
  }
}