import { Component, OnInit, OnDestroy, ViewContainerRef } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import get from 'lodash-es/get';
import { FormBuilder, FormControl, FormGroup, ValidatorFn, Validators } from '@angular/forms';
import { NzModalService } from 'ng-zorro-antd/modal';
import { BehaviorSubject, combineLatest, interval, of, Subscription } from 'rxjs';
import { concatMap, filter, map, switchMap } from 'rxjs/operators';
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
import Resolution from '../../@models/resolution.model';

@Component({
  selector: 'lib-utask-task',
  templateUrl: './task.html',
  styleUrls: ['./task.sass']
})
export class TaskComponent implements OnInit, OnDestroy {
  private _meta$ = new BehaviorSubject<Meta | null>(null);
  readonly meta$ = this._meta$.asObservable();

  private _task$ = new BehaviorSubject<Task | null>(null);
  readonly task$ = this._task$.asObservable();

  readonly template$ = this.task$.pipe(switchMap(task => {
    if (!task) {
      return of(null);
    }
    return this.api.template.get(task.template_name)
  }));

  private _resolution$ = new BehaviorSubject<Resolution | null>(null);
  readonly resolution$ = this._resolution$.asObservable();

  private _isResolver$ = combineLatest([
    this.meta$, this.template$, this.task$, this.resolution$
  ]).pipe(map(([meta, template, task, resolution]) => {
    if (meta?.user_is_admin) {
      return true;
    }

    // check if the current user is declared as resolver in the task template
    if (meta && (template?.allowed_resolver_usernames ?? []).includes(meta.username)) {
      return true;
    }

    // check if the current user has at least one group declared as resolver in the task template
    if (meta && (template?.allowed_resolver_groups ?? []).some(v => meta.user_groups.includes(v))) {
      return true;
    }

    // check if the current user is declared as resolver in the task
    if (meta && (task?.resolver_usernames ?? []).includes(meta.username)) {
      return true;
    }

    // check if the current user has at least one group declared as resolver in the task
    if (meta && (task?.resolver_groups ?? []).some(v => meta.user_groups.includes(v))) {
      return true;
    }

    // check if the current user is the resolution resolver
    if (meta && resolution && meta.username === resolution.resolver_username) {
      return true;
    }

    return false;
  }))

  readonly canStartOver$ = combineLatest([
    this.meta$, this.template$, this.resolution$, this._isResolver$
  ]).pipe(map(([meta, template, resolution, isResolver]) => {
    if (!['PAUSED', 'CANCELLED', 'BLOCKED_BADREQUEST'].includes(resolution?.state)) {
      return false;
    }

    if (meta?.user_is_admin) {
      return true;
    }

    if (!template?.allow_task_start_over) {
      return false;
    }

    return isResolver;
  }));

  readonly canRun$ = combineLatest([this.resolution$, this._isResolver$]).pipe(map(([resolution, isResolver]) => {
    if (['CANCELLED', 'RUNNING', 'DONE'].includes(resolution?.state)) {
      return false;
    }

    return isResolver;
  }));

  readonly canPause$ = combineLatest([this.resolution$, this._isResolver$]).pipe(map(([resolution, isResolver]) => {
    if (['CANCELLED', 'RUNNING', 'DONE', 'PAUSED'].includes(resolution?.state)) {
      return false;
    }

    return isResolver;
  }));

  readonly canExtend$ = combineLatest([this.resolution$, this._isResolver$]).pipe(map(([resolution, isResolver]) => {
    if (resolution?.state !== 'BLOCKED_MAXRETRIES') {
      return false;
    }

    return isResolver;
  }));

  readonly canCancel$ = combineLatest([this.resolution$, this._isResolver$]).pipe(map(([resolution, isResolver]) => {
    if (['CANCELLED', 'RUNNING', 'DONE'].includes(resolution?.state)) {
      return false;
    }

    return isResolver;
  }));

  readonly canEdit$ = combineLatest([this.resolution$, this.meta$]).pipe(map(([resolution, meta]) => {
    if (resolution?.state !== 'PAUSED') {
      return false;
    }

    return !!meta?.user_is_admin;
  }));

  readonly canEditRequest$ = combineLatest([this.resolution$, this._isResolver$]).pipe(map(([resolution, isResolver]) => {
    if (!['TODO', 'PAUSED'].includes(resolution?.state)) {
      return false;
    }

    return isResolver;
  }));;

  readonly canDeleteRequest$ = combineLatest([this.task$, this.meta$]).pipe(map(([task, meta]) => {
    if (['RUNNING', 'BLOCKED'].includes(task?.state)) {
      return false;
    }

    return !!meta?.user_is_admin;
  }));

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
  taskJson: string;
  taskIsResolvable = false;
  taskId = '';
  selectedStep = '';
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

    this._meta$.next(this.route.parent.snapshot.data.meta)
    this.route.params.subscribe(params => {
      this.errors.main = null;
      this.taskId = params.id;
      this.loadTask().then(() => {
        this.display.request = (!this._task$.value.result && !this._resolution$.value) || (!this._resolution$.value && this.taskIsResolvable);
        this.display.result = this._task$.value.state === 'DONE';
        this.display.execution = !!this._resolution$.value;
        this.display.reject = !this._resolution$.value && this.taskIsResolvable;
        this.display.resolution = !this._resolution$.value && this.taskIsResolvable;
        this.display.comments = this._task$.value.comments && this._task$.value.comments.length > 0;
      }).catch((err) => {
        console.log(err);
        if (!this._task$.value || this._task$.value.id !== params.id) {
          this.errors.main = err;
        }
      });
    });

