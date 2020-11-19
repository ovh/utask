import { Component, OnDestroy, OnInit, Input, NgZone, OnChanges, Output, EventEmitter } from '@angular/core';
import { of } from 'rxjs';
import { delay, repeat, startWith } from 'rxjs/operators';
import { ActiveInterval } from 'active-interval';
import * as bbPromise from 'bluebird';
import remove from 'lodash-es/remove';
import get from 'lodash-es/get';
import last from 'lodash-es/last';
import reverse from 'lodash-es/reverse';
import compact from 'lodash-es/compact';
import maxBy from 'lodash-es/maxBy';
import clone from 'lodash-es/clone';
import * as moment_ from 'moment';
const moment = moment_;
import Task from '../../@models/task.model';
import { ParamsListTasks, ApiService } from '../../@services/api.service';
import Meta from '../../@models/meta.model';
import { ResolutionService } from '../../@services/resolution.service';
import { TaskService } from '../../@services/task.service';

export class TasksListComponentOptions {
    public refreshTasks = 15000;
    public refreshTask = 1000;
    public routingTask: string = '/task/';

    public constructor(init?: Partial<TasksListComponentOptions>) {
        Object.assign(this, init);
    }
}

bbPromise.config({
    cancellation: true
});

@Component({
    selector: 'lib-utask-tasks-list',
    templateUrl: './tasks-list.html',
    styleUrls: ['./tasks-list.sass'],
})
export class TasksListComponent implements OnInit, OnDestroy, OnChanges {
    tasks: Task[] = [];
    @Input() params: ParamsListTasks;
    @Input() meta: Meta;
    @Input() options?: TasksListComponentOptions = new TasksListComponentOptions();
    @Output() public event: EventEmitter<any> = new EventEmitter();

