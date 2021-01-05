import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { NzModalContentWithErrorComponent } from './@modals/modal-content-with-error/modal-content-with-error.component';
import { EditorComponent } from './@components/editor/editor.component';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { ErrorMessageComponent } from './@components/error-message/error-message.component';
import { InputTagsComponent } from './@components/input-tags/input-tags.component';
import { TasksListComponent } from './@components/tasks-list/tasks-list.component';
import { FromNowPipe } from './@pipes/fromNow.pipe';
import { LoaderComponent } from './@components/loader/loader.component';
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
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzNotificationModule } from 'ng-zorro-antd/notification';
import { UTaskLibOptions, UtaskLibOptionsRefresh } from './@services/api.service';
import { en_US, NZ_I18N } from 'ng-zorro-antd/i18n';
import { ModalService } from './@services/modal.service';
import { ResolutionService } from './@services/resolution.service';
import { TaskService } from './@services/task.service';
import { FunctionsResolve } from './@resolves/functions.resolve';
import { MetaResolve } from './@resolves/meta.resolve';
import { StatsResolve } from './@resolves/stats.resolve';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { RequestService } from './@services/request.service';
import { WorkflowService } from './@services/workflow.service';

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
  BoxComponent,
  InputsFormComponent,
  ModalApiYamlComponent,
  ModalApiYamlEditComponent,
  NzModalContentWithErrorComponent,
  TaskStatusComponent,

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