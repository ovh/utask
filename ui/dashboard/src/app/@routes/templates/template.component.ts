import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';

@Component({
  template: `
    <div class="main">
      <header>
        <h1>Template - {{templateName}}</h1>
      </header>
      <section>
        <utask-template-details *ngIf="templateName" [templateName]="templateName"></utask-template-details>
      </section>
    </div>
  `,
})
export class TemplateComponent implements OnInit {
  templateName: string;

  constructor(private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.templateName = params.templateName;
    });
  }
}
