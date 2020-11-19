import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import Template from 'projects/utask-lib/src/lib/@models/template.model';
import { ApiService } from 'projects/utask-lib/src/lib/@services/api.service';
import omit from 'lodash-es/omit';
import clone from 'lodash-es/clone';
import get from 'lodash-es/get';
import isArray from 'lodash-es/isArray';

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

  constructor(
    private api: ApiService,
    private activatedRoute: ActivatedRoute,
    private router: Router
  ) { }

  ngOnInit() {
    this.templates = this.activatedRoute.snapshot.data.templates.sort((a, b) => {
      return a.description.toLowerCase() < b.description.toLowerCase() ? -1 : 1;
    });

    this.activatedRoute.queryParams.subscribe((values) => {
      const template = this.templates.find(t => t.name === values.template_name);
      if (template) {
        if (!this.selectedTemplate || this.selectedTemplate.name !== template.name) {
          this.selectedTemplate = template;
          this.newTask(template);
        }
        get(template, 'inputs', []).forEach((input) => {
          if (input.collection && !isArray(values[input.name])) {
            this.item.input[input.name] = values[input.name] ? [values[input.name]] : [];
          } else if (input.type === 'number' && get(values, input.name)) {
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
    const passwordFieldsName = get(this.selectedTemplate, 'inputs', []).filter(t => t.type === 'password').map((e: any) => e.name);
    const inputs = omit(clone(this.item.input), passwordFieldsName);
    this.router.navigate(
      [],
      {
        relativeTo: this.activatedRoute,
        queryParams: {
          ...inputs,
          template_name: this.item.template_name
        },
        queryParamsHandling: 'merge',
      });
  }
}
