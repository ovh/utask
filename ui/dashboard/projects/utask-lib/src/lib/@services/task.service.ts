import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalConfirmationApiComponent } from '../@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ApiService } from './api.service';
import get from 'lodash-es/get';
import { allSettled } from 'q';
import { Subject } from 'rxjs';
import Task from '../@models/task.model';
import environment from '../@services/config';
import { clone } from 'lodash-es';
import { NzModalService } from 'ng-zorro-antd/modal';

@Injectable()
export class TaskService {
  private localStorageTags = `${environment.localStorage}tags`;
  private tagsRaw: string[] = [];
  public tags = new Subject<string[]>();

  constructor(
    private modalService: NgbModal,
    private api: ApiService,
    private modal: NzModalService
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

  delete(taskId: string) {
    return new Promise((resolve, reject) => {
      this.modal.confirm({
        nzTitle: '<i>Are you sure you want to delete this task?</i>',
        nzContent: `Task ID: ${taskId}`,
        nzOkText: 'Yes',
        nzOkType: 'danger',
        nzOnOk: () => this.api.task.delete(taskId).subscribe(() => resolve()),
        nzCancelText: 'No',
        nzOnCancel: () => reject()
      });
    });
  }

  deleteAll(taskIds: string[]) {
    return new Promise((resolve, reject) => {
      const modal = this.modalService.open(ModalConfirmationApiComponent, {
        size: 'xl'
      });
      modal.componentInstance.question = `Are you sure you want to delete these ${taskIds.length} tasks ?`;
      modal.componentInstance.title = `Delete tasks`;
      modal.componentInstance.yes = `Yes, I'm sure`;
      modal.componentInstance.apiCall = () => {
        return new Promise((resolve, reject) => {
          const promises = [];
          taskIds.forEach((id) => {
            promises.push(this.api.task.delete(id).toPromise());
          });
          allSettled(promises).then((data: any[]) => {
            const tasksInError = [];
            taskIds.forEach((id, i) => {
              if (data[i].state === 'rejected') {
                tasksInError.push(id);
              }
            });
            if (tasksInError.length) {
              reject(`An error occured when trying to delete the task(s) '${tasksInError.join('\', \'')}'.`);
            } else {
              resolve(data);
            }
          }).catch((err) => {
            reject(err);
          });
        });
      };
      modal.componentInstance.dismiss = () => {
        modal.dismiss();
      };
      modal.componentInstance.close = () => {
        modal.close();
      };
      modal.result.then((res: any) => {
        resolve(res);
      }).catch((err) => {
        if (err !== 0 && err !== 1 && err !== 'Cross click') {
          reject(err);
        } else {
          reject('close');
        }
      });
    });
  }
}
