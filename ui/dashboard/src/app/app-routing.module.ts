import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatButtonModule } from '@angular/material/button';
import { MetaResolve } from './@resolves/meta.resolve';
import { TaskComponent } from './@routes/task/task.component';
import { BaseComponent } from './@routes/base/base.component';
import { ErrorComponent } from './@routes/error/error.component';
import { NewComponent } from './@routes/new/new.component';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { TemplatesComponent } from './@routes/templates/templates.component';
import { TemplateComponent } from './@routes/templates/template.component';
import { StatsComponent } from './@routes/stats/stats.component';
import { StatsResolve } from './@routes/stats/stats.resolve';
import { InfiniteScrollModule } from 'ngx-infinite-scroll';
import { TagInputModule } from 'ngx-chips';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { BrowserModule } from '@angular/platform-browser';
import { FormsModule } from '@angular/forms';
import { ToastrModule } from 'ngx-toastr';
import { UTaskModule } from 'src/utask-module/utask.module';
import { environment } from 'src/environments/environment';
import { FunctionsComponent } from './@routes/functions/functions.component';
import { FunctionComponent } from './@routes/functions/function.component';
import { FunctionsResolve } from './@resolves/functions.resolve';
import { UTaskLibModule } from 'projects/utask-lib/src/lib/utask-lib.module';
import { TasksComponent } from './@routes/tasks/tasks.component';
import { NzSelectModule } from 'ng-zorro-antd/select';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { IconsProviderModule } from './icons-provider.module';
import { NzBreadCrumbModule } from 'ng-zorro-antd/breadcrumb';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';
import { NzBadgeModule } from 'ng-zorro-antd/badge';

const routes: Routes = [
  { path: 'error', component: ErrorComponent },
  { path: '', redirectTo: '/tasks', pathMatch: 'full' },
  {
    path: '',
    component: BaseComponent,
    resolve: {
      meta: MetaResolve,
      templates: TemplatesResolve,
      functions: FunctionsResolve,
    },
    children: [
      {
        path: 'tasks',
        component: TasksComponent,
        data: { title: 'Tasks' }
      },
      {
        path: 'templates',
        component: TemplatesComponent,
        data: { title: 'Templates' },
      },
      {
        path: 'template/:templateName',
        component: TemplateComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: { title: 'Template' }
      },
      {
        path: 'functions',
        component: FunctionsComponent,
        data: { title: 'Functions' },
      },
      {
        path: 'function/:functionName',
        component: FunctionComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: { title: 'Function' },
      },
      {
        path: 'task/:id',
        component: TaskComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: { title: 'Task' }
      },
      {
        path: 'new',
        component: NewComponent,
        data: { title: 'New task' },
      },
      {
        path: 'stats',
        component: StatsComponent,
        data: { title: 'Stats' },
        resolve: {
          stats: StatsResolve
        }
      },
    ]
  }
];

@NgModule({
  declarations: [
    TasksComponent,
    BaseComponent,
    ErrorComponent,
    TemplatesComponent,
    TemplateComponent,
    FunctionComponent,
    FunctionsComponent,
    TaskComponent,
    NewComponent,
    StatsComponent,
  ],
  imports: [
    IconsProviderModule,
    NzLayoutModule,
    NzMenuModule,
    NzButtonModule,
    NzBreadCrumbModule,
    NzSelectModule,
    NzAvatarModule,
    NzBadgeModule,
    InfiniteScrollModule,
    TagInputModule,
    NgbModule,
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    MatButtonToggleModule,
    MatCheckboxModule,
    MatButtonModule,
    ToastrModule.forRoot({ positionClass: 'toast-bottom-right', }),
    UTaskModule,
    UTaskLibModule.forRoot({ apiBaseUrl: environment.apiBaseUrl }),
    RouterModule.forRoot(
      routes,
      {
        useHash: true,
        paramsInheritanceStrategy: 'always',
      }
    )
  ],
  exports: [RouterModule],
  providers: [
    MetaResolve,
    TemplatesResolve,
    StatsResolve,
    FunctionsResolve
  ]
})
export class AppRoutingModule { }
