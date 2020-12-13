import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { NzTableComponent } from 'ng-zorro-antd/table';
import Template from '../../@models/template.model';
import { UTaskLibOptions } from '../../@services/api.service';

@Component({
  templateUrl: './templates.html',
  styleUrls: ['./templates.sass']
})
export class TemplatesComponent implements OnInit {
  @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Template>;
  templates: Template[];
  display: { [key: string]: boolean } = {};
  uiBaseUrl: string;

  constructor(
    private _route: ActivatedRoute,
    private _options: UTaskLibOptions
  ) {
    this.uiBaseUrl = this._options.uiBaseUrl;
  }

  ngOnInit() {
    this.templates = this._route.parent.snapshot.data.templates;
  }
}
