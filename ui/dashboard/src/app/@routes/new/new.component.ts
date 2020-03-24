import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import { ApiService } from '../../@services/api.service';
import { HttpHeaders } from '@angular/common/http';
import Template from 'src/app/@models/template.model';

@Component({
  templateUrl: './new.html'
})
export class NewComponent implements OnInit {
  loaders: { [key: string]: boolean } = {};
  display: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  templates: Template[] = [];
  item: any = {};
  selectedTemplate: Template = null;
  Object = Object;

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.templates = _.orderBy(this.route.snapshot.data.templates, (t: any) => t.description.toLowerCase(), ['asc']);

    this.route.queryParams.subscribe((values) => {
      const tmp = _.find(this.templates, { name: values.template_name });
      this.selectedTemplate = tmp;
      this.newTask(tmp);
      Object.keys(this.item.input).forEach((inputName) => {
        if (values[inputName]) {
          this.item.input[inputName] = values[inputName];
        }
      });
    });
  }

  submit() {
    this.loaders.submit = true;

    return this.api.postTask(this.item).toPromise().then((data: any) => {
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
}
