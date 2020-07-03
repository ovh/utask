import { of } from 'rxjs';
import { Component, OnInit, OnDestroy, NgZone } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import { ToastrService } from 'ngx-toastr';

import { ApiService, ParamsListTasks } from 'utask-lib';
import Meta from 'utask-lib/@models/meta.model';
import { TaskService } from 'utask-lib';

@Component({
  templateUrl: './home.html',
  styleUrls: ['./home.sass'],
})
export class HomeComponent implements OnInit, OnDestroy {
  tags: string[] = [];
  meta: Meta = null;
  pagination: ParamsListTasks;
  params: ParamsListTasks;

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router, private taskService: TaskService, private zone: NgZone, private toastr: ToastrService) {
  }

  ngOnInit() {
    this.tags = _.clone(this.taskService.tagsRaw);
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
    this.pagination.tag = _.compact(text.split(' '));
    this.search();
  }

  ngOnDestroy() {
  }

  search() {
    this.zone.run(() => {
      this.params = _.clone(this.pagination);
    });

    this.router.navigate([], {
      queryParams: this.pagination,
      queryParamsHandling: 'merge'
    });
  }

  queryToSearchTask(p?: any): ParamsListTasks {
    const params = _.clone(p || this.router.routerState.snapshot.root.queryParams);
    if (params.tag && _.isString(params.tag)) {
      params.tag = [params.tag];
    }
    const item = new ParamsListTasks();
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
}