    this.refreshes.tasks = interval(this._options.refresh.task)
      .pipe(filter(() => {
        return !this.loaders.task && !this.loaders.refreshTask && this.autorefresh.actif;
      }))
      .pipe(concatMap(() => this.loadTask(true)))
      .subscribe();
  }

  addComment() {
    this.loaders.addComment = true;
    this.api.task.comment.add(this._task$.value.id, this.comment.content).toPromise().then((comment: Comment) => {
      this._task$.value.comments = get(this._task$.value, 'comments', []);
      this._task$.value.comments.push(comment);
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
        apiCall: () => this.api.resolution.getAsYaml(this._resolution$.value.id).toPromise()
      },
    });
  }

  editRequest() {
    this.requestService.edit(this._task$.value).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The request has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  editResolution(resolution: any) {
    this.resolutionService.edit(resolution).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The resolution has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  runResolution(resolution: any) {
    this.resolutionService.run(resolution.id).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The resolution has been run.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  pauseResolution(resolution: any) {
    this.resolutionService.pause(resolution.id).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The resolution has been paused.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  cancelResolution(resolution: any) {
    this.resolutionService.cancel(resolution.id).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The resolution has been cancelled.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  extendResolution(resolution: any) {
    this.resolutionService.extend(resolution.id).then((data: any) => {
      this.loadTask(true);
      this._notif.info('', 'The resolution has been extended.');
    }).catch((err) => {
      if (err !== 'close') {
        this._notif.error('', get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  deleteTask() {
    this.taskService.delete(this._task$.value.id).then((data: any) => {
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
    this.api.task.reject(this._task$.value.id).toPromise().then((res: any) => {
      this.errors.rejectTask = null;
      this.loadTask(true);
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
      this.loadTask(true);
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
      const template = this.route.parent.snapshot.data.templates.find((t: Template) => t.name === templateName);
      if (template) {
        resolve(template);
      } else {
        this.api.template.get(templateName).toPromise().then((t) => {
          resolve(t);
        }).catch((err) => {
          reject(err);
        });
      }
    });
  }

  taskChanged() {
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
    this.item.task_id = this._task$.value.id;
  }

  loadTask(refresh: boolean = false) {
    return new Promise<void>((resolve, reject) => {
      this.loaders.task = !refresh;
      this.loaders.refreshTask = refresh;
      Promise.all([
        this.api.task.get(this.taskId).toPromise(),
        this.api.task.list({
          page_size: 10,
          type: this._meta$.value.user_is_admin ? 'all' : 'own',
          tag: '_utask_parent_task_id=' + this.taskId
        } as any).toPromise(),
      ]).then(async (data) => {
        this._task$.next(data[0]);
        this.taskJson = JSON.stringify(data[0].result, null, 2);
        this.haveAtLeastOneChilTask = data[1].body.length > 0;
        this._task$.value.comments = get(this._task$.value, 'comments', []).sort((a, b) => a.created < b.created ? -1 : 1);

        if (this.template?.name !== this._task$.value.template_name) {
          try {
            this.template = await this.getTemplate(this._task$.value.template_name);
            this.resolverInputs = this.template.resolver_inputs;
          } catch (err) {
            reject(err);
          }
        }

        if (!refresh) {
          this.taskChanged();
        }

        this.taskIsResolvable = this.requestService.isResolvable(this._task$.value, this._meta$.value, this.template);
        if (['DONE', 'WONTFIX', 'CANCELLED'].indexOf(this._task$.value.state) > -1) {
          this.autorefresh.enable = false;
          this.autorefresh.actif = false;
        } else {
          this.autorefresh.enable = true;
          if (!this.autorefresh.hasChanged) {
            this.autorefresh.actif = ['TODO', 'RUNNING', 'TO_AUTORUN', 'WAITING'].indexOf(this._task$.value.state) > -1;
          }
        }

        if (this._task$.value.resolution) {
          this.loaders.resolution = !refresh;
          this.loaders.refreshResolution = !refresh;
          this.loadResolution(this._task$.value.resolution).then(rData => {
            if (!this._resolution$.value && rData) {
              this.display.execution = true;
              this.display.request = false;
            }
            this._resolution$.next(rData);
            resolve();
          }).catch((err) => {
            reject(err);
          }).finally(() => {
            this.loaders.resolution = false;
            this.loaders.refreshResolution = false;
          });
        } else {
          this._resolution$.next(null);
          resolve();
        }
      }).catch((err: any) => {
        console.log(err);
        reject(err);
      }).finally(() => {
        this.loaders.task = false;
        this.loaders.refreshTask = false;
      });
    });
  }

  loadResolution(resolutionId: string): any {
    return new Promise((resolve, reject) => {
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

  restartTask() {
    for (const i in this.validateResolveForm.controls) {
      if (Object.prototype.hasOwnProperty.call(this.validateResolveForm.controls, i)) {
        this.validateResolveForm.controls[i].markAsDirty();
        this.validateResolveForm.controls[i].updateValueAndValidity();
      }
    }
    if (this.validateResolveForm.invalid) {
      return;
    }

    this.loaders.restartTask = true;
    this.api.resolution.add({
      ...this.item,
      resolver_inputs: InputsFormComponent.getInputs(this.validateResolveForm.value),
      start_over: true,
    }).toPromise().then((res: any) => {
      this.errors.restartTask = null;
      this.display.restartTask = false;
      this.loadTask(true);
    }).catch((err) => {
      this.errors.restartTask = err;
    }).finally(() => {
      this.loaders.restartTask = false;
    });
  }
}
