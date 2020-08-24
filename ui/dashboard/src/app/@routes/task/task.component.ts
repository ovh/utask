import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ActiveInterval } from 'active-interval';
import { ToastrService } from 'ngx-toastr';
import { environment } from 'src/environments/environment';
import { ApiService, ModalYamlPreviewComponent } from 'utask-lib';
import EditorConfig from 'utask-lib/@models/editorconfig.model';
import { TaskService } from 'utask-lib';
import Task, { Comment } from 'utask-lib/@models/task.model';
import Meta from 'utask-lib/@models/meta.model';
import { ResolutionService } from 'utask-lib';
import { RequestService } from 'utask-lib';
import Template from 'utask-lib/@models/template.model';

@Component({
  templateUrl: './task.html',
  styleUrls: ['./task.sass'],
})
export class TaskComponent implements OnInit, OnDestroy {
  objectKeys = Object.keys;
  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  display: { [key: string]: boolean } = {};
  confirm: { [key: string]: boolean } = {};
  refreshes: { [key: string]: any } = {};
  textarea: { [key: string]: boolean } = {};
  item: any = {
    resolver_inputs: {},
    task_id: null
  };
  task: Task = null;
  taskIsResolvable = false;
  taskId = '';
  resolution: any = null;
  editorConfigResult: EditorConfig = {
    readonly: true,
    mode: 'ace/mode/json',
    theme: 'ace/theme/monokai',
    maxLines: 25,
  };
  selectedStep = '';
  meta: Meta = null;

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

  constructor(private modalService: NgbModal, private api: ApiService, private route: ActivatedRoute, private resolutionService: ResolutionService, private requestService: RequestService, private taskService: TaskService, private router: Router, private toastr: ToastrService) {
  }

  ngOnDestroy() {
    this.refreshes.tasks.stopInterval();
  }

  ngOnInit() {
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
        if (!this.task || this.task.id !== params.id) {
          this.errors.main = err;
        }
      });
    });

    this.refreshes.tasks = new ActiveInterval();
    this.refreshes.tasks.setInterval(() => {
      if (!this.loaders.task && this.autorefresh.actif) {
        this.loadTask();
      }
    }, environment.refresh.task, false);
  }

  addComment() {
    this.loaders.addComment = true;
    this.api.task.comment.add(this.task.id, this.comment.content).toPromise().then((comment: Comment) => {
      this.task.comments = _.get(this.task, 'comments', []);
      this.task.comments.push(comment);
      this.errors.addComment = null;
      this.comment.content = '';
    }).catch((err) => {
      this.errors.addComment = err;
    }).finally(() => {
      this.loaders.addComment = false;
    });
  }

  previewDetails(obj: any, title: string) {
    const previewModal = this.modalService.open(ModalYamlPreviewComponent, {
      size: 'xl'
    });
    previewModal.componentInstance.value = obj;
    previewModal.componentInstance.title = title;
    previewModal.componentInstance.close = () => {
      previewModal.close();
    };
    previewModal.componentInstance.dismiss = () => {
      previewModal.dismiss();
    };
  }

  editRequest(task: Task) {
    this.requestService.edit(task).then((data: any) => {
      this.loadTask();
      this.toastr.info('The request has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  editResolution(resolution: any) {
    this.resolutionService.edit(resolution).then((data: any) => {
      this.loadTask();
      this.toastr.info('The resolution has been edited.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  runResolution(resolution: any) {
    this.resolutionService.run(resolution.id).then((data: any) => {
      this.loadTask();
      this.toastr.info('The resolution has been run.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  pauseResolution(resolution: any) {
    this.resolutionService.pause(resolution.id).then((data: any) => {
      this.loadTask();
      this.toastr.info('The resolution has been paused.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  cancelResolution(resolution: any) {
    this.resolutionService.cancel(resolution.id).then((data: any) => {
      this.loadTask();
      this.toastr.info('The resolution has been cancelled.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  extendResolution(resolution: any) {
    this.resolutionService.extend(resolution.id).then((data: any) => {
      this.loadTask();
      this.toastr.info('The resolution has been extended.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  deleteTask(taskId: string) {
    this.taskService.delete(taskId).then((data: any) => {
      this.router.navigate([`/home`]);
      this.toastr.info('The task has been deleted.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  rejectTask() {
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
    this.loaders.resolveTask = true;
    this.api.resolution.add(this.item).toPromise().then((res: any) => {
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

  loadTask() {
    return new Promise((resolve, reject) => {
      this.loaders.task = true;
      this.api.task.get(this.taskId).subscribe((data: Task) => {
        this.task = data;
        this.task.comments = _.orderBy(_.get(this.task, 'comments', []), ['created'], ['asc']);
        this.item.task_id = this.task.id;
        this.template = _.find(this.route.parent.snapshot.data.templates, { name: this.task.template_name });
        const resolvable = this.requestService.isResolvable(this.task, this.meta, this.template.allowed_resolver_usernames || []);
        if (['DONE', 'WONTFIX', 'CANCELLED'].indexOf(this.task.state) > -1) {
          this.autorefresh.enable = false;
          this.autorefresh.actif = false;
        } else {
          this.autorefresh.enable = true;
          if (!this.autorefresh.hasChanged) {
            this.autorefresh.actif = ['TODO', 'RUNNING', 'TO_AUTORUN'].indexOf(this.task.state) > -1;
          }
        }
        if (!this.taskIsResolvable && resolvable) {
          _.get(this.template, 'resolver_inputs', []).forEach((field: any) => {
            if (field.type === 'bool' && field.default === null) {
              this.item.resolver_inputs[field.name] = false;
            } else {
              this.item.resolver_inputs[field.name] = field.default;
            }
          })
        }
        this.taskIsResolvable = resolvable;
        if (this.task.resolution) {
          this.loadResolution(this.task.resolution).then((data) => {
            if (!this.resolution && data) {
              this.display.execution = true;
              this.display.request = false;
            }
            this.resolution = data;
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
      }, (err: any) => {
        reject(err);
      }, () => {
        this.loaders.task = false;
      });
    });
  }

  loadResolution(resolutionId: string) {
    return new Promise((resolve, reject) => {
      this.loaders.resolution = true;
      this.api.resolution.get(resolutionId).subscribe((data) => {
        resolve(data);
      }, (err: any) => {
        reject(err);
      });
    });
  }
}
