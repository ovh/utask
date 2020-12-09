import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import get from 'lodash-es/get';
import { Subject, throwError } from 'rxjs';
import Task from '../@models/task.model';
import environment from '../@services/config';
import { ModalService } from './modal.service';
import { catchError } from 'rxjs/operators';
import clone from 'lodash-es/clone';

@Injectable()
export class TaskService {
  private localStorageTags = `${environment.localStorage}tags`;
  private tagsRaw: string[] = [];
  public tags = new Subject<string[]>();

  constructor(
    private _modalService: ModalService,
    private api: ApiService
  ) {
    this.tagsRaw = localStorage.getItem(this.localStorageTags) ? JSON.parse(localStorage.getItem(this.localStorageTags)) : [];
  }

  getTagsRaw(): Array<string> {
    return clone(this.tagsRaw);
  }

  registerTags(task: Task): void {
    let hasNewTags = false;
    const tags = Object.keys(get(task, 'tags', {}));
    tags.forEach((t: string) => {
      if (this.tagsRaw.indexOf(t) === -1) {
        this.tagsRaw.push(t);
        hasNewTags = true;
      }
    });
    if (hasNewTags) {
      this.tags.next(this.tagsRaw);
      localStorage.setItem(this.localStorageTags, JSON.stringify(this.tagsRaw));
    }
  }

  delete(taskId: string): Promise<any> {
    return this._modalService.confirm(
      '<i>Are you sure you want to delete this task?</i>',
      `Task ID: ${taskId}`,
      'danger',
      this.api.task.delete(taskId)
    );
  }

  deleteAll(taskIds: string[]) {
    return this._modalService.confirmAll(
      `<i>Are you sure you want to delete these ${taskIds.length} tasks?</i>`,
      '',
      'danger',
      ...taskIds.map(id => {
        return this.api.task.delete(id)
          .pipe(catchError(() => throwError(`Can't delete the task with id: ${id}`)));
      })
    );
  }
}
