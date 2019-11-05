import { ErrorHandler } from '@angular/core';

export class MyErrorHandler implements ErrorHandler {

    handleError(error: any) {
        console.error('🤖 Development error 🤖');
        console.error('Please bring back the problem to the development team');
        console.log(error);
    }
}
