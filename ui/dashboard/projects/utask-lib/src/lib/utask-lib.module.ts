import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { NzModalContentWithErrorComponent } from './@modals/modal-content-with-error/modal-content-with-error.component';
import { EditorComponent } from './@components/editor/editor.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { ErrorMessageComponent } from './@components/error-message/error-message.component';
import { InputTagsComponent } from './@components/input-tags/input-tags.component';
import { TasksListComponent } from './@components/tasks-list/tasks-list.component';
import { TaskComponent } from './@routes/task/task.component';
import { TasksComponent } from './@routes/tasks/tasks.component';
import { ResolutionService } from './@services/resolution.service';
import { TaskService } from './@services/task.service';
import { FromNowPipe } from './@pipes/fromNow.pipe';
import { LoaderComponent } from './@components/loader/loader.component';
import { RequestService } from './@services/request.service';
import { WorkflowService } from './@services/workflow.service';
import { ModalService } from './@services/modal.service';
import { StepsListComponent } from './@components/stepslist/stepslist.component';
import { TemplateDetailsComponent } from './@components/template-details/template-details.component';
import { ChartTaskStatesComponent } from './@components/chart-task-states/chart-task-states.component';
import { ModalApiYamlEditComponent } from './@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ModalApiYamlComponent } from './@modals/modal-api-yaml/modal-api-yaml.component';
import { StepsViewerComponent } from './@components/steps-viewer/steps-viewer.component';
import { StepNodeComponent } from './@components/step-node/step-node.component';
import { TaskStatusComponent } from './@components/task-status/task-status.component';
import { InputsFormComponent } from './@components/inputs-form/inputs-form.component';
import { NsAutoHeightTableDirective } from './@directives/ns-auto-height-table.directive';
import { FullHeightDirective } from './@directives/fullheight.directive';
import { AutofocusDirective } from './@directives/autofocus.directive';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzAutocompleteModule } from 'ng-zorro-antd/auto-complete';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzModalModule } from 'ng-zorro-antd/modal';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { BoxComponent } from './@components/box/box.component';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzCodeEditorModule } from 'ng-zorro-antd/code-editor';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { ChartCommonModule, PieChartModule } from '@swimlane/ngx-charts';
import { CommonModule } from '@angular/common';
import { NzCommentModule } from 'ng-zorro-antd/comment';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';
import { NzListModule } from 'ng-zorro-antd/list';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzPageHeaderModule } from 'ng-zorro-antd/page-header';
import { utaskLibRouting } from './utask-lib.routing';
import { FunctionsResolve } from './@resolves/functions.resolve';
import { MetaResolve } from './@resolves/meta.resolve';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { TemplatesComponent } from './@routes/templates/templates.component';
import { TemplateComponent } from './@routes/template/template.component';
import { FunctionsComponent } from './@routes/functions/functions.component';
import { FunctionComponent } from './@routes/function/function.component';
import { NewComponent } from './@routes/new/new.component';
import { StatsComponent } from './@routes/stats/stats.component';
import { ErrorComponent } from './@routes/error/error.component';
import { StatsResolve } from './@resolves/stats.resolve';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzNotificationModule } from 'ng-zorro-antd/notification';
import { UTaskLibOptions, UtaskLibOptionsRefresh } from './@services/api.service';

const components: any[] = [
  // Components
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
  BoxComponent,
  InputsFormComponent,
  ModalApiYamlComponent,
  ModalApiYamlEditComponent,
  NzModalContentWithErrorComponent,
  TaskStatusComponent,

  // Routes
  TasksComponent,
  TaskComponent,
  NewComponent,
  TemplatesComponent,
  TemplateComponent,
  FunctionsComponent,
  FunctionComponent,
  StatsComponent,
  ErrorComponent,

  FromNowPipe,

  NsAutoHeightTableDirective,
  FullHeightDirective,
  AutofocusDirective
];

@NgModule({
  declarations: components,
  imports: [
    CommonModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    RouterModule,
    utaskLibRouting,

    ChartCommonModule,
    PieChartModule,

    NzTableModule,
    NzButtonModule,
    NzIconModule,
    NzDividerModule,
    NzDropDownModule,
    NzInputModule,
    NzSelectModule,
    NzAutocompleteModule,
    NzModalModule,
    NzGridModule,
    NzAlertModule,
    NzCollapseModule,
    NzCodeEditorModule,
    NzFormModule,
    NzCheckboxModule,
    NzSpinModule,
    NzDescriptionsModule,
    NzToolTipModule,
    NzCommentModule,
    NzAvatarModule,
    NzListModule,
    NzSwitchModule,
    NzPageHeaderModule,
    NzResultModule,
    NzNotificationModule
  ],
  exports: components,
  bootstrap: [],
  entryComponents: [
    ModalApiYamlComponent,
    ModalApiYamlEditComponent,
    NzModalContentWithErrorComponent
  ],
  providers: [
    ModalService,
    ResolutionService,
    TaskService,
    RequestService,
    WorkflowService,
    MetaResolve,
    TemplatesResolve,
    FunctionsResolve,
    StatsResolve
  ]
})
export class UTaskLibModule { }

export function UTaskLibOptionsFactory(apiBaseUrl: string, uiBaseUrl: string, refresh: UtaskLibOptionsRefresh): any {
  const res = () => new UTaskLibOptions(apiBaseUrl, uiBaseUrl, refresh);
  return res;
}