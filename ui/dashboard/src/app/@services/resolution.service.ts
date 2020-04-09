import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalConfirmationApiComponent } from '../@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ApiService } from './api.service';
import { ModalEditResolutionComponent } from '../@modals/modal-edit-resolution/modal-edit-resolution.component';

@Injectable()
export class ResolutionService {

    constructor(private modalService: NgbModal, private api: ApiService) { }

    pause(resolutionId: string) {
        return this.api.pauseResolution(resolutionId).toPromise();
    }

    cancel(resolutionId: string) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to cancel this resolution #${resolutionId}`;
            modal.componentInstance.title = `Cancel resolution`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return this.api.cancelResolution(resolutionId);
            };
            modal.result.then((res: any) => {
                resolve(res);
            }).catch((err) => {
                reject(err);
            });
        });
    }

    extend(resolutionId: string) {
        return this.api.extendResolution(resolutionId).toPromise();
    }

    run(resolutionId: string) {
        return this.api.runResolution(resolutionId).toPromise();
    }

    edit(resolution: any) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalEditResolutionComponent, {
                size: 'xl'
            });
            modal.componentInstance.value = resolution;
            modal.result.then((res: any) => {
                resolve(res);
            }).catch((err) => {
                reject(err);
            });
        });
    }
}
