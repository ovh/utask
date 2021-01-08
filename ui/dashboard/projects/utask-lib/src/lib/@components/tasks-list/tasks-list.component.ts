import {
    Component,
    OnDestroy,
    OnInit,
    Input,
    Output,
    EventEmitter,
    ViewChild,
    ChangeDetectorRef,
    ChangeDetectionStrategy,
    OnChanges,
    NgZone,
    AfterViewInit
} from '@angular/core';
import {
    forkJoin,
    interval,
    Observable,
    Subject,
    Subscription,
    throwError
} from 'rxjs';
import {
    catchError,
    concatMap,
    filter,
    finalize,
    first,
    map,
    mergeMap,
    tap
} from 'rxjs/operators';
import get from 'lodash-es/get';
import * as moment_ from 'moment';
const moment = moment_;
import Task, { TaskType } from '../../@models/task.model';
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
    more: boolean;

    constructor(t: Task, m: Meta) {
        this.run = t.resolution && t.state !== 'DONE' && t.state !== 'CANCELLED';
        this.cancel = t.resolution && t.state !== 'DONE' && t.state !== 'CANCELLED';
        this.pause = t.resolution && t.state !== 'DONE' && t.state !== 'CANCELLED';
        this.extend = t.resolution && t.state !== 'DONE' && t.state !== 'CANCELLED';
        this.delete = !(t.state === 'BLOCKED' || !m.user_is_admin);
        this.more = this.pause || this.extend || this.delete;
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
                extend: res.extend && ta.extend,
            } as TaskActions
        });
    }
}

export class TasksListComponentOptions {
    public refreshTasks = 15000;
    public refreshTask = 1000;
    public routingTaskPath = '';

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
export class TasksListComponent implements OnInit, OnDestroy, OnChanges, AfterViewInit {
    @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Task>;

    @Input() set params(data: ParamsListTasks) { this.registrerScroll.next(data); }
    get params() { return this._params; }
    _params = new ParamsListTasks();
    @Input() meta: Meta;
    @Input() options?: TasksListComponentOptions = new TasksListComponentOptions();
    @Output() public event: EventEmitter<any> = new EventEmitter();

    tasks: Task[] = [];
    tasksActions: TaskActions[] = [];
    loadingTasks: boolean;
    firstLoad: boolean;
    hasMore: boolean;
    intervalLoadNewTasks: Subscription;
    newTasks: Task[] = [];
    intervalRefreshTasks: Subscription;
    tasksToRefresh: { [key: string]: number } = {};

    bulkAllSelected: boolean;
    bulkSelection: { [key: string]: boolean } = {};
    bulkActions: TaskActions;

    errors: { [key: string]: any } = {};
    scroll = new Subject<void>();
    scrollSub: Subscription;
    registrerScroll = new Subject<ParamsListTasks>();

    titleWidth: string;

    constructor(
        private _api: ApiService,
        private _resolutionService: ResolutionService,
        private _taskService: TaskService,
        private _cd: ChangeDetectorRef,
        private _zone: NgZone
    ) {
        this.registrerScroll
            .pipe(filter(data => !this._params || !ParamsListTasks.equals(this._params, data)))
            .pipe(tap(data => this._params = { ...data }))
            .pipe(concatMap(() => this.registerInfiniteScroll()))
            .subscribe();
    }

    ngAfterViewInit() {
        let offsetWidth = (this.nzTableComponent as any).elementRef.nativeElement.offsetWidth;
        if (offsetWidth > 1100) {
            this.titleWidth = `${offsetWidth - 800}px`;
        } else {
            this.titleWidth = '300px';
        }
    }

    ngOnInit() {
        this.initLoadNewTasks();
        this.initRefreshTasks();
    }

    ngOnDestroy() {
        this.registrerScroll.complete();
        this.scroll.complete();
        this.cancelScrollSub();
        this.cancelLoadNewTasks();
        this.cancelRefreshTasks();
    }

