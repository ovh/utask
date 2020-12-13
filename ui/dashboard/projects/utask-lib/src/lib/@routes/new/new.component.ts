import { ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import get from 'lodash-es/get';
import { FormBuilder, FormControl, FormGroup, ValidatorFn, Validators } from '@angular/forms';
import isArray from 'lodash-es/isArray';
import Template from '../../@models/template.model';
import { ApiService, NewTask } from '../../@services/api.service';
import { InputsFormComponent } from '../../@components/inputs-form/inputs-form.component';

@Component({
  templateUrl: './new.html',
  styleUrls: ['./new.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class NewComponent implements OnInit {
  validateForm!: FormGroup;

  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  templates: Template[] = [];
  selectedTemplate: Template = null;
  inputControls: Array<string> = [];

  constructor(
    private _api: ApiService,
    private _activatedRoute: ActivatedRoute,
    private _router: Router,
    private _fb: FormBuilder,
    private _cd: ChangeDetectorRef
  ) { }

  ngOnInit() {
    this.validateForm = this._fb.group({
      template: [null, [Validators.required]],
      watchers: [null, []]
    });

    this.templates = this._activatedRoute.snapshot.data.templates.sort((a, b) => {
      return a.description.toLowerCase() < b.description.toLowerCase() ? -1 : 1;
    });

    this._activatedRoute.queryParams.subscribe((values) => {
      const template = this.templates.find(t => t.name === values.template_name);
      if (template) {
        if (!this.selectedTemplate || this.selectedTemplate.name !== template.name) {
          this.templateChange(template);
          const inputs = {
            template,
            watchers: values.watcher_usernames ?
              (isArray(values.watcher_usernames) ? values.watcher_usernames : [values.watcher_usernames])
              : []
          };
          get(template, 'inputs', []).forEach(input => {
            if (input.collection && !isArray(values[input.name])) {
              inputs['input_' + input.name] = values[input.name] ? [values[input.name]] : [];
            } else if (input.type === 'number' && get(values, input.name)) {
              inputs['input_' + input.name] = +values[input.name];
            } else if (input.type === 'bool') {
              inputs['input_' + input.name] = values[input.name] === 'true';
            } else if (input.type !== 'password') {
              inputs['input_' + input.name] = values[input.name];
            }
          });
          this.validateForm.patchValue(inputs, { emitEvent: false, onlySelf: true });
        }
      }
    });

    this.validateForm.valueChanges.subscribe(data => {
      const item = this.formValuesToNewTask(data);
      this.saveFormInQueryParams(item);
    });
  }

  templateChange(t: Template): void {
    this.inputControls.forEach(key => this.validateForm.removeControl(key));
    if (t) {
      t.inputs.forEach(input => {
        const validators: Array<ValidatorFn> = [];
        if (!input.optional) {
          validators.push(Validators.required);
        }
        this.validateForm.addControl('input_' + input.name, new FormControl(null, validators))
      });
      this.inputControls = t.inputs.map(input => 'input_' + input.name);
      this.selectedTemplate = t;
      this._cd.markForCheck();
    }
  }

  formValuesToNewTask(values: any): NewTask {
    const item = new NewTask();
    item.template_name = values.template.name;
    item.input = InputsFormComponent.getInputs(values);
    item.watcher_usernames = values.watchers ? values.watchers : [];
    return item;
  }

  submitForm(): void {
    for (const i in this.validateForm.controls) {
      if (Object.prototype.hasOwnProperty.call(this.validateForm.controls, i)) {
        this.validateForm.controls[i].markAsDirty();
        this.validateForm.controls[i].updateValueAndValidity();
      }
    }
    if (this.validateForm.invalid) {
      return;
    }

    const item = this.formValuesToNewTask(this.validateForm.value);

    this.loaders.submit = true;
    this._cd.markForCheck();
    this._api.task.add(item).toPromise().then((data: any) => {
      this._router.navigate([`/task/${data.id}`]);
    }).catch((err) => {
      this.errors.submit = err;
    }).finally(() => {
      this.loaders.submit = false;
      this._cd.markForCheck();
    });
  }

  saveFormInQueryParams(item: NewTask) {
    const passwordFieldsName = get(this.selectedTemplate, 'inputs', [])
      .filter(t => t.type === 'password').map((e: any) => e.name);

    const queryParams = {
      template_name: item.template_name,
      watcher_usernames: item.watcher_usernames
    };

    Object.keys(item.input)
      .filter(k => passwordFieldsName.indexOf(k) === -1)
      .forEach(k => {
        queryParams[k] = item.input[k];
      });

    this._router.navigate([], {
      relativeTo: this._activatedRoute,
      queryParams,
      replaceUrl: true
    });
  }
}
