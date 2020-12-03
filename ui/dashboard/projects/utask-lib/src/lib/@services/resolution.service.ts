import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiService } from './api.service';
import { ModalApiYamlEditComponent } from '../@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ModalService } from './modal.service';
import { throwError } from 'rxjs';
import { catchError } from 'rxjs/operators';

@Injectable()
export class ResolutionService {

    constructor(
        private modalService: NgbModal,
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
            'danger',
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
            'danger',
            this.api.resolution.cancel(resolutionId)
        );
    }

    cancelAll(resolutionIds: string[]) {
        return this._modalService.confirmAll(
            `<i>Are you sure you want to cancel these ${resolutionIds.length} tasks?</i>`,
            '',
            'danger',
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
            ...resolutionIds.map(id => {
                return this.api.resolution.run(id)
                    .pipe(catchError(() => throwError(`Can't run the resolution with id: ${id}`)));
            })
        );
    }

    edit(resolution: any) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalApiYamlEditComponent, {
                size: 'xl'
            });
            modal.componentInstance.title = 'Edit resolution';
            modal.componentInstance.apiCall = () => this.api.resolution.getAsYaml(resolution.id).toPromise();
            modal.componentInstance.apiCallSubmit = (data: any) => this.api.resolution.updateAsYaml(resolution.id, data).toPromise();
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
