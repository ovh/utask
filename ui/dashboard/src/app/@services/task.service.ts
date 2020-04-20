import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalConfirmationApiComponent } from '../@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ApiService } from './api.service';
import * as _ from 'lodash';
import { allSettled } from 'q';
import { environment } from 'src/environments/environment';
import { Subject } from 'rxjs';
import Task from '../@models/task.model';

@Injectable()
export class TaskService {
  private localStorageTags = `${environment.localStorage}tags`;
  public tagsRaw: string[] = localStorage.getItem(this.localStorageTags) ? JSON.parse(localStorage.getItem(this.localStorageTags)) : [];
  public tags = new Subject<string[]>();

  constructor(private modalService: NgbModal, private api: ApiService) {
  }

  registerTags(task: Task): any {
    let hasNewTags = false;
    const tags = Object.keys(_.get(task, 'tags', {}));
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
    return task;
  }

  delete(taskId: string) {
    return new Promise((resolve, reject) => {
      const modal = this.modalService.open(ModalConfirmationApiComponent, {
        size: 'xl'
      });
      modal.componentInstance.question = `Are you sure you want to delete this task #${taskId} ?`;
      modal.componentInstance.title = `Delete task`;
      modal.componentInstance.yes = `Yes, I'm sure`;
      modal.componentInstance.apiCall = () => {
        return this.api.deleteTask(taskId).toPromise();
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
            promises.push(this.api.deleteTask(id).toPromise());
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
