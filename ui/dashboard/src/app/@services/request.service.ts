import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiService } from './api.service';
import { ModalEditRequestComponent } from '../@modals/modal-edit-request/modal-edit-request.component';
import MetaUtask from '../@models/meta-utask.model';
import Task from '../@models/task.model';

@Injectable()
export class RequestService {

    constructor(private modalService: NgbModal) { }

    edit(task: Task) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalEditRequestComponent, {
                size: 'xl'
            });
            modal.componentInstance.value = task;
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

    isResolvable(task: Task, meta: MetaUtask, allowedResolverUsernames: string[]): boolean {
        return !task.resolution && task.state !== 'WONTFIX' &&
            (
                meta.user_is_admin || (allowedResolverUsernames || []).indexOf(meta.username) > -1 || (task.resolver_usernames || []).indexOf(meta.username) > -1
            );
    }
}
