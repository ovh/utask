import { ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import isArray from 'lodash-es/isArray';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';
import { ParamsListTasks } from 'projects/utask-lib/src/lib/@services/api.service';
import { TaskService } from 'projects/utask-lib/src/lib/@services/task.service';
import { TaskState, TaskType } from 'projects/utask-lib/src/lib/@models/task.model';
import { NzNotificationService } from 'ng-zorro-antd/notification';

@Component({
  templateUrl: './tasks.html',
  styleUrls: ['./tasks.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class TasksComponent implements OnInit {
  tags: string[] = [];
  meta: Meta = null;
  pagination: ParamsListTasks = new ParamsListTasks();

  constructor(
    private _activateRoute: ActivatedRoute,
    private router: Router,
    private taskService: TaskService,
    private _cd: ChangeDetectorRef,
    private _notif: NzNotificationService
  ) { }

  ngOnInit() {
    this.tags = this.taskService.getTagsRaw();
    this.taskService.tags.asObservable().subscribe((tags: string[]) => {
      this.tags = tags;
    });
    this.meta = this._activateRoute.snapshot.data.meta as Meta;
    this._activateRoute.queryParams.subscribe(params => {
      this.pagination = this.queryToSearchTask(params);
      this._cd.markForCheck();
    });
    this.search(true);
  }

  routeTo(taskId: string) {
    this.router.navigate(['/task/' + taskId]);
  }

  toastError(message: string) {
    this._notif.error('', message);
  }

  toastInfo(message: string) {
    this._notif.info('', message);
  }

  search(replaceUrl: boolean = false) {
    let cleanParams = {};
    Object.keys(this.pagination).filter(key => {
      if (isArray(this.pagination[key])) {
        return this.pagination[key].length > 0;
      }
      return !!this.pagination[key];
    }).forEach(key => {
      cleanParams[key] = isArray(this.pagination[key]) ? JSON.stringify(this.pagination[key]) : this.pagination[key];
    });
    this.router.navigate([], {
      queryParams: cleanParams,
      replaceUrl, // Useful to prevent adding a new entry in router history and keep the back/next navigation working
    });
  }

  queryToSearchTask(queryParams): ParamsListTasks {
    const params = new ParamsListTasks();
    const pageSize = parseInt(queryParams.page_size, 10)
    params.page_size = pageSize && 10 <= pageSize && pageSize <= 1000 ? pageSize : 10;
    const defaultType = this.meta.user_is_admin ? 'all' : 'own';
    params.type = queryParams.type || defaultType;
    params.last = '';
    params.state = queryParams.state || '';
    params.tag = queryParams.tag ? JSON.parse(queryParams.tag) : [];
    return params;
  }

  paginationTypeChange(type: TaskType) {
    this.pagination.type = type;
    this.search();
  }

  paginationStateChange(state: TaskState) {
    this.pagination.state = state;
    this.search();
  }

  inputTagsChanged(tags: Array<string>) {
    this.pagination.tag = tags;
    this.search();
  }
}
