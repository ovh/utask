import { Injectable } from '@angular/core';
import { NzButtonType } from 'ng-zorro-antd/button';
import { NzModalService } from 'ng-zorro-antd/modal';
import { forkJoin, Observable, of } from 'rxjs';
import { catchError } from 'rxjs/operators';
import { NzModalContentWithErrorComponent } from '../@modals/modal-content-with-error/modal-content-with-error.component';

@Injectable()
export class ModalService {
	constructor(
		private _modal: NzModalService
	) { }

	confirmAll<T>(title: string, content: string, okBtnType: NzButtonType, ...actions: Array<Observable<T>>): Promise<Array<T>> {
		return new Promise((resolveModal, _) => {
			this._modal.confirm({
				nzTitle: title,
				nzContent: NzModalContentWithErrorComponent,
				nzComponentParams: { content },
				nzOkText: 'Yes',
				nzCancelText: 'No',
				nzOkType: okBtnType,
				nzOnOk: modal => {
					let resolveClose: any;
					modal.errors = [];
					const closeModal = new Promise(res => { resolveClose = res; })
					forkJoin(actions.map(a => a.pipe(catchError(err => of({ isError: true, error: err })))))
						.subscribe(results => {
							const errs = results.filter(res => !!res && (res as any).isError).map(res => (res as any).error);
							if (errs.length > 0) {
								resolveClose(false);
								modal.errors = errs;
							} else {
								resolveClose(true);
								resolveModal(results as Array<T>);
							}
						})
					return closeModal;
				}
			});
		});
	}

	confirm<T>(title: string, content: string, okBtnType: NzButtonType, action: Observable<T>): Promise<T> {
		return new Promise((resolveModal, _) => {
			this._modal.confirm({
				nzTitle: title,
				nzContent: NzModalContentWithErrorComponent,
				nzComponentParams: { content },
				nzOkText: 'Yes',
				nzCancelText: 'No',
				nzOkType: okBtnType,
				nzOnOk: modal => {
					let resolveClose: any;
					modal.errors = [];
					const closeModal = new Promise(res => { resolveClose = res; })
					action.pipe(catchError(err => of({ isError: true, error: err })))
						.subscribe(result => {
							if (result && (result as any).isError) {
								resolveClose(false);
								modal.errors = [(result as any).error];
							} else {
								resolveClose(true);
								resolveModal(result as T);
							}
						})
					return closeModal;
				}
			});
		});
	}
}
