import { NgModule, ErrorHandler } from '@angular/core';
import { faUserShield, faCheckCircle, faTimesCircle, faBan, faHistory, faSync, faHourglassHalf, faQuestionCircle, faCaretDown, faCaretUp } from '@fortawesome/fontawesome-free-solid';
import fontawesome from '@fortawesome/fontawesome';
fontawesome.library.add(faUserShield, faCheckCircle, faTimesCircle, faBan, faHistory, faSync, faHourglassHalf, faQuestionCircle, faCaretDown, faCaretUp);
import { AppRoutingModule } from './app-routing.module';
import { TagInputModule } from 'ngx-chips';
TagInputModule.withDefaults({
  tagInput: {
    placeholder: 'Add filter',
    secondaryPlaceholder: 'Filter steps'
  }
});
import { AppComponent } from './app.component';
import { MyErrorHandler } from './handlers/error.handler';
import { UTaskModule } from 'src/utask-module/utask.module';

const pages = [
  AppComponent,
];

@NgModule({
  declarations: pages,
  imports: [
    AppRoutingModule,
  ],
  providers: [{ provide: ErrorHandler, useClass: MyErrorHandler }],
  bootstrap: [AppComponent],
  entryComponents: [
  ]
})
export class AppModule { }
