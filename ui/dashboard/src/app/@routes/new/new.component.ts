import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import { HttpHeaders } from '@angular/common/http';
import { ApiService } from 'utask-lib';
import Template from 'utask-lib/@models/template.model';

@Component({
  templateUrl: './new.html'
})
export class NewComponent implements OnInit {
  loaders: { [key: string]: boolean } = {};
  display: { [key: string]: boolean } = {};
  textarea: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  templates: Template[] = [];
  item: any = {};
  selectedTemplate: Template = null;
  Object = Object;

  constructor(private api: ApiService, private activatedRoute: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.templates = _.orderBy(this.activatedRoute.snapshot.data.templates, (t: any) => t.description.toLowerCase(), ['asc']);

    this.activatedRoute.queryParams.subscribe((values) => {
      const template = _.find(this.templates, { name: values.template_name });
      if (template) {
        if (!this.selectedTemplate || this.selectedTemplate.name !== template.name) {
          this.selectedTemplate = template;
          this.newTask(template);
        }
        _.get(template, 'inputs', []).forEach((input) => {
          if (input.collection && !_.isArray(values[input.name])) {
            this.item.input[input.name] = values[input.name] ? [values[input.name]] : [];
          } else if (input.type === 'number' && _.get(values, input.name)) {
            this.item.input[input.name] = +values[input.name];
          } else if (input.type === 'bool') {
            this.item.input[input.name] = values[input.name] === 'true';
          } else if (input.type !== 'password') {
            this.item.input[input.name] = values[input.name];
          }
        })
      }
    });
  }

  submit() {
    this.loaders.submit = true;

    return this.api.task.add(this.item).toPromise().then((data: any) => {
      this.errors.submit = null;
      this.router.navigate([`/task/${data.id}`]);
    }).catch((err) => {
      this.errors.submit = err;
    }).finally(() => {
      this.loaders.submit = false;
    });
  }

  newTask(template: Template) {
    this.item.template_name = template.name;
    this.item.input = Object.assign({}, ...(template.inputs || []).map((i: any) => {
      const o = {};
      o[i.name] = i.default;
      return o;
    }));
  }

  saveFormInQueryParams() {
    const passwordFieldsName = _.filter(_.get(this.selectedTemplate, 'inputs', []), { type: 'password' }).map((e: any) => e.name);
    let inputs = _.omit(_.clone(this.item.input), passwordFieldsName);
    this.router.navigate(
      [],
      {
        relativeTo: this.activatedRoute,
        queryParams: _.merge({
          template_name: this.item.template_name
        }, inputs),
        queryParamsHandling: 'merge',
      });
  }
}
