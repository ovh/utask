import { of } from 'rxjs';
import { Component, OnInit, OnDestroy, NgZone } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../@services/api.service';
import * as _ from 'lodash';
import MetaUtask from 'src/app/@models/meta-utask.model';
import { ResolutionService } from 'src/app/@services/resolution.service';
import { TaskService } from 'src/app/@services/task.service';
import { delay, repeat, startWith } from 'rxjs/operators';
import { ActiveInterval } from 'active-interval';
import * as bbPromise from 'bluebird';
import * as moment from 'moment';
import { ToastrService } from 'ngx-toastr';
import Task from 'src/app/@models/task.model';
import { environment } from 'src/environments/environment';
bbPromise.config({
  cancellation: true
});

export class SearchTask {
  page_size?: number;
  last?: string;
  type?: string;
  state?: string;
  tag?: string[];
}

@Component({
  templateUrl: './home.html',
  styleUrls: ['./home.sass'],
})
export class HomeComponent implements OnInit, OnDestroy {
  tags: string[] = [];
  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  meta: MetaUtask = null;
  tasks: Task[] = [];
  pagination: SearchTask = {};
  hasMore = true;
  percentages: { [key: string]: number } = {};
  interval: ActiveInterval;
  refresh: { [key: string]: bbPromise<any> } = {};
  display: { [key: string]: boolean } = {};
  hide: { [key: string]: boolean } = {};
  bulkActions = {
    enable: false,
    selection: {},
    actions: {},
    all: false,
  };

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router, private resolutionService: ResolutionService, private taskService: TaskService, private zone: NgZone, private toastr: ToastrService) {
  }

  ngOnInit() {
    this.tags = this.taskService.tagsRaw;
    this.taskService.tags.asObservable().subscribe((tags: string[]) => {
      this.tags = tags;
    });

    this.meta = this.route.parent.snapshot.data.meta as MetaUtask;
    this.route.queryParams.subscribe(params => {
      this.pagination = this.queryToSearchTask(params);
      this.loaders.tasks = true;
      this.loadTasks().then((tasks: Task[]) => {
        this.errors.tasks = null;
        this.tasks = tasks.map(t => this.taskService.registerTags(t));
        this.generateProgressBars(this.tasks);
        this.hasMore = (tasks as any[]).length === this.pagination.page_size;
      }).catch((err) => {
        this.errors.tasks = err;
      }).finally(() => {
        this.loaders.tasks = false;
      });
    });

    this.interval = new ActiveInterval();
    this.interval.setInterval(() => {
      if (this.tasks.length && !this.loaders.tasks) {
        const lastActivity = moment(_.maxBy(this.tasks, t => t.last_activity).last_activity).toDate();
        this.refresh.lastActivities = this.fetchLastActivities(lastActivity).then((tasks: Task[]) => {
          if (tasks.length) {
            tasks.forEach((task: Task) => {
              const t = _.find(this.tasks, { id: task.id });
              if (t) {
                this.zone.run(() => {
                  this.mergeTask(t);
                });
                this.refreshTask(t.id, 4, environment.refresh.home.task);
              } else {
                this.zone.run(() => {
                  this.display.newTasks = true;
                  this.hide[task.id] = true;
                  this.tasks.unshift(task);
                });
                this.refreshTask(task.id, 4, environment.refresh.home.task);
              }
            });
            this.generateProgressBars(this.tasks);
          }
        }).catch((err) => {
          console.log(err);
        });
      }
    }, environment.refresh.home.tasks, false);
  }

  inputTagsChanged(text: string) {
    this.pagination.tag = _.compact(text.split(' '));
    this.search();
  }

  selectAll() {
    this.tasks.forEach((t: Task) => {
      if (!this.hide[t.id]) {
        this.bulkActions.selection[t.id] = true
      }
    });
    this.checkActions();
  }

  ngOnDestroy() {
    this.cancelRefresh();
    this.interval.stopInterval();
  }

  displayNewTasks() {
    this.zone.run(() => {
      this.display.newTasks = false;
      this.tasks.forEach((t: Task) => {
        this.hide[t.id] = false;
      });
    });
  }

  fetchLastActivities(lastActivity: Date, allTasks: Task[] = [], last: string = '') {
    return new bbPromise((resolve, reject, onCancel) => {
      const loadTasks = this.loadTasks(last).then((tasks: Task[]) => {
        if (tasks.length) {
          let task;
          for (let i = 0; i < tasks.length; i++) {
            task = tasks[i];
            if (moment(task.last_activity).toDate() > lastActivity) {
              allTasks.push(task);
            }
          }
          if (allTasks.length === this.pagination.page_size) {
            this.fetchLastActivities(lastActivity, allTasks, task.id).then((data) => {
              resolve(data);
            }).catch((err) => {
              reject(err);
            });
          } else {
            resolve(_.reverse(allTasks));
          }
        } else {
          resolve(_.reverse(allTasks));
        }
      }).catch((err) => {
        reject(err);
      });

      onCancel(() => {
        loadTasks.cancel();
      });
    });
  }

  cancelRefresh() {
    if (this.refresh.lastActivities) {
      this.refresh.lastActivities.cancel();
      this.refresh.lastActivities = null;
    }
  }

  mergeTask(task) {
    const t = _.find(this.tasks, { id: task.id });
    if (t) {
      t.title = task.title;
      t.state = task.state;
      t.steps_done = task.steps_done;
      t.steps_total = task.steps_total;
      t.last_activity = task.last_activity;
      t.resolution = task.resolution;
      t.last_start = task.last_start;
      t.last_stop = task.last_stop;
      t.resolver_username = task.resolver_username;
      this.generateProgressBars([t]);
    }
  }

  refreshTask(id: string, times: number = 1, delayMillisecond: number = 2000) {
    if (!this.loaders[`task${id}`]) {
      const sub = of(id).pipe(delay(delayMillisecond)).pipe(repeat(times - 1)).pipe(startWith(id)).subscribe((id: string) => {
        this.zone.run(() => {
          this.loaders[`task${id}`] = true;
        });
        this.api.task(id).toPromise().then((task: Task) => {
          this.zone.run(() => {
            this.mergeTask(task);
          });
          if (['DONE', 'CANCELLED'].indexOf(task.state) > -1) {
            sub.unsubscribe();
          }
        });
      }, (err) => {
        if (err.status === 404) {
          _.remove(this.tasks, {
            id
          });
        }
      }, () => {
        this.zone.run(() => {
          this.loaders[`task${id}`] = false;
        });
      });
    }
  }

  loadTasks(last: string = '') {
    return new bbPromise((resolve, reject, onCancel) => {
      const params: SearchTask = _.clone(this.pagination);
      params.last = last;
      const sub = this.api.tasks(params).subscribe((data) => {
        resolve(data.body);
      }, (err) => {
        reject(err);
      });
      onCancel(() => {
        sub.unsubscribe();
      });
    });
  }

  runResolution(resolutionId: string, taskId: string) {
    this.resolutionService.run(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 4, environment.refresh.home.task);
      this.toastr.info('The resolution has been run.');
    }).catch((err) => {
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  pauseResolution(resolutionId: string, taskId: string) {
    this.resolutionService.pause(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 4, environment.refresh.home.task);
      this.toastr.info('The resolution has been paused.');
    }).catch((err) => {
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  cancelResolution(resolutionId: string, taskId: string) {
    this.resolutionService.cancel(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 1, environment.refresh.home.task);
      this.toastr.info('The resolution has been cancelled.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  deleteTask(id: string) {
    this.taskService.delete(id).then((data: any) => {
      _.remove(this.tasks, {
        id
      });
      this.toastr.info('The task has been deleted.');
    }).catch((err) => {
      if (err !== 'close') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  extendResolution(resolutionId: string, taskId: string) {
    this.resolutionService.extend(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 4, environment.refresh.home.task);
      this.toastr.info('The resolution has been extended.');
    }).catch((err) => {
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  queryToSearchTask(p?: any): SearchTask {
    const params = _.clone(p || this.router.routerState.snapshot.root.queryParams);
    if (params.tag && _.isString(params.tag)) {
      params.tag = [params.tag];
    }
    const item = new SearchTask();
    if (params.itemPerPage && _.isNumber(+params.itemPerPage) && +params.itemPerPage <= 1000 && +params.itemPerPage >= 10) {
      item.page_size = +params.itemPerPage;
    } else {
      item.page_size = 20;
    }
    const defaultType = this.meta.user_is_admin ? 'all' : 'own';
    item.type = params.type || defaultType;
    item.last = '';
    item.state = params.state || '';
    item.tag = params.tag || [];
    return item;
  }

  next(force: boolean = false) {
    if (!this.loaders.next && (this.hasMore || force)) {
      this.loaders.next = true;
      this.loadTasks(_.last(this.tasks).id).then((tasks: Task[]) => {
        this.errors.next = null;
        this.tasks = this.tasks.concat(tasks);
        this.generateProgressBars(this.tasks);
        this.hasMore = (tasks as any[]).length === this.pagination.page_size;
      }).catch((err) => {
        this.errors.next = err;
      }).finally(() => {
        this.loaders.next = false;
      });
    }
  }

  search() {
    this.cancelRefresh();
    this.zone.run(() => {
      this.bulkActions.selection = {};
      this.bulkActions.all = false;
      this.bulkActions.enable = false;
    });
    if (_.isEqual(this.pagination, this.queryToSearchTask())) {
      this.loaders.tasks = true;
      this.loadTasks().then((tasks: Task[]) => {
        this.errors.tasks = null;
        this.tasks = tasks;
        this.generateProgressBars(this.tasks);
        this.hasMore = (tasks as any[]).length === this.pagination.page_size;
      }).catch((err) => {
        this.errors.tasks = err;
      }).finally(() => {
        this.loaders.tasks = false;
      });
    } else {
      this.router.navigate([], {
        queryParams: this.pagination,
        queryParamsHandling: 'merge'
      });
    }
  }

  generateProgressBars(tasks: Task[]) {
    tasks.forEach((task: Task) => {
      this.percentages[task.id] = Math.round(task.steps_done / task.steps_total * 100);
    });
  }

  cancelAll() {
    const tmpIds = Object.keys(this.bulkActions.selection);
    const taskIds = [];
    for (let i = 0; i < tmpIds.length; i++) {
      if (this.bulkActions.selection[tmpIds[i]]) {
        taskIds.push(tmpIds[i]);
      }
    }
    const resolutionIds = _.compact(taskIds.map((id) => {
      return _.get(_.find(this.tasks, { id: id }), 'resolution');
    }));
    this.resolutionService.cancelAll(resolutionIds).then(() => {
      taskIds.forEach((id) => {
        this.refreshTask(id, 4, environment.refresh.home.task);
      });
      this.toastr.info('The tasks have been cancelled.');
      this.bulkActions.selection = {};
      this.bulkActions.all = false;
    }).catch((err) => {
      if (err !== 'close') {
        this.search();
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    }).finally(() => {
      this.checkActions();
    });
  }

  pauseAll() {
    const tmpIds = Object.keys(this.bulkActions.selection);
    const taskIds = [];
    for (let i = 0; i < tmpIds.length; i++) {
      if (this.bulkActions.selection[tmpIds[i]]) {
        taskIds.push(tmpIds[i]);
      }
    }
    const resolutionIds = _.compact(taskIds.map((id) => {
      return _.get(_.find(this.tasks, { id: id }), 'resolution');
    }));
    this.resolutionService.pauseAll(resolutionIds).then(() => {
      taskIds.forEach((id) => {
        this.refreshTask(id, 4, environment.refresh.home.task);
      });
      this.toastr.info('The tasks have been paused.');
      this.bulkActions.selection = {};
      this.bulkActions.all = false;
    }).catch((err) => {
      if (err !== 'close') {
        this.search();
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    }).finally(() => {
      this.checkActions();
    });
  }

  extendAll() {
    const tmpIds = Object.keys(this.bulkActions.selection);
    const taskIds = [];
    for (let i = 0; i < tmpIds.length; i++) {
      if (this.bulkActions.selection[tmpIds[i]]) {
        taskIds.push(tmpIds[i]);
      }
    }
    const resolutionIds = _.compact(taskIds.map((id) => {
      return _.get(_.find(this.tasks, { id: id }), 'resolution');
    }));
    this.resolutionService.extendAll(resolutionIds).then(() => {
      taskIds.forEach((id) => {
        this.refreshTask(id, 4, environment.refresh.home.task);
      });
      this.toastr.info('The tasks have been extended.');
      this.bulkActions.selection = {};
      this.bulkActions.all = false;
    }).catch((err) => {
      if (err !== 'close') {
        this.search();
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    }).finally(() => {
      this.checkActions();
    });
  }

  deleteAll() {
    const tmpIds = Object.keys(this.bulkActions.selection);
    const taskIds = [];
    for (let i = 0; i < tmpIds.length; i++) {
      if (this.bulkActions.selection[tmpIds[i]]) {
        taskIds.push(tmpIds[i]);
      }
    }
    this.taskService.deleteAll(taskIds).then(() => {
      taskIds.forEach((id: string) => {
        _.remove(this.tasks, {
          id
        });
      })
      this.toastr.info('The tasks have been deleted.');
    }).catch((err) => {
      if (err !== 'close') {
        this.search();
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
      taskIds.forEach((id) => {
        this.refreshTask(id, 1);
      });
    }).finally(() => {
      this.checkActions();
    });
  }

  runAll() {
    const tmpIds = Object.keys(this.bulkActions.selection);
    const taskIds = [];
    for (let i = 0; i < tmpIds.length; i++) {
      if (this.bulkActions.selection[tmpIds[i]]) {
        taskIds.push(tmpIds[i]);
      }
    }
    const resolutionIds = _.compact(taskIds.map((id) => {
      return _.get(_.find(this.tasks, { id: id }), 'resolution');
    }));
    this.resolutionService.runAll(resolutionIds).then(() => {
      taskIds.forEach((id) => {
        this.refreshTask(id, 4, environment.refresh.home.task);
      });
      this.toastr.info('The tasks have been run.');
      this.bulkActions.selection = {};
      this.bulkActions.all = false;
    }).catch((err) => {
      if (err !== 'close') {
        this.search();
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    }).finally(() => {
      this.checkActions();
    });
  }

  checkActions() {
    this.bulkActions.enable = false;
    if (this.meta.user_is_admin) {
      this.bulkActions.actions['delete'] = true;
    }
    this.bulkActions.actions['cancel'] = true;
    this.bulkActions.actions['run'] = true;
    this.bulkActions.actions['pause'] = true;
    this.bulkActions.actions['extend'] = true;
    const ids = Object.keys(this.bulkActions.selection);
    for (let i = 0; i < ids.length; i++) {
      const task = _.find(this.tasks, { id: ids[i] });
      if (task && this.bulkActions.selection[ids[i]] === true) {
        this.bulkActions.enable = true;
        if ((!task.resolution || task.state === 'DONE' || task.state === 'CANCELLED')) {
          this.bulkActions.actions['cancel'] = false;
          this.bulkActions.actions['run'] = false;
          this.bulkActions.actions['pause'] = false;
          this.bulkActions.actions['extend'] = false;
          if (!this.meta.user_is_admin) {
            this.toastr.info(`The task '${ids[i]}' has no resolution or is finished, you can\'t make multi actions on it.`);
          }
          break;
        } else if (task.state === 'BLOCKED') {
          this.bulkActions.actions['delete'] = false;
        }
      }
    }
  }
}
