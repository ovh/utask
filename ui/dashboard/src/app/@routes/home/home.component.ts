import { of } from 'rxjs';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../@services/api.service';
import * as _ from 'lodash';
import MetaUtask from 'src/app/@models/meta-utask.model';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ResolutionService } from 'src/app/@services/resolution.service';
import { TaskService } from 'src/app/@services/task.service';
import { delay, repeat } from 'rxjs/operators';

export class SearchTask {
  page_size?: number;
  last?: string;
  type?: string;
  state?: string;
}

@Component({
  templateUrl: './home.html',
})
export class HomeComponent implements OnInit {
  loaders: { [key: string]: boolean } = {};
  errors: { [key: string]: any } = {};
  meta: MetaUtask = null;
  tasks: any = [];
  pagination: SearchTask = {};
  hasMore = true;
  percentages: { [key: string]: number } = {};

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router, private modalService: NgbModal, private resolutionService: ResolutionService, private taskService: TaskService) {
  }

  loadTask(id: string, times: number = 1, delayMillisecond: number = 2000) {
    if (!this.loaders[`task${id}`]) {
      this.loaders[`task${id}`] = true;
      this.api.task(id).toPromise().then(data => {
        const index = _.findIndex(this.tasks, { id });
        this.tasks.splice(index, 1, data);
      }).finally(() => {
        this.loaders[`task${id}`] = false;
      });
    }
    of(id).pipe(delay(delayMillisecond)).pipe(repeat(times)).subscribe((id: string) => {
      if (!this.loaders[`task${id}`]) {
        this.loaders[`task${id}`] = true;
        this.api.task(id).toPromise().then(data => {
          const index = _.findIndex(this.tasks, { id });
          this.tasks.splice(index, 1, data);
        }).finally(() => {
          this.loaders[`task${id}`] = false;
        });
      }
    });
  }

  runResolution(resolutionId: string, taskId: string) {
    this.resolutionService.run(resolutionId).then((data: any) => {
      this.loadTask(taskId, 4, 2500);
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  deleteTask(id: string) {
    this.taskService.delete(id).then((data: any) => {
      _.remove(this.tasks, {
        id
      });
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  pauseResolution(resolutionId: string, taskId: string) {
    this.resolutionService.pause(resolutionId).then((data: any) => {
      this.loadTask(taskId, 4, 2500);
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  cancelResolution(resolutionId: string, taskId: string) {
    this.resolutionService.cancel(resolutionId).then((data: any) => {
      this.loadTask(taskId, 0, 2500);
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  extendResolution(resolutionId: string, taskId: string) {
    this.resolutionService.extend(resolutionId).then((data: any) => {
      this.loadTask(taskId, 4, 2500);
    }).catch((err) => {
      if (err !== 0) {
        console.log(err);
      }
    });
  }

  // previewDetails(obj: any, title: string) {
  //   const previewModal = this.modalService.open(ModalYamlPreviewComponent, {
  //     size: 'xl'
  //   });
  //   previewModal.componentInstance.value = obj;
  //   previewModal.componentInstance.title = title;
  // }

  ngOnInit() {
    this.meta = this.route.parent.snapshot.data.meta as MetaUtask;
    this.route.queryParams.subscribe(params => {
      this.pagination = this.queryToSearchTask(params);
      this.loadTasks();
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
    item.type = params.type ? params.type : 'own';
    item.last = '';
    item.state = params.state ? params.state : '';
    return item;
  }

  next(force: boolean = false) {
    if (!this.loaders.next && (this.hasMore || force)) {
      this.loadTasks(_.last(this.tasks).id);
    }
  }

  search() {
    if (_.isEqual(this.pagination, this.queryToSearchTask())) {
      this.loadTasks();
    } else {
      this.router.navigate([], {
        queryParams: this.pagination,
        queryParamsHandling: 'merge'
      });
    }
  }

  loadTasks(last: string = '') {
    if (last) {
      this.loaders.next = true;
    } else {
      this.loaders.tasks = true;
    }
    const params: SearchTask = _.clone(this.pagination);
    params.last = last;

    this.api.tasks(params).subscribe((data) => {
      if (params.last) {
        this.tasks = this.tasks.concat(data.body);
      } else {
        this.tasks = data.body;
      }

      (data.body as any[]).forEach((task: any) => {
        this.percentages[task.id] = Math.round(task.steps_done / task.steps_total * 100);
      });
      this.errors.tasks = null;
      this.hasMore = (data.body as any[]).length === this.pagination.page_size;
    }, (err: any) => {
      if (last) {
        this.hasMore = false;
      } else {
        this.errors.tasks = err;
      }
    }).add(() => {
      this.loaders.tasks = false;
      this.loaders.next = false;
    });
  }
}
