import { Component, Input, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import * as _ from 'lodash';

@Component({
  selector: 'app-modal-confirmation-api',
  templateUrl: './modal-confirmation-api.component.html'
})
export class ModalConfirmationApiComponent {
  @Input() question: string;
  @Input() title: string;
  @Input() yes: string;
  @Input() apiCall: any;
  loading = false;
  error = null;

  constructor(public activeModal: NgbActiveModal) {
  }

  submit() {
    this.loading = true;
    this.apiCall().then((data: any) => {
      this.error = null;
      this.activeModal.close(data);
    }).catch((err: any) => {
      console.log(err);
      if (err) {
        this.error = err;
      }
    }).finally(() => {
      this.loading = false;
    });
  }
}
