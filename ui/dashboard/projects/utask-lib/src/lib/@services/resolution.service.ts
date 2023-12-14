import { Injectable } from '@angular/core';
import { ApiService } from './api.service';
import { ModalApiYamlEditComponent } from '../@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ModalService } from './modal.service';
import { throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { NzModalService } from 'ng-zorro-antd/modal';

@Injectable()
export class ResolutionService {

    constructor(
        private modal: NzModalService,
        private _modalService: ModalService,
        private api: ApiService
    ) { }

    pause(resolutionId: string) {
        return this.api.resolution.pause(resolutionId).toPromise();
    }

    pauseAll(resolutionIds: string[]) {
        return this._modalService.confirmAll(
            `<i>Are you sure you want to pause these ${resolutionIds.length} tasks?</i>`,
            '',
            'default',
            true,
            ...resolutionIds.map(id => {
                return this.api.resolution.pause(id)
                    .pipe(catchError(() => throwError(`Can't pause the resolution with id: ${id}`)));
            })
        );
    }

    cancel(resolutionId: string): Promise<any> {
        return this._modalService.confirm(
            '<i>Are you sure you want to cancel this resolution?</i>',
            `Resolution ID: ${resolutionId}`,
            'default',
            true,
            this.api.resolution.cancel(resolutionId)
        );
    }

    cancelAll(resolutionIds: string[]) {
        return this._modalService.confirmAll(
            `<i>Are you sure you want to cancel these ${resolutionIds.length} tasks?</i>`,
            '',
            'default',
            true,
            ...resolutionIds.map(id => {
                return this.api.resolution.cancel(id)
                    .pipe(catchError(() => throwError(`Can't cancel the resolution with id: ${id}`)));
            })
        );
    }

    extend(resolutionId: string) {
        return this.api.resolution.extend(resolutionId).toPromise();
    }

    extendAll(resolutionIds: string[]) {
        return this._modalService.confirmAll(
            `<i>Are you sure you want to extend these ${resolutionIds.length} tasks?</i>`,
            '',
            'primary',
            false,
            ...resolutionIds.map(id => {
                return this.api.resolution.extend(id)
                    .pipe(catchError(() => throwError(`Can't extend the resolution with id: ${id}`)));
            })
        );
    }

    run(resolutionId: string) {
        return this.api.resolution.run(resolutionId).toPromise();
    }

    runAll(resolutionIds: string[]) {
        return this._modalService.confirmAll(
            `<i>Are you sure you want to run these ${resolutionIds.length} tasks?</i>`,
            '',
            'primary',
            false,
            ...resolutionIds.map(id => {
                return this.api.resolution.run(id)
                    .pipe(catchError(() => throwError(`Can't run the resolution with id: ${id}`)));
            })
        );
    }

    edit(resolution: any) {
        return new Promise((resolve, reject) => {
            this.modal.create({
                nzTitle: 'Resolution preview',
                nzContent: ModalApiYamlEditComponent,
                nzWidth: '80%',
                nzData: {
                    apiCall: () => this.api.resolution.getAsYaml(resolution.id).toPromise(),
                    apiCallSubmit: (data: any) => this.api.resolution.updateAsYaml(resolution.id, data).toPromise()
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
}
