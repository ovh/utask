import { Component } from '@angular/core';
@Component({
    selector: 'loader',
    template: `
    <div class="d-flex justify-content-center">
        <div class="spinner-border" role="status">
        <span class="sr-only">Loading...</span>
        </div>
    </div>
    `,
})
export class LoaderComponent {
    constructor() { }
};