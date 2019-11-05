import { Component } from '@angular/core';

@Component({
  template: `
    <div>
      <div class="alert alert-danger">
        <strong>Error !</strong> An error just occured, please refresh the page or contact the administrators.
      </div>
      <a class="btn btn-link" routerLink="/home">Back to Home</a>
    </div>
  `,
})
export class ErrorComponent {
  constructor() {
  }
}
