import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import Template from 'projects/utask-lib/src/lib/@models/template.model';

@Component({
  templateUrl: './templates.html',
})
export class TemplatesComponent implements OnInit {
  templates: Template[];
  display: { [key: string]: boolean } = {};

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.templates = this.route.parent.snapshot.data.templates;
  }
}
