import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../@services/api.service';
import * as _ from 'lodash';
import Template from 'src/app/@models/template.model';

@Component({
  templateUrl: './templates.html',
})
export class TemplatesComponent implements OnInit {
  templates: Template[];
  display: { [key: string]: boolean } = {};

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.templates = this.route.parent.snapshot.data.templates;
  }
}
