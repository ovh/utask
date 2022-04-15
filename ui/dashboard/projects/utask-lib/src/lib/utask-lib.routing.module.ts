import { CommonModule } from '@angular/common';
import { HttpClientModule } from '@angular/common/http';
import { NgModule } from '@angular/core';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { RouterModule, Routes } from '@angular/router';
import { ChartCommonModule, PieChartModule } from '@swimlane/ngx-charts';
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
import { NzGraphModule } from 'ng-zorro-antd/graph';
import { NzGridModule } from 'ng-zorro-antd/grid';
import { en_US, NZ_I18N } from 'ng-zorro-antd/i18n';
import { NzIconModule } from 'ng-zorro-antd/icon';
import { NzInputModule } from 'ng-zorro-antd/input';
import { NzListModule } from 'ng-zorro-antd/list';
import { NzModalModule } from 'ng-zorro-antd/modal';
import { NzNotificationModule } from 'ng-zorro-antd/notification';
import { NzPageHeaderModule } from 'ng-zorro-antd/page-header';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzSpinModule } from 'ng-zorro-antd/spin';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzTableModule } from 'ng-zorro-antd/table';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { FunctionsResolve } from './@resolves/functions.resolve';
import { MetaResolve } from './@resolves/meta.resolve';
import { StatsResolve } from './@resolves/stats.resolve';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { ErrorComponent } from './@routes/error/error.component';
import { FunctionComponent } from './@routes/function/function.component';
import { FunctionsComponent } from './@routes/functions/functions.component';
import { NewComponent } from './@routes/new/new.component';
import { StatsComponent } from './@routes/stats/stats.component';
import { TaskComponent } from './@routes/task/task.component';
import { TasksComponent } from './@routes/tasks/tasks.component';
import { TemplateComponent } from './@routes/template/template.component';
import { TemplatesComponent } from './@routes/templates/templates.component';
import { UTaskLibModule } from './utask-lib.module';

const utaskLibRoutes: Routes = [
	{
		path: '',
		children: [
			{
				path: '',
				resolve: {
					meta: MetaResolve,
					templates: TemplatesResolve,
					functions: FunctionsResolve,
				},
				children: [
					{ path: 'templates', component: TemplatesComponent, data: { title: 'Templates' } },
					{
						path: 'template/:templateName',
						component: TemplateComponent,
						runGuardsAndResolvers: 'paramsOrQueryParamsChange',
						data: { title: 'Template' }
					},
					{ path: 'functions', component: FunctionsComponent, data: { title: 'Functions' } },
					{
						path: 'function/:functionName',
						component: FunctionComponent,
						runGuardsAndResolvers: 'paramsOrQueryParamsChange',
						data: { title: 'Function' },
					},
					{ path: 'tasks', component: TasksComponent, data: { title: 'Tasks' } },
					{
						path: 'task/:id',
						component: TaskComponent,
						runGuardsAndResolvers: 'paramsOrQueryParamsChange',
						data: { title: 'Task' }
					},
					{ path: 'new', component: NewComponent, data: { title: 'New task' } },
					{
						path: 'stats',
						component: StatsComponent,
						data: { title: 'Stats' },
						resolve: {
							stats: StatsResolve
						}
					}
				]
			}
		]
	}
];

@NgModule({
	declarations: [
		TasksComponent,
		TaskComponent,
		NewComponent,
		TemplatesComponent,
		TemplateComponent,
		FunctionsComponent,
		FunctionComponent,
		StatsComponent,
		ErrorComponent
	],
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
		NzGraphModule,
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
		NzNotificationModule,

		UTaskLibModule,

		RouterModule.forChild(utaskLibRoutes)
	],
	providers: [
		{ provide: NZ_I18N, useValue: en_US }
	]
})
export class UTaskLibRoutingModule { }

