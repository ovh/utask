import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import Template from 'projects/utask-lib/src/lib/@models/template.model';
import { NzTableComponent } from 'ng-zorro-antd/table';

@Component({
  templateUrl: './templates.html',
  styleUrls: ['./templates.sass'],
})
export class TemplatesComponent implements OnInit {
  templates: Template[];
  display: { [key: string]: boolean } = {};
  @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Template>;

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.templates = this.route.parent.snapshot.data.templates;
  }
}
