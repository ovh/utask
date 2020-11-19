import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  template: `
    <div class="main">
      <header>
        <h1>Template - {{templateName}}</h1>
      </header>
      <section>
        <app-template-details *ngIf="templateName" [templateName]="templateName"></app-template-details>
      </section>
    </div>
  `,
})
export class TemplateComponent implements OnInit {
  templateName: string;

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.templateName = params.templateName;
    });
  }
}
