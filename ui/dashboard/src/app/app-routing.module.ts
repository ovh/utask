import { NgModule } from '@angular/core';
import { Routes, RouterModule, Router } from '@angular/router';
import { MatButtonToggleModule } from '@angular/material/button-toggle';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatButtonModule } from '@angular/material/button';
import { MetaResolve } from './@resolves/meta.resolve';
import { HomeComponent } from './@routes/home/home.component';
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

const routes: Routes = [
  {
    path: 'error', component: ErrorComponent
  },
  {
    path: '', redirectTo: '/home', pathMatch: 'full'
  },
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
        path: 'home',
        component: HomeComponent,
        data: {
          title: {
            value: '{0}',
            args: ['meta.application_name']
          }
        }
      },
      {
        path: 'templates',
        component: TemplatesComponent,
        data: {
          title: {
            value: 'Templates - {0}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'template/:templateName',
        component: TemplateComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: {
          title: {
            value: 'Template - {1}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'functions',
        component: FunctionsComponent,
        data: {
          title: {
            value: 'Functions - {0}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'function/:functionName',
        component: FunctionComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: {
          title: {
            value: 'Function - {1}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'task/:id',
        component: TaskComponent,
        runGuardsAndResolvers: 'paramsOrQueryParamsChange',
        data: {
          title: {
            value: 'Task - {0}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'new',
        component: NewComponent,
        data: {
          title: {
            value: 'New task - {0}',
            args: ['meta.application_name']
          }
        },
      },
      {
        path: 'stats',
        component: StatsComponent,
        data: {
          title: {
            value: 'Stats - {0}',
            args: ['meta.application_name']
          }
        },
        resolve: {
          stats: StatsResolve
        }
      },
    ]
  }
];

@NgModule({
  declarations: [
    HomeComponent,
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
    InfiniteScrollModule,
    TagInputModule,
    NgbModule,
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    MatButtonToggleModule,
    MatCheckboxModule,
    MatButtonModule,
    ToastrModule.forRoot({
      positionClass: 'toast-bottom-right',
    }),
    UTaskModule,
    UTaskLibModule.forRoot({
      apiBaseUrl: environment.apiBaseUrl
    }),
    RouterModule.forRoot(
      routes,
      {
        useHash: true,
        paramsInheritanceStrategy: 'always',
      }
    )
  ],
  exports: [RouterModule],
  providers: [MetaResolve, TemplatesResolve, StatsResolve, FunctionsResolve]
})
export class AppRoutingModule { }