    ngOnChanges(): void {
        this.clickCheckAll(false);
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

    async registerInfiniteScroll() {
        this.cancelScrollSub();
        this.tasks = [];
        this.tasksActions = [];
        this.firstLoad = true;
        this.hasMore = true;

        // Wait for the table to be rendered
        await interval(10)
            .pipe(map(() => !!this.nzTableComponent.nzTableInnerScrollComponent))
            .pipe(filter(exists => exists))
            .pipe(first())
            .toPromise()

        // Handle first load usecase to call load tasks until the table is not full and there is more tasks available
        const scrollComponent = this.nzTableComponent.nzTableInnerScrollComponent;
        const tableHeight = parseInt(this.nzTableComponent.nzScroll.y.split('px')[0], 10);
        let contentHeight = 0;
        while (this.firstLoad || (tableHeight >= contentHeight && this.hasMore)) {
            this.firstLoad = false;
            await this.loadMore(true);
            this._cd.detectChanges(); // force detect change to render the table and get its new size
            contentHeight = scrollComponent.tableBodyElement.nativeElement.scrollHeight;
        }

        scrollComponent.tableBodyElement.nativeElement.onscroll = () => { this.scroll.next(); };

        this._zone.runOutsideAngular(() => {
            this.scrollSub = this.scroll
                .pipe(mergeMap(() => this.loadMore()))
                .subscribe();
        });
    }

    cancelScrollSub(): void {
        if (this.scrollSub) { this.scrollSub.unsubscribe() };
    }

    async loadMore(skipCheckScroll = false) {
        if (this.loadingTasks) { return; }

        const scrollComponent = this.nzTableComponent.nzTableInnerScrollComponent;
        const height = scrollComponent.tableBodyElement.nativeElement.offsetHeight;
        const innerHeight = scrollComponent.tableBodyElement.nativeElement.scrollHeight;
        const scrollTop = scrollComponent.tableBodyElement.nativeElement.scrollTop;
        const scrollBottom = innerHeight - (scrollTop + height)
        if (!skipCheckScroll && (scrollBottom > 200 || !this.hasMore)) {
            return;
        }

        const paramLast = this.tasks.length > 0 ? this.tasks[this.tasks.length - 1].id : '';

        this.loadingTasks = true;
        this.errors.loadMore = null;
        this._cd.markForCheck();
        const tasks = await this.loadTasks(paramLast)
            .pipe(catchError(err => {
                this.errors.loadMore = err;
                return throwError(err);
            }))
            .pipe(finalize(() => {
                this.loadingTasks = false;
                this._cd.markForCheck();
            })).toPromise();

        this.tasks = this.tasks.concat(tasks);
        this.computeTaskActions();
        tasks.forEach(t => this._taskService.registerTags(t));
        this.hasMore = tasks.length === this.params.page_size;
        this._cd.detectChanges();
    }

    loadTasks(paramLast: string = ''): Observable<Array<Task>> {
        // Trick to get both own and resolvable task for non admin user
        // We ignore the last param for resolvable tasks so we will only get the first ones
        if (this.params.type === TaskType.both) {
            return forkJoin({
                resolvable: this._api.task.list({
                    ...this.params,
                    type: TaskType.resolvable
                }),
                own: this._api.task.list({
                    ...this.params,
                    type: TaskType.own,
                    last: paramLast
                })
            }).pipe(map(r => r.own.body.concat(r.resolvable.body)));
        } else {
            return this._api.task.list({
                ...this.params,
                last: paramLast
            }).pipe(map(res => res.body));
        }
    }

    clickShowMore(): void {
        this.scroll.next();
    }

    trackInput(index: number, task: Task) {
        return task.id + task.last_activity;
    }

    // Manage fetch task
    // search params should not be listed in browser url if empty

    initLoadNewTasks() {
        this._zone.runOutsideAngular(() => {
            this.intervalLoadNewTasks = interval(this.options.refreshTasks)
                .pipe(filter(() => this.tasks.length > 0))
                .pipe(map(() => {
                    const taskLastChanged = this.tasks.sort((a, b) => a.last_activity > b.last_activity ? -1 : 1)[0];
                    return moment(taskLastChanged.last_activity).toDate();
                }))
                .pipe(concatMap((lastActivity) => this.fetchNewTasks(lastActivity)))
                .subscribe(() => { });
        });
    }

    async fetchNewTasks(lastActivity: Date) {
        this._zone.run(() => {
            this.errors.fetchNewTasks = null;
            this._cd.markForCheck();
        });
        const newTasks = await this.loadTasks()
            .pipe(catchError(err => {
                this._zone.run(() => {
                    this.errors.fetchNewTasks = err;
                    this._cd.markForCheck();
                });
                return throwError(err);
            }))
            .toPromise();

        this._zone.run(() => {
            // Last tasks will be added to new tasks list waiting for the user to display it
            this.newTasks = newTasks
                .filter(t => moment(t.last_activity).toDate() > lastActivity)
                .filter(t => !this.tasks.find(ta => ta.id === t.id))
                .sort((a, b) => a.last_activity > b.last_activity ? -1 : 1);

            // Also we update the tasks that are already visible
            // We don't refresh all the tasks in the table but the only the one returns when searching for new tasks
            this.tasks = this.tasks.map(t => {
                const updatedTask = newTasks.find(ta => ta.id === t.id);
                return updatedTask ? updatedTask : t;
            });
            this.computeTaskActions();

            this._cd.markForCheck();
        });
    }

    cancelLoadNewTasks() {
        if (this.intervalLoadNewTasks) { this.intervalLoadNewTasks.unsubscribe(); }
    }

    clickShowNewTasks() {
        this.tasks = this.newTasks.concat(this.tasks);
        this.computeTaskActions();
        this.newTasks = [];
        this._cd.markForCheck();
    }

    initRefreshTasks() {
        this.intervalRefreshTasks = interval(this.options.refreshTask)
            .pipe(concatMap(() => this.refreshTasks()))
            .subscribe(() => { });
    }

    cancelRefreshTasks() {
        if (this.intervalRefreshTasks) { this.intervalRefreshTasks.unsubscribe(); }
    }

    async refreshTasks() {
        const taskIDs = Object.keys(this.tasksToRefresh)
            .filter(k => this.tasksToRefresh[k] > 0);

        if (taskIDs.length === 0) {
            return;
        }

        this.errors.refreshTasks = null;
        this._cd.markForCheck();

        let tasks: Array<Task>;
        try {
            tasks = await Promise.all(taskIDs.map(k => this._api.task.get(k).toPromise()))
        } catch (e) {
            this.errors.refreshTasks = e;
            this._cd.markForCheck();
            return;
        }

        this.tasks = this.tasks.map(t => {
            const updatedTask = tasks.find(ta => ta.id === t.id);
            return updatedTask ? updatedTask : t;
        });
        this.computeTaskActions();

        // Decrement or stop task refresh
        tasks.forEach(t => {
            if (['DONE', 'CANCELLED'].indexOf(t.state) > -1) {
                this.tasksToRefresh[t.id] = 0;
            } else {
                this.tasksToRefresh[t.id]--;
            }
        });

        this._cd.markForCheck();
    }

    computeTaskActions() {
        this.tasksActions = this.tasks.map(t => new TaskActions(t, this.meta));
    }

    // Manage actions on task

    cancelResolution(resolutionId: string, taskId: string) {
        this._resolutionService.cancel(resolutionId).then((data: any) => {
            this.tasksToRefresh[taskId] = 1;
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
        this._resolutionService.cancelAll(resolutionIds).then(() => {
            resolutionIds.forEach(id => { this.tasksToRefresh[id] = 4; });
            this.event.emit({ type: 'info', message: 'The tasks have been cancelled.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
        });
    }

    pauseResolution(resolutionId: string, taskId: string) {
        this._resolutionService.pause(resolutionId).then((data: any) => {
            this.tasksToRefresh[taskId] = 4;
            this.event.emit({ type: 'info', message: 'The resolution has been paused.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }

    pauseAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);

        this._resolutionService.pauseAll(resolutionIds).then(() => {
            resolutionIds.forEach(id => { this.tasksToRefresh[id] = 4; });
            this.event.emit({ type: 'info', message: 'The tasks have been paused.' });
        }).catch((err) => {
            if (err === 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
        });
    }

    extendResolution(resolutionId: string, taskId: string) {
        this._resolutionService.extend(resolutionId).then((data: any) => {
            this.tasksToRefresh[taskId] = 4;
            this.event.emit({ type: 'info', message: 'The resolution has been extended.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }

    extendAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);
        this._resolutionService.extendAll(resolutionIds).then(() => {
            resolutionIds.forEach(id => { this.tasksToRefresh[id] = 4; });
            this.event.emit({ type: 'info', message: 'The tasks have been extended.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
        });
    }

    deleteTask(id: string) {
        this._taskService.delete(id).then(() => {
            this.tasks = this.tasks.filter(t => t.id !== id);
            this.computeTaskActions();
            this._cd.markForCheck();
            this.event.emit({ type: 'info', message: 'The task has been deleted.' });
        });
    }

    deleteAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        this._taskService.deleteAll(selectedTaskIDs).then(() => {
            this.tasks = this.tasks.filter(t => !selectedTaskIDs.find(id => id === t.id));
            this.computeTaskActions();
            this._cd.markForCheck();
            this.event.emit({ type: 'info', message: 'The tasks have been deleted.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => {
            this.clickCheckAll(false);
        });
    }

    runResolution(resolutionId: string, taskId: string) {
        this._resolutionService.run(resolutionId).then((data: any) => {
            this.tasksToRefresh[taskId] = 4;
            this.event.emit({ type: 'info', message: 'The resolution has been run.' });
        }).catch((err) => {
            this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
        });
    }

    runAll() {
        const selectedTaskIDs = Object.keys(this.bulkSelection).filter(key => this.bulkSelection[key]);
        const selectedTask = this.tasks.filter(t => selectedTaskIDs.find(id => id === t.id));
        const resolutionIds = selectedTask.map(t => t.resolution);
        this._resolutionService.runAll(resolutionIds).then(() => {
            resolutionIds.forEach(id => { this.tasksToRefresh[id] = 4; });
            this.event.emit({ type: 'info', message: 'The tasks have been run.' });
        }).catch((err) => {
            if (err !== 'close') {
                this.event.emit({ type: 'error', message: get(err, 'error.error', 'An error just occured, please retry') });
            }
        }).finally(() => { this.clickCheckAll(false); });
    }
}





