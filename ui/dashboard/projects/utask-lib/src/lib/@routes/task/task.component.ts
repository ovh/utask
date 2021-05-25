import { Component, OnInit, OnDestroy, ViewContainerRef, ChangeDetectionStrategy, ChangeDetectorRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import get from 'lodash-es/get';
import { FormBuilder, FormControl, FormGroup, ValidatorFn, Validators } from '@angular/forms';
import { NzModalService } from 'ng-zorro-antd/modal';
import { interval, Subscription } from 'rxjs';
import { concatMap, filter } from 'rxjs/operators';
import { NzNotificationService } from 'ng-zorro-antd/notification';
import { ApiService, UTaskLibOptions } from '../../@services/api.service';
import { ResolutionService } from '../../@services/resolution.service';
import { RequestService } from '../../@services/request.service';
import { TaskService } from '../../@services/task.service';
import Template from '../../@models/template.model';
import Meta from '../../@models/meta.model';
import Task, { Comment, ResolverInput } from '../../@models/task.model';
import { ModalApiYamlComponent } from '../../@modals/modal-api-yaml/modal-api-yaml.component';
import { InputsFormComponent } from '../../@components/inputs-form/inputs-form.component';
import { TasksListComponentOptions } from '../../@components/tasks-list/tasks-list.component';

@Component({
  selector: 'lib-utask-task',
  templateUrl: './task.html',
  styleUrls: ['./task.sass']
})
export class TaskComponent implements OnInit, OnDestroy {
  validateResolveForm!: FormGroup;
  validateRejectForm!: FormGroup;
  inputControls: Array<string> = [];

  objectKeys = Object.keys;
  loaders: { [key: string]: boolean } = {};
  haveAtLeastOneChilTask = false;
  errors: { [key: string]: any } = {};
  display: { [key: string]: boolean } = {};
  confirm: { [key: string]: boolean } = {};
  refreshes: { [key: string]: Subscription } = {};
  textarea: { [key: string]: boolean } = {};
  item: any = {
    resolver_inputs: {},
    task_id: null
  };
  task: Task = null;
  taskIsResolvable = false;
  taskId = '';
  resolution: any = null;
  selectedStep = '';
  meta: Meta = null;
  resolverInputs: Array<ResolverInput> = [];

  JSON = JSON;
  template: Template;
  comment: any = {
    content: ''
  };
  autorefresh: any = {
    hasChanged: false,
    enable: false,
    actif: false
  };
  uiBaseUrl: string;
  listOptions = new TasksListComponentOptions();

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private resolutionService: ResolutionService,
    private requestService: RequestService,
    private taskService: TaskService,
    private router: Router,
    private _fb: FormBuilder,
    private modal: NzModalService,
    private viewContainerRef: ViewContainerRef,
    private _notif: NzNotificationService,
    private _options: UTaskLibOptions
  ) {
    this.uiBaseUrl = this._options.uiBaseUrl;
    this.listOptions.routingTaskPath = this._options.uiBaseUrl + '/task/';
  }

  ngOnDestroy() {
    if (this.refreshes.tasks) {
      this.refreshes.tasks.unsubscribe();
    }
  }

  ngOnInit() {
    this.validateResolveForm = this._fb.group({});
    this.validateRejectForm = this._fb.group({
      agree: [false, [Validators.requiredTrue]]
    });

    this.meta = this.route.parent.snapshot.data.meta;
    this.route.params.subscribe(params => {
      this.errors.main = null;
      this.taskId = params.id;
      this.loadTask().then(() => {
        this.display.request = (!this.task.result && !this.resolution) || (!this.resolution && this.taskIsResolvable);
        this.display.result = this.task.state === 'DONE';
        this.display.execution = !!this.resolution;
        this.display.reject = !this.resolution && this.taskIsResolvable;
        this.display.resolution = !this.resolution && this.taskIsResolvable;
        this.display.comments = this.task.comments && this.task.comments.length > 0;
      }).catch((err) => {
        console.log(err);
        if (!this.task || this.task.id !== params.id) {
          this.errors.main = err;
        }
      });
    });

    this.refreshes.tasks = interval(this._options.refresh.task)
      .pipe(filter(() => {
        return !this.loaders.task && this.autorefresh.actif;
      }))
      .pipe(concatMap(() => this.loadTask()))
      .subscribe();
  }

  addComment() {
    this.loaders.addComment = true;
    this.api.task.comment.add(this.task.id, this.comment.content).toPromise().then((comment: Comment) => {
      this.task.comments = get(this.task, 'comments', []);
      this.task.comments.push(comment);
      this.errors.addComment = null;
      this.comment.content = '';
    }).catch((err) => {
      this.errors.addComment = err;
    }).finally(() => {
      this.loaders.addComment = false;
    });
  }

  previewTask() {
    this.modal.create({
      nzTitle: 'Request preview',
      nzContent: ModalApiYamlComponent,
      nzWidth: '80%',
      nzViewContainerRef: this.viewContainerRef,
      nzComponentParams: {
        apiCall: () => this.api.task.getAsYaml(this.taskId).toPromise()
      },
    });
  }

  previewResolution() {
    this.modal.create({
      nzTitle: 'Resolution preview',
      nzContent: ModalApiYamlComponent,
      nzWidth: '80%',
      nzViewContainerRef: this.viewContainerRef,
      nzComponentParams: {
        apiCall: () => this.api.resolution.getAsYaml(this.resolution.id).toPromise()
      },
    });
  }

  editRequest(task: Task) {
    this.requestService.edit(task).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The request has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  editResolution(resolution: any) {
    this.resolutionService.edit(resolution).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The resolution has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  runResolution(resolution: any) {
    this.resolutionService.run(resolution.id).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The resolution has been run.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  pauseResolution(resolution: any) {
    this.resolutionService.pause(resolution.id).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The resolution has been paused.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  cancelResolution(resolution: any) {
    this.resolutionService.cancel(resolution.id).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The resolution has been cancelled.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  extendResolution(resolution: any) {
    this.resolutionService.extend(resolution.id).then((data: any) => {
      this.loadTask();
      this._notif.info('', 'The resolution has been extended.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  deleteTask(taskId: string) {
    this.taskService.delete(taskId).then((data: any) => {
      this.router.navigate([this._options.uiBaseUrl + '/']);
      this._notif.info('', 'The task has been deleted.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  rejectTask() {
    for (const i in this.validateRejectForm.controls) {
      if (Object.prototype.hasOwnProperty.call(this.validateResolveForm.controls, i)) {
        this.validateRejectForm.controls[i].markAsDirty();
        this.validateRejectForm.controls[i].updateValueAndValidity();
      }
    }
    if (this.validateRejectForm.invalid) {
      return;
    }

    this.loaders.rejectTask = true;
    this.api.task.reject(this.task.id).toPromise().then((res: any) => {
      this.errors.rejectTask = null;
      this.loadTask();
    }).catch((err) => {
      this.errors.rejectTask = err;
    }).finally(() => {
      this.loaders.rejectTask = false;
    });
  }

  resolveTask() {
    for (const i in this.validateResolveForm.controls) {
      if (Object.prototype.hasOwnProperty.call(this.validateResolveForm.controls, i)) {
        this.validateResolveForm.controls[i].markAsDirty();
        this.validateResolveForm.controls[i].updateValueAndValidity();
      }
    }
    if (this.validateResolveForm.invalid) {
      return;
    }

    this.loaders.resolveTask = true;
    this.api.resolution.add({
      ...this.item,
      resolver_inputs: InputsFormComponent.getInputs(this.validateResolveForm.value)
    }).toPromise().then((res: any) => {
      this.errors.resolveTask = null;
      this.loadTask();
    }).catch((err) => {
      this.errors.resolveTask = err;
    }).finally(() => {
      this.loaders.resolveTask = false;
    });
  }

  selectStepFromViewer(step) {
    this.selectedStep = step || '';
  }

  async getTemplate(templateName): Promise<Template> {
    return new Promise((resolve, reject) => {
      if (this.template?.name === templateName) {
        resolve(this.template);
      } else {
        const template = this.route.parent.snapshot.data.templates.find(t => t.name === templateName);
        if (template) {
          resolve(template);
        } else {
          this.api.template.get(templateName).toPromise().then((t) => {
            resolve(t);
          }).catch((err) => {
            reject(err);
          });
        }
      }
    });
  }

  async templateChange(templateName: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.getTemplate(templateName).then((template) => {
        this.template = template;
        this.resolverInputs = this.template.resolver_inputs;

        this.inputControls.forEach(key => this.validateResolveForm.removeControl(key));
        if (this.resolverInputs) {
          this.resolverInputs.forEach(input => {
            const validators: Array<ValidatorFn> = [];
            if (!input.optional && input.type !== 'bool') {
              validators.push(Validators.required);
            }
            let defaultValue: any;
            if (input.type === 'bool') {
              defaultValue = !!input.default;
            } else {
              defaultValue = input.default;
            }
            this.validateResolveForm.addControl('input_' + input.name, new FormControl(defaultValue, validators));
          });
          this.inputControls = this.resolverInputs.map(input => 'input_' + input.name);
        }
        resolve();
      }).catch((err) => {
        reject(err);
      });
    });
  }

  loadTask() {
    return new Promise<void>((resolve, reject) => {
      this.loaders.task = true;
      Promise.all([
        this.api.task.get(this.taskId).toPromise(),
        this.api.task.list({
          page_size: 10,
          type: this.meta.user_is_admin ? 'all' : 'own',
          tag: '_utask_parent_task_id=' + this.taskId
        } as any).toPromise(),
      ]).then(async (data) => {
        this.task = data[0];
        this.haveAtLeastOneChilTask = data[1].body.length > 0;
        this.task.comments = get(this.task, 'comments', []).sort((a, b) => a.created < b.created ? -1 : 1);
        this.item.task_id = this.task.id;
        await this.templateChange(this.task.template_name);
        this.taskIsResolvable = this.requestService.isResolvable(this.task, this.meta, this.template);
        if (['DONE', 'WONTFIX', 'CANCELLED'].indexOf(this.task.state) > -1) {
          this.autorefresh.enable = false;
          this.autorefresh.actif = false;
        } else {
          this.autorefresh.enable = true;
          if (!this.autorefresh.hasChanged) {
            this.autorefresh.actif = ['TODO', 'RUNNING', 'TO_AUTORUN'].indexOf(this.task.state) > -1;
          }
        }

        if (this.task.resolution) {
          this.loadResolution(this.task.resolution).then(rData => {
            if (!this.resolution && rData) {
              this.display.execution = true;
              this.display.request = false;
            }
            this.resolution = rData;
            resolve();
          }).catch((err) => {
            reject(err);
          }).finally(() => {
            this.loaders.resolution = false;
          });
        } else {
          this.resolution = null;
          resolve();
        }
      }).catch((err: any) => {
        console.log(err);
        reject(err);
      }).finally(() => {
        this.loaders.task = false;
      });
    });
  }

  loadResolution(resolutionId: string): any {
    return new Promise((resolve, reject) => {
      this.loaders.resolution = true;
      this.api.resolution.get(resolutionId).subscribe(data => {
        resolve(data);
      }, (err: any) => {
        reject(err);
      });
    });
  }

  eventUtask(event: any) {
    this._notif.create(event.type, '', event.message);
  }
}