    hasMore = true;
    firstLoad = true;
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

    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};
    iterableDiffer: any;

    constructor(private api: ApiService, private resolutionService: ResolutionService, private taskService: TaskService, private zone: NgZone) {
    }

    ngOnInit() {
        this.loaders.tasks = true;
        this.loadTasks().then((tasks: Task[]) => {
            this.errors.tasks = null;
            this.tasks = tasks.map(t => this.taskService.registerTags(t));
            this.generateProgressBars(this.tasks);
            this.hasMore = (tasks as any[]).length === this.params.page_size;
        }).catch((err) => {
            this.errors.tasks = err;
        }).finally(() => {
            this.loaders.tasks = false;
        });

        this.interval = new ActiveInterval();
        this.interval.setInterval(() => {
            if (this.tasks.length && !this.loaders.tasks) {
                const lastActivity = moment(maxBy(this.tasks, t => t.last_activity).last_activity).toDate();
                this.refresh.lastActivities = this.fetchLastActivities(lastActivity).then((tasks: Task[]) => {
                    if (tasks.length) {
                        tasks.forEach((task: Task) => {
                            const t = this.tasks.find(ta => ta.id === task.id);
                            if (t) {
                                this.zone.run(() => {
                                    this.mergeTask(t);
                                });
                                this.refreshTask(t.id, 4, this.options.refreshTask);
                            } else {
                                this.zone.run(() => {
                                    this.display.newTasks = true;
                                    this.hide[task.id] = true;
                                    this.tasks.unshift(task);
                                });
                                this.refreshTask(task.id, 4, this.options.refreshTask);
                            }
                        });
                        this.generateProgressBars(this.tasks);
                    }
                }).catch((err) => {
                    console.log(err);
                });
            }
        }, this.options.refreshTasks, false);
    }

    ngOnDestroy() {
        this.cancelRefresh();
        this.interval.stopInterval();
    }

    ngOnChanges() {
        if (!this.firstLoad) {
            this.search();
        }
        this.firstLoad = false;
    }

    selectAll() {
        this.tasks.forEach((t: Task) => {
            if (!this.hide[t.id]) {
                this.bulkActions.selection[t.id] = true
            }
        });
        this.checkActions();
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
            const pLoadTasks = this.loadTasks(last).then((tasks: Task[]) => {
                if (tasks.length) {
                    let task;
                    for (let i = 0; i < tasks.length; i++) {
                        task = tasks[i];
                        if (moment(task.last_activity).toDate() > lastActivity) {
                            allTasks.push(task);
                        }
                    }
                    if (allTasks.length === this.params.page_size) {
                        this.fetchLastActivities(lastActivity, allTasks, task.id).then((data) => {
                            resolve(data);
                        }).catch((err) => {
                            reject(err);
                        });
                    } else {
                        resolve(reverse(allTasks));
                    }
                } else {
                    resolve(reverse(allTasks));
                }
            }).catch((err) => {
                reject(err);
            });

            if (onCancel) {
                onCancel(() => {
                    pLoadTasks.cancel();
                });
            }
        });
    }

    cancelRefresh() {
        if (this.refresh.lastActivities) {
            this.refresh.lastActivities.cancel();
            this.refresh.lastActivities = null;
        }
    }

    mergeTask(task) {
        const t = this.tasks.find(ta => ta.id === task.id);
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
                this.api.task.get(id).toPromise().then((task: Task) => {
                    this.zone.run(() => {
                        this.mergeTask(task);
                    });
                    if (['DONE', 'CANCELLED'].indexOf(task.state) > -1) {
                        sub.unsubscribe();
                    }
                });
            }, (err) => {
                if (err.status === 404) {
                    remove(this.tasks, { id });
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
            const params: ParamsListTasks = clone(this.params);
            params.last = last;
            const sub = this.api.task.list(params).subscribe((data: any) => {
                resolve(data.body);
            }, (err) => {
                reject(err);
            });
            if (onCancel) {
                onCancel(() => {
                    sub.unsubscribe();
                });
            }
        });
    }

    runResolution(resolutionId: string, taskId: string) {
        this.resolutionService.run(resolutionId).then((data: any) => {
            this.refreshTask(taskId, 4, this.options.refreshTask);
            this.event.emit({ type: 'info', message: 'The resolution has been run.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }

    pauseResolution(resolutionId: string, taskId: string) {
        this.resolutionService.pause(resolutionId).then((data: any) => {
            this.refreshTask(taskId, 4, this.options.refreshTask);
            this.event.emit({ type: 'info', message: 'The resolution has been paused.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }

    cancelResolution(resolutionId: string, taskId: string) {
        this.resolutionService.cancel(resolutionId).then((data: any) => {
            this.refreshTask(taskId, 1, this.options.refreshTask);
            this.event.emit({ type: 'info', message: 'The resolution has been cancelled.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        });
    }

    deleteTask(id: string) {
        this.taskService.delete(id).then((data: any) => {
            remove(this.tasks, { id });
            this.event.emit({ type: 'info', message: 'The task has been deleted.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        });
    }

    extendResolution(resolutionId: string, taskId: string) {
        this.resolutionService.extend(resolutionId).then((data: any) => {
            this.refreshTask(taskId, 4, this.options.refreshTask);
            this.event.emit({ type: 'info', message: 'The resolution has been extended.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }


    next(force: boolean = false) {
        if (!this.loaders.next && (this.hasMore || force)) {
            this.loaders.next = true;
            this.loadTasks(last(this.tasks).id).then((tasks: Task[]) => {
                this.errors.next = null;
                this.tasks = this.tasks.concat(tasks);
                this.generateProgressBars(this.tasks);
                this.hasMore = (tasks as any[]).length === this.params.page_size;
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

        this.loaders.tasks = true;
        this.loadTasks().then((tasks: Task[]) => {
            this.errors.tasks = null;
            this.tasks = tasks;
            this.generateProgressBars(this.tasks);
            this.hasMore = (tasks as any[]).length === this.params.page_size;
        }).catch((err) => {
            this.errors.tasks = err;
        }).finally(() => {
            this.loaders.tasks = false;
        });
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
        const resolutionIds = compact(taskIds.map((id) => {
            return get(this.tasks.find(t => t.id === id), 'resolution');
        }));
        this.resolutionService.cancelAll(resolutionIds).then(() => {
            taskIds.forEach((id) => {
                this.refreshTask(id, 4, this.options.refreshTask);
            });
            this.event.emit({ type: 'info', message: 'The tasks have been cancelled.' });
            this.bulkActions.selection = {};
            this.bulkActions.all = false;
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
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
        const resolutionIds = compact(taskIds.map((id) => {
            return get(this.tasks.find(t => t.id === id), 'resolution');
        }));
        this.resolutionService.pauseAll(resolutionIds).then(() => {
            taskIds.forEach((id) => {
                this.refreshTask(id, 4, this.options.refreshTask);
            });
            this.event.emit({ type: 'info', message: 'The tasks have been paused.' });
            this.bulkActions.selection = {};
            this.bulkActions.all = false;
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
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
        const resolutionIds = compact(taskIds.map((id) => {
            return get(this.tasks.find(t => t.id === id), 'resolution');
        }));
        this.resolutionService.extendAll(resolutionIds).then(() => {
            taskIds.forEach((id) => {
                this.refreshTask(id, 4, this.options.refreshTask);
            });
            this.event.emit({ type: 'info', message: 'The tasks have been extended.' });
            this.bulkActions.selection = {};
            this.bulkActions.all = false;
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
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
                remove(this.tasks, { id });
            })
            this.event.emit({ type: 'info', message: 'The tasks have been deleted.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
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
        const resolutionIds = compact(taskIds.map((id) => {
            return get(this.tasks.find(t => t.id === id), 'resolution');
        }));
        this.resolutionService.runAll(resolutionIds).then(() => {
            taskIds.forEach((id) => {
                this.refreshTask(id, 4, this.options.refreshTask);
            });
            this.event.emit({ type: 'info', message: 'The tasks have been run.' });
            this.bulkActions.selection = {};
            this.bulkActions.all = false;
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
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
            const task = this.tasks.find(t => t.id === ids[i]);
            if (task && this.bulkActions.selection[ids[i]] === true) {
                this.bulkActions.enable = true;
                if ((!task.resolution || task.state === 'DONE' || task.state === 'CANCELLED')) {
                    this.bulkActions.actions['cancel'] = false;
                    this.bulkActions.actions['run'] = false;
                    this.bulkActions.actions['pause'] = false;
                    this.bulkActions.actions['extend'] = false;
                    if (!this.meta.user_is_admin) {
                        this.event.emit({ type: 'info', message: `The task '${ids[i]}' has no resolution or is finished, you can\'t make multi actions on it.` });
                    }
                    break;
                } else if (task.state === 'BLOCKED') {
                    this.bulkActions.actions['delete'] = false;
                }
            }
        }
    }
}