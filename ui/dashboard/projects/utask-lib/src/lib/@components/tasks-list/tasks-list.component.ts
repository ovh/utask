import { Component, OnDestroy, OnInit, Input, NgZone, OnChanges, Output, EventEmitter, ViewChild, ChangeDetectorRef, ChangeDetectionStrategy } from '@angular/core';
import { Observable, of, Subscription } from 'rxjs';
import { catchError, delay, finalize, map, repeat, startWith } from 'rxjs/operators';
import { ActiveInterval } from 'active-interval';
import remove from 'lodash-es/remove';
import get from 'lodash-es/get';
import last from 'lodash-es/last';
import reverse from 'lodash-es/reverse';
import maxBy from 'lodash-es/maxBy';
import * as moment_ from 'moment';
const moment = moment_;
import Task from '../../@models/task.model';
import { ParamsListTasks, ApiService } from '../../@services/api.service';
import Meta from '../../@models/meta.model';
import { ResolutionService } from '../../@services/resolution.service';
import { TaskService } from '../../@services/task.service';
import { NzTableComponent } from 'ng-zorro-antd/table';

export class TaskActions {
    delete: boolean;
    cancel: boolean;
    run: boolean;
    pause: boolean;
    extend: boolean;

    constructor(t: Task, m: Meta) {
        this.delete = m.user_is_admin && t.state !== 'BLOCKED';
        this.cancel = !(!t.resolution || t.state === 'DONE' || t.state === 'CANCELLED');
        this.run = !(!t.resolution || t.state === 'DONE' || t.state === 'CANCELLED');
        this.pause = !(!t.resolution || t.state === 'DONE' || t.state === 'CANCELLED');
        this.extend = !(!t.resolution || t.state === 'DONE' || t.state === 'CANCELLED');
    }

    public static mergeTaskActions(tas: Array<TaskActions>): TaskActions {
        if (tas.length === 0) {
            return null;
        }
        return tas.reduce((res, ta) => {
            if (!res) {
                return { ...ta };
            }
            return {
                delete: res.delete && ta.delete,
                cancel: res.cancel && ta.cancel,
                run: res.run && ta.run,
                pause: res.pause && ta.pause,
                extend: res.extend && ta.extend
            }
        });
    }
}

export class TasksListComponentOptions {
    public refreshTasks = 15000;
    public refreshTask = 1000;
    public routingTask = '/task/';

    public constructor(init?: Partial<TasksListComponentOptions>) {
        Object.assign(this, init);
    }
}

