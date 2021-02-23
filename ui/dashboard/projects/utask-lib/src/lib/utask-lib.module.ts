import { AutofocusDirective } from './@directives/autofocus.directive';
import { BoxComponent } from './@components/box/box.component';
import { ChartCommonModule, PieChartModule } from '@swimlane/ngx-charts';
import { ChartTaskStatesComponent } from './@components/chart-task-states/chart-task-states.component';
import { CommonModule } from '@angular/common';
import { EditorComponent } from './@components/editor/editor.component';
import { en_US, NZ_I18N } from 'ng-zorro-antd/i18n';
import { ErrorMessageComponent } from './@components/error-message/error-message.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { FromNowPipe } from './@pipes/fromNow.pipe';
import { FullHeightDirective } from './@directives/fullheight.directive';
import { FunctionsResolve } from './@resolves/functions.resolve';
import { HttpClientModule } from '@angular/common/http';
import { InputsFormComponent } from './@components/inputs-form/inputs-form.component';
import { InputTagsComponent } from './@components/input-tags/input-tags.component';
import { InputEditorComponent } from './@components/input-editor/input-editor.component';
import { LoaderComponent } from './@components/loader/loader.component';
import { MetaResolve } from './@resolves/meta.resolve';
import { ModalApiYamlComponent } from './@modals/modal-api-yaml/modal-api-yaml.component';
import { ModalApiYamlEditComponent } from './@modals/modal-api-yaml-edit/modal-api-yaml-edit.component';
import { ModalService } from './@services/modal.service';
import { NgModule } from '@angular/core';
import { NsAutoHeightTableDirective } from './@directives/ns-auto-height-table.directive';
import { NzAlertModule } from 'ng-zorro-antd/alert';
import { NzAutocompleteModule } from 'ng-zorro-antd/auto-complete';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzCheckboxModule } from 'ng-zorro-antd/checkbox';
import { NzCodeEditorModule } from 'ng-zorro-antd/code-editor';
import { NzCollapseModule } from 'ng-zorro-antd/collapse';
import { NzCommentModule } from 'ng-zorro-antd/comment';
import { NzDescriptionsModule } from 'ng-zorro-antd/descriptions';
import { NzDividerModule } from 'ng-zorro-antd/divider';
import { NzDropDownModule } from 'ng-zorro-antd/dropdown';
import { NzFormModule } from 'ng-zorro-antd/form';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzListModule } from 'ng-zorro-antd/list';
import { NzModalContentWithErrorComponent } from './@modals/modal-content-with-error/modal-content-with-error.component';
import { NzModalModule } from 'ng-zorro-antd/modal';
import { NzNotificationModule } from 'ng-zorro-antd/notification';
import { NzPageHeaderModule } from 'ng-zorro-antd/page-header';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { RequestService } from './@services/request.service';
import { ResolutionService } from './@services/resolution.service';
import { RouterModule } from '@angular/router';
import { StatsResolve } from './@resolves/stats.resolve';
import { StepNodeComponent } from './@components/step-node/step-node.component';
import { StepsListComponent } from './@components/stepslist/stepslist.component';
import { StepsViewerComponent } from './@components/steps-viewer/steps-viewer.component';
import { TaskService } from './@services/task.service';
import { TasksListComponent } from './@components/tasks-list/tasks-list.component';
import { TaskStatusComponent } from './@components/task-status/task-status.component';
import { TemplateDetailsComponent } from './@components/template-details/template-details.component';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { UTaskLibOptions, UtaskLibOptionsRefresh } from './@services/api.service';
import { WorkflowService } from './@services/workflow.service';

const components: any[] = [
  BoxComponent,
  ChartTaskStatesComponent,
  EditorComponent,
  ErrorMessageComponent,
  InputEditorComponent,
  InputsFormComponent,
  InputTagsComponent,
  LoaderComponent,
  ModalApiYamlComponent,
  ModalApiYamlEditComponent,
  NzModalContentWithErrorComponent,
  StepNodeComponent,
  StepsListComponent,
  StepsViewerComponent,
  TasksListComponent,
  TaskStatusComponent,
  TemplateDetailsComponent,

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
    { provide: NZ_I18N, useValue: en_US },
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