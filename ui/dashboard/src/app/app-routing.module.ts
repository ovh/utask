import { NgModule } from '@angular/core';
import { Routes, RouterModule, Router } from '@angular/router';

import { MetaResolve } from './@resolves/meta.resolve';
import { HomeComponent } from './@routes/home/home.component';
import { TaskComponent } from './@routes/task/task.component';
import { BaseComponent } from './@routes/base/base.component';
import { ErrorComponent } from './@routes/error/error.component';
import { NewComponent } from './@routes/new/new.component';
import { TemplatesResolve } from './@resolves/templates.resolve';
import { TemplatesComponent } from './@routes/templates/templates.component';
import { TemplateComponent } from './@routes/templates/template.component';


// const routes: Routes = [
//   { path: 'home', component: HomeComponent, resolve: { data: HomeResolve } },
//   { path: '', redirectTo: '/home', pathMatch: 'full' }
// ];

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
      templates: TemplatesResolve
    },
    children: [
      {
        path: 'home',
        component: HomeComponent,
      },
      {
        path: 'templates',
        component: TemplatesComponent,
      },
      {
        path: 'template/:templateName',
        component: TemplateComponent,
      },
      {
        path: 'task/:id',
        component: TaskComponent,
      },
      {
        path: 'new',
        component: NewComponent,
      }
    ]
  }
];

@NgModule({
  imports: [
    RouterModule.forRoot(
      routes,
      {
        useHash: true,
      }
    )
  ],
  exports: [RouterModule],
  providers: [MetaResolve, TemplatesResolve]
})
export class AppRoutingModule { }
