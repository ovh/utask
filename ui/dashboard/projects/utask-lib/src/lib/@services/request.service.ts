import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import Task from '../@models/task.model';
import Meta from '../@models/meta.model';
import { ModalApiYamlEditComponent } from '../@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ApiService } from './api.service';

@Injectable()
export class RequestService {

    constructor(private modalService: NgbModal, private api: ApiService) { }

    edit(task: Task) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalApiYamlEditComponent, {
                size: 'xl'
            });
            modal.componentInstance.title = 'Edit request';
            modal.componentInstance.apiCall = () => this.api.task.getAsYaml(task.id).toPromise();
            modal.componentInstance.apiCallSubmit = (data: any) => this.api.task.updateAsYaml(task.id, data).toPromise();
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

    isResolvable(task: Task, meta: Meta, allowedResolverUsernames: string[]): boolean {
        return !task.resolution && task.state !== 'WONTFIX' &&
            (
                meta.user_is_admin || (allowedResolverUsernames || []).indexOf(meta.username) > -1 || (task.resolver_usernames || []).indexOf(meta.username) > -1
            );
    }
}
