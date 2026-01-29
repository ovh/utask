import { Component } from '@angular/core';

@Component({
    template: `
    <nz-result nzStatus="error" nzSubTitle="An error just occured, please refresh the page or contact the administrators.">
      <div nz-result-extra>
        <button nz-button nzType="primary" routerLink="/">Back to Home</button>
      </div>
    </nz-result>
  `,
    standalone: false
})
export class ErrorComponent {
  constructor() {
  }
}
