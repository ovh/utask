import { Component, Input } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'lib-modal-confirmation-api',
  templateUrl: './modal-confirmation-api.component.html'
})
export class ModalConfirmationApiComponent {
  @Input() question: string;
  @Input() warning: string;
  @Input() title: string;
  @Input() yes: string;
  @Input() apiCall: any;

  loading = false;
  error = null;

  constructor(
    public activeModal: NgbActiveModal
  ) { }

  submit() {
    this.loading = true;
    this.apiCall().then((data) => {
      this.error = null;
      this.activeModal.close(data);
    }).catch((err: any) => {
      this.error = err;
    }).finally(() => {
      this.loading = false;
    });
  }
}