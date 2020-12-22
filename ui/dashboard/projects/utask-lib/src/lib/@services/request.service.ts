import { Injectable } from '@angular/core';
import Task from '../@models/task.model';
import Meta from '../@models/meta.model';
import { ModalApiYamlEditComponent } from '../@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ApiService } from './api.service';
import { NzModalService } from 'ng-zorro-antd/modal';

@Injectable()
export class RequestService {

    constructor(
        private modal: NzModalService,
        private api: ApiService) {
    }

    edit(task: Task) {
        return new Promise((resolve, reject) => {
            this.modal.create({
                nzTitle: 'Resolution preview',
                nzContent: ModalApiYamlEditComponent,
                nzWidth: '80%',
                nzComponentParams: {
                    apiCall: () => this.api.task.getAsYaml(task.id).toPromise(),
                    apiCallSubmit: (data: any) => this.api.task.updateAsYaml(task.id, data).toPromise()
                },
                nzOnOk: (data: ModalApiYamlEditComponent) => {
                    resolve(data.result);
                },
                nzOnCancel: () => {
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