@Component({
    selector: 'lib-utask-tasks-list',
    templateUrl: './tasks-list.html',
    styleUrls: ['./tasks-list.sass'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class TasksListComponent implements OnInit, OnDestroy, OnChanges {
    @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Task>;

    @Input() params: ParamsListTasks;
    @Input() meta: Meta;
    @Input() options?: TasksListComponentOptions = new TasksListComponentOptions();
    @Output() public event: EventEmitter<any> = new EventEmitter();

    tasks: Task[] = [];
    hasMore = true;
    firstLoad = true;
    interval: ActiveInterval;
    refresh: { [key: string]: Subscription } = {};
    display: { [key: string]: boolean } = {};

    bulkAllSelected: boolean;
    bulkSelection: { [key: string]: boolean } = {};
    bulkActions: TaskActions;

    loaders: { [key: string]: boolean } = {};
    errors: { [key: string]: any } = {};
    iterableDiffer: any;
    scrollSub: Subscription;

    constructor(
        private api: ApiService,
        private resolutionService: ResolutionService,
        private taskService: TaskService,
        private zone: NgZone,
        private _cd: ChangeDetectorRef
    ) { }

    ngOnInit() { this.initLoad(); }

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


    // Manage task selection

    clickCheckAll(checked: boolean): void {
        this.bulkSelection = {};
        this.bulkActions = null;
        if (checked) {
            this.tasks.forEach(t => { this.bulkSelection[t.id] = true; });
            this.bulkActions = TaskActions.mergeTaskActions(this.tasks.map(t => new TaskActions(t, this.meta)));
        }
        this._cd.markForCheck();
        this.refreshCheckAllState();
    }

    clickCheckTask(taskID: number, checked: boolean): void {
        this.bulkSelection[taskID] = checked;
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        this.bulkActions = TaskActions.mergeTaskActions(selectedTask.map(t => new TaskActions(t, this.meta)));
        this._cd.markForCheck();
        this.refreshCheckAllState();
    }

    refreshCheckAllState(): void {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        this.bulkAllSelected = this.tasks && this.tasks.length > 0 && this.tasks.length === selectedTaskIDs.length;
        this._cd.markForCheck();
    }


    // Manage infinite scroll
    // TODO infinite scroll should load new tzdk

    registerScroll(): void {
        if (this.scrollSub) { this.scrollSub.unsubscribe() }
        this.scrollSub = this.nzTableComponent?.cdkVirtualScrollViewport?.scrolledIndexChange.subscribe((data: number) => {
            console.log('scroll index to', data);
        });
    }

    trackByIndex(n: number, data: Task): number {
        return n;
    }


    // Manage fetch task
    // TODO subscribe for new task and refresh edited tasks
    // search params should not be listed in browser url if empty

    initLoad() {
        this.loaders.tasks = true;
        this._cd.markForCheck();
        this.loadTasks()
            .pipe(catchError((err, tasks) => {
                this.errors.tasks = err;
                return tasks;
            }))
            .pipe(finalize(() => {
                this.loaders.tasks = false;
                this._cd.markForCheck();
            }))
            .subscribe((tasks: Task[]) => {
                this.errors.tasks = null;
                this.tasks = tasks.map(t => this.taskService.registerTags(t));
                this.hasMore = tasks.length === this.params.page_size;
            });

        this.interval = new ActiveInterval();
        this.interval.setInterval(() => {
            if (this.tasks.length > 0 && !this.loaders.tasks) {
                const lastActivity = moment(maxBy(this.tasks, t => t.last_activity).last_activity).toDate();
                this.refresh.lastActivities = this.fetchLastActivities(lastActivity)
                    .subscribe((tasks: Task[]) => {
                        tasks.forEach((task: Task) => {
                            const t = this.tasks.find(ta => ta.id === task.id);
                            if (t) {
                                this.zone.run(() => {
                                    this.mergeTask(t);
                                });
                                this.refreshTask(t.id, 4, this.options.refreshTask);
                            } else {
                                this.zone.run(() => {
                                    this.tasks.unshift(task);
                                });
                                this.refreshTask(task.id, 4, this.options.refreshTask);
                            }
                        });
                    });
            }
        }, this.options.refreshTasks, false);
    }

    mergeTask(task: Task) {
        const i = this.tasks.findIndex(ta => ta.id === task.id);
        if (i < 0) {
            return;
        }
        this.tasks[i].title = task.title;
        this.tasks[i].state = task.state;
        this.tasks[i].steps_done = task.steps_done;
        this.tasks[i].steps_total = task.steps_total;
        this.tasks[i].last_activity = task.last_activity;
        this.tasks[i].resolution = task.resolution;
        this.tasks[i].last_start = task.last_start;
        this.tasks[i].last_stop = task.last_stop;
        this.tasks[i].resolver_username = task.resolver_username;
    }

    fetchLastActivities(lastActivity: Date, allTasks: Task[] = [], last: string = ''): Observable<Array<Task>> {
        return this.loadTasks(last)
            .pipe(map((tasks) => {
                return reverse(tasks.filter(t => moment(t.last_activity).toDate() > lastActivity));
            }));
    }

    cancelRefresh() {
        if (this.refresh.lastActivities) {
            this.refresh.lastActivities = null;
        }
    }

    refreshTask(id: string, times: number = 1, delayMillisecond: number = 2000) {
        if (this.loaders[`task${id}`]) {
            return;
        }
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

    loadTasks(paramLast: string = ''): Observable<Array<Task>> {
        return this.api.task.list({
            ...this.params,
            last: paramLast
        }).pipe(map(res => res.body));
    }

    next(force: boolean = false) {
        if (this.loaders.next || (!this.hasMore && !force)) {
            return;
        }

        this.loaders.next = true;
        this.loadTasks(last(this.tasks).id)
            .pipe(catchError((err, tasks) => {
                this.errors.next = err;
                return tasks;
            }))
            .pipe(finalize(() => {
                this.loaders.next = false;
                this._cd.markForCheck();
            }))
            .subscribe((tasks: Task[]) => {
                this.errors.next = null;
                this.tasks = this.tasks.concat(tasks);
                this.hasMore = (tasks as any[]).length === this.params.page_size;
            });
    }

    search() {
        this.clickCheckAll(false);
        this.cancelRefresh();

        this.loaders.tasks = true;
        this.loadTasks()
            .pipe(catchError((err, tasks) => {
                this.errors.tasks = err;
                return tasks;
            }))
            .pipe(finalize(() => {
                this.loaders.tasks = false;
                this._cd.markForCheck();
            }))
            .subscribe((tasks: Task[]) => {
                this.errors.tasks = null;
                this.tasks = tasks;
                this.hasMore = (tasks as any[]).length === this.params.page_size;
            });
    }


    // Manage actions on task

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

    cancelAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);

        this.resolutionService.cancelAll(resolutionIds).then(() => {
            // TODO fix refresh tasks
            // taskIds.forEach((id) => {
            //     this.refreshTask(id, 4, this.options.refreshTask);
            // });
            this.event.emit({ type: 'info', message: 'The tasks have been cancelled.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
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

    pauseAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);

        this.resolutionService.pauseAll(resolutionIds).then(() => {
            // TODO fix refresh tasks
            //taskIds.forEach((id) => {
            //    this.refreshTask(id, 4, this.options.refreshTask);
            //});
            this.event.emit({ type: 'info', message: 'The tasks have been paused.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
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

    extendAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);

        this.resolutionService.extendAll(resolutionIds).then(() => {
            // TODO fix refresh tasks
            //taskIds.forEach((id) => {
            //    this.refreshTask(id, 4, this.options.refreshTask);
            //});
            this.event.emit({ type: 'info', message: 'The tasks have been extended.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
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

    deleteAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);

        this.taskService.deleteAll(selectedTaskIDs).then(() => {
            // TODO fix refresh tasks
            //taskIds.forEach((id: string) => {
            //    remove(this.tasks, { id });
            //})
            this.event.emit({ type: 'info', message: 'The tasks have been deleted.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
            // TODO fix refresh tasks
            //taskIds.forEach((id) => {
            //    this.refreshTask(id, 1);
            //});
        }).finally(() => {
            this.clickCheckAll(false);
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

    runAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);

        this.resolutionService.runAll(resolutionIds).then(() => {
            // TODO fix refresh tasks
            //taskIds.forEach((id) => {
            //    this.refreshTask(id, 4, this.options.refreshTask);
            //});
            this.event.emit({ type: 'info', message: 'The tasks have been run.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.search();
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => { this.clickCheckAll(false); });
    }
}





