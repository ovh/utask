import { of } from 'rxjs';
import { Component, OnInit, OnDestroy, NgZone } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../@services/api.service';
import * as _ from 'lodash';
import MetaUtask from 'src/app/@models/meta-utask.model';
import { ResolutionService } from 'src/app/@services/resolution.service';
import { TaskService } from 'src/app/@services/task.service';
import { delay, repeat } from 'rxjs/operators';
import { ActiveInterval } from 'active-interval';
import * as bbPromise from 'bluebird';
import * as moment from 'moment';
import { ToastrService } from 'ngx-toastr';
bbPromise.config({
  cancellation: true
});

export class SearchTask {
  page_size?: number;
  last?: string;
  type?: string;
  state?: string;
}

@Component({
  templateUrl: './home.html',
})
export class HomeComponent implements OnInit, OnDestroy {
  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  meta: MetaUtask = null;
  tasks: any = [];
  pagination: SearchTask = {};
  hasMore = true;
  percentages: { [key: string]: number } = {};
  interval: ActiveInterval;
  refresh: { [key: string]: bbPromise<any> } = {};
  display: { [key: string]: boolean } = {};
  displayTest: boolean = false;

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router, private resolutionService: ResolutionService, private taskService: TaskService, private zone: NgZone, private toastr: ToastrService) {
  }

  ngOnInit() {
    this.meta = this.route.parent.snapshot.data.meta as MetaUtask;
    this.route.queryParams.subscribe(params => {
      this.pagination = this.queryToSearchTask(params);
      this.loaders.tasks = true;
      this.loadTasks().then((tasks) => {
        this.errors.tasks = null;
        this.tasks = tasks;
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
      if (this.tasks.length) {
        const lastActivity = moment(_.maxBy(this.tasks, t => t.last_activity).last_activity).toDate();
        this.refresh.lastActivities = this.fetchLastActivities(lastActivity).then((tasks: any[]) => {
          if (tasks.length) {
            tasks.forEach((task: any) => {
              const t = _.find(this.tasks, { id: task.id });
              if (t) {
                this.zone.run(() => {
                  this.mergeTask(t);
                });
                this.refreshTask(t.id, 4, 1000);
              } else {
                this.zone.run(() => {
                  this.display.newTasks = true;
                  task.hide = true;
                  this.tasks.unshift(task);
                });
                this.refreshTask(task.id, 4, 1000);
              }
            });
            this.generateProgressBars(this.tasks);
          }
        }).catch((err) => {
          console.log(err);
        });
      }
    }, 15000, false);
  }

  ngOnDestroy() {
    this.cancelRefresh();
    this.interval.stopInterval();
  }

  displayNewTasks() {
    this.zone.run(() => {
      this.display.newTasks = false;
      this.tasks.forEach((t) => {
        t.hide = false;
      });
    });
  }

  fetchLastActivities(lastActivity: Date, allTasks: any[] = [], last: string = '') {
    return new bbPromise((resolve, reject, onCancel) => {
      const loadTasks = this.loadTasks(last).then((tasks: any[]) => {
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
      const sub = of(id).pipe(delay(delayMillisecond)).pipe(repeat(times - 1)).subscribe((id: string) => {
        this.zone.run(() => {
          this.loaders[`task${id}`] = true;
        });
        this.api.task(id).toPromise().then((task: any) => {
          this.zone.run(() => {
            this.mergeTask(task);
          });
          if (['DONE', 'CANCELLED'].indexOf(task.state) > -1) {
            sub.unsubscribe();
          }
        });
      }, (err) => {
        console.log(err);
      }, () => {
        this.zone.run(() => {
          this.loaders[`task${id}`] = false;
        });
      });

      this.loaders[`task${id}`] = true;
      this.api.task(id).toPromise().then((task: any) => {
        if (['DONE', 'CANCELLED'].indexOf(task.state) > -1 && times > 1) {
          sub.unsubscribe();
        }
        this.zone.run(() => {
          this.mergeTask(task);
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
      this.refreshTask(taskId, 4, 1000);
      this.toastr.info('The resolution has been run.');
    }).catch((err) => {
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  pauseResolution(resolutionId: string, taskId: string) {
    this.resolutionService.pause(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 4, 1000);
      this.toastr.info('The resolution has been paused.');
    }).catch((err) => {
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  cancelResolution(resolutionId: string, taskId: string) {
    this.resolutionService.cancel(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 1, 1000);
      this.toastr.info('The resolution has been cancelled.');
    }).catch((err) => {
      if (err !== 0 && err !== 'Cross click') {
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
      if (err !== 0 && err !== 'Cross click') {
        this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
      }
    });
  }

  extendResolution(resolutionId: string, taskId: string) {
    this.resolutionService.extend(resolutionId).then((data: any) => {
      this.refreshTask(taskId, 4, 1000);
      this.toastr.info('The resolution has been extended.');
    }).catch((err) => {
      console.log(err);
      this.toastr.error(_.get(err, 'error.error', 'An error just occured, please retry'));
    });
  }

  queryToSearchTask(p?: any): SearchTask {
    const params = p || this.router.routerState.snapshot.root.queryParams;
    const item = new SearchTask();
    if (params.itemPerPage && _.isNumber(+params.itemPerPage) && +params.itemPerPage <= 1000 && +params.itemPerPage >= 10) {
      item.page_size = +params.itemPerPage;
    } else {
      item.page_size = 20;
    }
    const defaultType = this.meta.user_is_admin ? 'all' : 'own';
    item.type = params.type ? params.type : defaultType;
    item.last = '';
    item.state = params.state ? params.state : '';
    return item;
  }

  next(force: boolean = false) {
    if (!this.loaders.next && (this.hasMore || force)) {
      this.loaders.next = true;
      this.loadTasks(_.last(this.tasks).id).then((tasks) => {
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
    if (_.isEqual(this.pagination, this.queryToSearchTask())) {
      this.loaders.tasks = true;
      this.loadTasks().then((tasks) => {
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

  generateProgressBars(tasks: any[]) {
    tasks.forEach((task: any) => {
      this.percentages[task.id] = Math.round(task.steps_done / task.steps_total * 100);
    });
  }
}
