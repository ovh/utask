import { Component, Input, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import { ApiService } from '../../@services/api.service';

@Component({
  selector: 'app-modal-confirmation-api',
  templateUrl: './modal-confirmation-api.component.html'
})
export class ModalConfirmationApiComponent {
  @Input() question: string;
  @Input() title: string;
  @Input() yes: string;
  @Input() apiCall: any;
  errors: any[];
  loading = false;
  error = null;

  constructor(public activeModal: NgbActiveModal) {
  }

  submit() {
    this.loading = true;
    this.apiCall().subscribe((data: any) => {
      this.error = null;
      this.activeModal.close(data);
    }, (err: any) => {
      this.error = err;
    }).add(() => {
      this.loading = false;
    });
  }
}
