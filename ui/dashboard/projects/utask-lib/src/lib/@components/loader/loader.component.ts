import { Component } from '@angular/core';
@Component({
    selector: 'utask-loader',
    template: `
        <div style="text-align: center"><nz-spin nzSimple></nz-spin></div>
    `,
})
export class LoaderComponent {
    constructor() { }
};