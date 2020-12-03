import { Component, Input } from '@angular/core';

@Component({
	selector: 'lib-modal-content-with-error',
  templateUrl: './modal-content-with-error.html',
	styleUrls: ['./modal-content-with-error.sass']
})
export class NzModalContentWithErrorComponent {
	@Input() content?: string;
	@Input() errors?: Array<any>;

	constructor() { }
}