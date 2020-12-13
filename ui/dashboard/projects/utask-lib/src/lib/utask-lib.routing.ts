import { ModuleWithProviders } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
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
			{ path: 'error', component: ErrorComponent },
			{ path: '', redirectTo: '/tasks', pathMatch: 'full' },
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

export const utaskLibRouting: ModuleWithProviders<UTaskLibModule> = RouterModule.forChild(utaskLibRoutes);
