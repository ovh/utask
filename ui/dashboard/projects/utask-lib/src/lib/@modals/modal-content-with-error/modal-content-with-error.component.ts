import { Component, Input } from '@angular/core';

interface IModalData {
	content?: string;
	errors?: Array<any>;
}

@Component({
    selector: 'lib-modal-content-with-error',
    templateUrl: './modal-content-with-error.html',
    styleUrls: ['./modal-content-with-error.sass'],
    standalone: false
})
export class NzModalContentWithErrorComponent {
	@Input() content?: string;
	@Input() errors?: Array<any>;

	constructor() { }
}