import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalConfirmationApiComponent } from '../@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ApiService } from './api.service';

@Injectable()
export class TaskService {

    constructor(private modalService: NgbModal, private api: ApiService) { }

    delete(taskId: string) {
        return new Promise((resolve, reject) => {
          const modal = this.modalService.open(ModalConfirmationApiComponent, {
            size: 'xl'
          });
          modal.componentInstance.question = `Are you sure you want to delete this task #${taskId}`;
          modal.componentInstance.title = `Delete task`;
          modal.componentInstance.yes = `Yes, i'm sure`;
          modal.componentInstance.apiCall = () => {
            return this.api.deleteTask(taskId);
          };
          modal.result.then((res: any) => {
            resolve(res);
          }).catch((err) => {
            reject(err);
          });
        });
      }
}
