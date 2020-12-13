import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { NzTableComponent } from 'ng-zorro-antd/table';
import Template from '../../@models/template.model';

@Component({
  templateUrl: './templates.html',
  styleUrls: ['./templates.sass']
})
export class TemplatesComponent implements OnInit {
  templates: Template[];
  display: { [key: string]: boolean } = {};
  @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Template>;

  constructor(
    private _route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.templates = this._route.parent.snapshot.data.templates;
  }
}
