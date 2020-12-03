import { NgModule, ModuleWithProviders } from '@angular/core';
import { ApiServiceOptions } from './@services/api.service';
import { HttpClientModule, HttpClient } from '@angular/common/http';
import { ModalConfirmationApiComponent } from './@modals/modal-confirmation-api/modal-confirmation-api.component';
import { EditorComponent } from './@components/editor/editor.component';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { ErrorMessageComponent } from './@components/error-message/error-message.component';
import { InputTagsComponent } from './@components/input-tags/input-tags.component';
import { TasksListComponent } from './@components/tasks-list/tasks-list.component';
import { ResolutionService } from './@services/resolution.service';
import { TaskService } from './@services/task.service';
import { FromNowPipe } from './@pipes/fromNow.pipe';
import { LoaderComponent } from './@components/loader/loader.component';
import { RequestService } from './@services/request.service';
import { WorkflowService } from './@services/workflow.service';
import { InfiniteScrollModule } from 'ngx-infinite-scroll';
import { StepsListComponent } from './@components/stepslist/stepslist.component';
import { TagInputModule } from 'ngx-chips';
import { TemplateDetailsComponent } from './@components/template-details/template-details.component';
import { ChartTaskStatesComponent } from './@components/chart-task-states/chart-task-states.component';
import { ModalApiYamlEditComponent } from './@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ModalApiYamlComponent } from './@modals/modal-api-yaml/modal-api-yaml.component';
import { StepsViewerComponent } from './@components/steps-viewer/steps-viewer.component';
import { StepNodeComponent } from './@components/step-node/step-node.component';
import { TaskStatusComponent } from './@components/task-status/task-status.component';
import { NsAutoHeightTableDirective } from './@directives/ns-auto-height-table.directive';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NZ_I18N, en_US } from 'ng-zorro-antd/i18n';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzAutocompleteModule } from 'ng-zorro-antd/auto-complete';
import { NzGridModule } from 'ng-zorro-antd/grid';

const components: any[] = [
  LoaderComponent,
  ErrorMessageComponent,
  InputTagsComponent,
  EditorComponent,
  TasksListComponent,
  StepsViewerComponent,
  StepNodeComponent,
  StepsListComponent,
  TemplateDetailsComponent,
  ChartTaskStatesComponent,

  ModalConfirmationApiComponent,
  ModalApiYamlComponent,
  ModalApiYamlEditComponent,

  FromNowPipe,
  TaskStatusComponent,
  NsAutoHeightTableDirective
];

interface UtaskLibConfiguration {
  apiBaseUrl: string;
};

@NgModule({
  declarations: components,
  imports: [
    HttpClientModule,
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    RouterModule,
    TagInputModule,
    NgbModule,
    InfiniteScrollModule,
    NzTableModule,
    NzButtonModule,
    NzIconModule,
    NzDividerModule,
    NzDropDownModule,
    NzInputModule,
    NzSelectModule,
    NzAutocompleteModule,
    NzGridModule,
  ],
  exports: components,
  bootstrap: [],
  entryComponents: [
    ModalConfirmationApiComponent,
    ModalApiYamlComponent,
    ModalApiYamlEditComponent
  ],
})
export class UTaskLibModule {
  static forRoot(conf: UtaskLibConfiguration): ModuleWithProviders<UTaskLibModule> {
    return {
      ngModule: UTaskLibModule,
      providers: [
        {
          provide: ApiServiceOptions,
          useFactory: ApiServiceOptionsFactory(conf.apiBaseUrl),
        },
        ResolutionService,
        TaskService,
        RequestService,
        WorkflowService,
        { provide: NZ_I18N, useValue: en_US }
      ]
    }
  }
}

export function ApiServiceOptionsFactory(apiBaseUrl: string): any {
  const res = (http: HttpClient) => new ApiServiceOptions(apiBaseUrl);
  return res;
}