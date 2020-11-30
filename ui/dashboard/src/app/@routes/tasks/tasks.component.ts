import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import compact from 'lodash-es/compact';
import isString from 'lodash-es/isString';
import isArray from 'lodash-es/isArray';
import { ToastrService } from 'ngx-toastr';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';
import { ParamsListTasks } from 'projects/utask-lib/src/lib/@services/api.service';
import { TaskService } from 'projects/utask-lib/src/lib/@services/task.service';

@Component({
  templateUrl: './tasks.html',
  styleUrls: ['./tasks.sass'],
})
export class TasksComponent implements OnInit {
  tags: string[] = [];
  meta: Meta = null;
  pagination: ParamsListTasks;

  constructor(
    private _activateRoute: ActivatedRoute,
    private router: Router,
    private taskService: TaskService,
    private toastr: ToastrService
  ) { }

  ngOnInit() {
    this.tags = this.taskService.getTagsRaw();
    this.taskService.tags.asObservable().subscribe((tags: string[]) => {
      this.tags = tags;
    });
    this.meta = this._activateRoute.snapshot.data.meta as Meta;
    this._activateRoute.queryParams.subscribe(params => {
      this.pagination = this.queryToSearchTask(params);
      this.search();
    });
  }

  routeTo(taskId: string) {
    this.router.navigate(['/task/' + taskId]);
  }

  toastError(message: string) {
    this.toastr.error(message);
  }

  toastInfo(message: string) {
    this.toastr.info(message);
  }

  inputTagsChanged(text: string) {
    this.pagination.tag = compact(text.split(' '));
    this.search();
  }

  search() {
    let cleanParams = {};
    Object.keys(this.pagination).filter(key => {
      if (isArray(this.pagination[key])) {
        return this.pagination[key].length > 0;
      }
      return !!this.pagination[key];
    }).forEach(key => {
      cleanParams[key] = this.pagination[key];
    });
    this.router.navigate([], {
      queryParams: cleanParams,
      queryParamsHandling: 'merge',
      replaceUrl: true // Useful to prevent adding a new entry in router history and keep the back/next navigation working
    });
  }

  queryToSearchTask(params): ParamsListTasks {
    const item = new ParamsListTasks();
    const pageSize = parseInt(params.page_size, 10)
    item.page_size = pageSize && 10 <= pageSize && pageSize <= 1000 ? pageSize : 10;
    const defaultType = this.meta.user_is_admin ? 'all' : 'own';
    item.type = params.type || defaultType;
    item.last = '';
    item.state = params.state || '';
    if (params.tag && isString(params.tag)) {
      params.tag = [params.tag];
    }
    item.tag = params.tag || [];
    return item;
  }
}
