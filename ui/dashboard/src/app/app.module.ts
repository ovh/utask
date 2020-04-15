import { NgModule, ErrorHandler } from '@angular/core';
/*
  Toaster
*/
import { ToastrModule } from 'ngx-toastr';
/*
  Font AWESOME
*/
import { faUserShield, faCheckCircle, faTimesCircle, faBan, faHistory, faSync, faHourglassHalf, faQuestionCircle, faCaretDown, faCaretUp } from '@fortawesome/fontawesome-free-solid';
import fontawesome from '@fortawesome/fontawesome';
fontawesome.library.add(faUserShield, faCheckCircle, faTimesCircle, faBan, faHistory, faSync, faHourglassHalf, faQuestionCircle, faCaretDown, faCaretUp);

/*
  @Modules
*/
import { BrowserModule } from '@angular/platform-browser';
import { AppRoutingModule } from './app-routing.module';
import { HttpClientModule } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
/*
  Infinite Scroll
*/
import { InfiniteScrollModule } from 'ngx-infinite-scroll';
/*
  Modules NGX-CHIPS
*/
import { TagInputModule } from 'ngx-chips';
TagInputModule.withDefaults({
  tagInput: {
    placeholder: 'Add filter',
    secondaryPlaceholder: 'Filter steps'
  }
});
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
/*
  Modules NG-BOOTSTRAP
*/
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';

/*
  @App
*/
import { AppComponent } from './app.component';
/*
  @Routes
*/
import { ErrorComponent } from './@routes/error/error.component';
import { BaseComponent } from './@routes/base/base.component';
import { HomeComponent } from './@routes/home/home.component';
import { TaskComponent } from './@routes/task/task.component';
import { NewComponent } from './@routes/new/new.component';
/*
  @Component
*/
import { LoaderComponent } from './@components/loader/loader.component';
import { ErrorMessageComponent } from './@components/error-message/error-message.component';
import { EditorComponent } from './@components/editor/editor.component';
import { StepsViewerComponent } from './@components/stepsviewer/stepsviewer.component';
import { StepsListComponent } from './@components/stepslist/stepslist.component';
import { TemplateDetailsComponent } from './@components/template-details/template-details.component';

import { ModalYamlPreviewComponent } from './@modals/modal-yaml-preview/modal-yaml-preview.component';
import { ModalEditRequestComponent } from './@modals/modal-edit-request/modal-edit-request.component';
import { ModalConfirmationApiComponent } from './@modals/modal-confirmation-api/modal-confirmation-api.component';
import { ModalEditResolutionComponent } from './@modals/modal-edit-resolution/modal-edit-resolution.component';


/*
  @Services Injectable
*/
import { ApiService } from './@services/api.service';
import { ResolutionService } from './@services/resolution.service';
import { TaskService } from './@services/task.service';
import { RequestService } from './@services/request.service';
/*
  @Directives
*/
import { FullHeightDirective } from './@directives/fullheight.directive';
/*
  @Pipes
*/
import { FromNowPipe } from './@pipes/fromNow.pipe';
import { TemplatesComponent } from './@routes/templates/templates.component';
import { TemplateComponent } from './@routes/templates/template.component';
import { MyErrorHandler } from './handlers/error.handler';
import { GraphService } from './@services/graph.service';
import { StatsComponent } from './@routes/stats/stats.component';
import { ChartTaskStatesComponent } from './@components/chart-task-states/chart-task-states.component';
import { AutofocusDirective } from './@directives/autofocus.directive';

const pages = [
  AppComponent,

  BaseComponent,
  ErrorComponent,

  HomeComponent,
  TemplatesComponent,
  TemplateComponent,
  TaskComponent,
  NewComponent,
  StatsComponent,

  LoaderComponent,
  ErrorMessageComponent,
  EditorComponent,
  StepsViewerComponent,
  StepsListComponent,
  TemplateDetailsComponent,
  ChartTaskStatesComponent,

  ModalYamlPreviewComponent,
  ModalEditRequestComponent,
  ModalConfirmationApiComponent,
  ModalEditResolutionComponent,

  FullHeightDirective,
  AutofocusDirective,

  FromNowPipe
];

@NgModule({
  declarations: pages,
  imports: [
    InfiniteScrollModule,
    TagInputModule,
    NgbModule,
    BrowserAnimationsModule,
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ToastrModule.forRoot({
      positionClass: 'toast-bottom-right',
    }),
  ],
  providers: [GraphService, ApiService, ResolutionService, TaskService, RequestService, { provide: ErrorHandler, useClass: MyErrorHandler }],
  bootstrap: [AppComponent],
  entryComponents: [
    ModalYamlPreviewComponent,
    ModalEditRequestComponent,
    ModalConfirmationApiComponent,
    ModalEditResolutionComponent,
  ]
})
export class AppModule { }
