import { Injectable } from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ModalConfirmationApiComponent } from '../@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ApiService } from './api.service';
import { ModalEditResolutionComponent } from '../@modals/modal-edit-resolution/modal-edit-resolution.component';
import { allSettled } from 'q';

@Injectable()
export class ResolutionService {

    constructor(private modalService: NgbModal, private api: ApiService) { }

    pause(resolutionId: string) {
        return this.api.pauseResolution(resolutionId).toPromise();
    }

    pauseAll(resolutionIds: string[]) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to pause these ${resolutionIds.length} tasks ?`;
            modal.componentInstance.title = `Pause tasks`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return new Promise((resolve, reject) => {
                    const promises = [];
                    resolutionIds.forEach((id) => {
                        promises.push(this.api.pauseResolution(id).toPromise());
                    });
                    allSettled(promises).then((data: any[]) => {
                        const resolutionsInError = [];
                        resolutionIds.forEach((id, i) => {
                            if (data[i].state === 'rejected') {
                                resolutionsInError.push(id);
                            }
                        });
                        if (resolutionsInError.length) {
                            reject(`An error occured when trying to pause the resolution(s) '${resolutionsInError.join('\', \'')}'.`);
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

    cancel(resolutionId: string) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to cancel this resolution #${resolutionId}`;
            modal.componentInstance.title = `Cancel resolution`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return this.api.cancelResolution(resolutionId).toPromise();
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

    cancelAll(resolutionIds: string[]) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to cancel these ${resolutionIds.length} tasks ?`;
            modal.componentInstance.title = `Cancel tasks`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return new Promise((resolve, reject) => {
                    const promises = [];
                    resolutionIds.forEach((id) => {
                        promises.push(this.api.cancelResolution(id).toPromise());
                    });
                    allSettled(promises).then((data: any[]) => {
                        const resolutionsInError = [];
                        resolutionIds.forEach((id, i) => {
                            if (data[i].state === 'rejected') {
                                resolutionsInError.push(id);
                            }
                        });
                        if (resolutionsInError.length) {
                            reject(`An error occured when trying to cancel the resolution(s) '${resolutionsInError.join('\', \'')}'.`);
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

    extend(resolutionId: string) {
        return this.api.extendResolution(resolutionId).toPromise();
    }

    extendAll(resolutionIds: string[]) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to extend these ${resolutionIds.length} tasks ?`;
            modal.componentInstance.title = `Extend tasks`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return new Promise((resolve, reject) => {
                    const promises = [];
                    resolutionIds.forEach((id) => {
                        promises.push(this.api.extendResolution(id).toPromise());
                    });
                    allSettled(promises).then((data: any[]) => {
                        const resolutionsInError = [];
                        resolutionIds.forEach((id, i) => {
                            if (data[i].state === 'rejected') {
                                resolutionsInError.push(id);
                            }
                        });
                        if (resolutionsInError.length) {
                            reject(`An error occured when trying to extend the resolution(s) '${resolutionsInError.join('\', \'')}'.`);
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

    run(resolutionId: string) {
        return this.api.runResolution(resolutionId).toPromise();
    }

    runAll(resolutionIds: string[]) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalConfirmationApiComponent, {
                size: 'xl'
            });
            modal.componentInstance.question = `Are you sure you want to run these ${resolutionIds.length} tasks ?`;
            modal.componentInstance.title = `Run tasks`;
            modal.componentInstance.yes = `Yes, I'm sure`;
            modal.componentInstance.apiCall = () => {
                return new Promise((resolve, reject) => {
                    const promises = [];
                    resolutionIds.forEach((id) => {
                        promises.push(this.api.runResolution(id).toPromise());
                    });
                    allSettled(promises).then((data: any[]) => {
                        const resolutionsInError = [];
                        resolutionIds.forEach((id, i) => {
                            if (data[i].state === 'rejected') {
                                resolutionsInError.push(id);
                            }
                        });
                        if (resolutionsInError.length) {
                            reject(`An error occured when trying to run the resolution(s) '${resolutionsInError.join('\', \'')}'.`);
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

    edit(resolution: any) {
        return new Promise((resolve, reject) => {
            const modal = this.modalService.open(ModalEditResolutionComponent, {
                size: 'xl'
            });
            modal.componentInstance.value = resolution;
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
