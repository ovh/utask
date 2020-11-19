import { Component, OnInit, OnDestroy, NgZone } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import compact from 'lodash-es/compact';
import clone from 'lodash-es/clone';
import isString from 'lodash-es/isString';
import isNumber from 'lodash-es/isNumber';
import { ToastrService } from 'ngx-toastr';
import Meta from 'projects/utask-lib/src/lib/@models/meta.model';
import { ApiService, ParamsListTasks } from 'projects/utask-lib/src/lib/@services/api.service';
import { TaskService } from 'projects/utask-lib/src/lib/@services/task.service';

@Component({
  templateUrl: './home.html',
  styleUrls: ['./home.sass'],
})
export class HomeComponent implements OnInit, OnDestroy {
  tags: string[] = [];
  meta: Meta = null;
  pagination: ParamsListTasks;
  params: ParamsListTasks;

  constructor(
    private api: ApiService,
    private route: ActivatedRoute,
    private router: Router,
    private taskService: TaskService,
    private zone: NgZone,
    private toastr: ToastrService
  ) { }

  ngOnInit() {
    this.tags = clone(this.taskService.tagsRaw);
    this.taskService.tags.asObservable().subscribe((tags: string[]) => {
      this.tags = tags;
    });

    this.meta = this.route.parent.snapshot.data.meta as Meta;
    this.route.queryParams.subscribe(params => {
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

  ngOnDestroy() {
  }

  search() {
    this.zone.run(() => {
      this.params = clone(this.pagination);
    });

    this.router.navigate([], {
      queryParams: this.pagination,
      queryParamsHandling: 'merge'
    });
  }

  queryToSearchTask(p?: any): ParamsListTasks {
    const params = clone(p || this.router.routerState.snapshot.root.queryParams);
    if (params.tag && isString(params.tag)) {
      params.tag = [params.tag];
    }
    const item = new ParamsListTasks();
    if (params.itemPerPage && isNumber(+params.itemPerPage) && +params.itemPerPage <= 1000 && +params.itemPerPage >= 10) {
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
}
