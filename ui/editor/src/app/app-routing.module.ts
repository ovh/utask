import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';

import { EditorComponent } from './editor/editor.component';

import { HomeObject, HomeResolve } from './resolve/home.resolve';

const routes: Routes = [
  { path: 'editor', component: EditorComponent },
  { path: '', redirectTo: '/editor', pathMatch: 'full' }
];

@NgModule({
  imports: [
    RouterModule.forRoot(
      routes,
      { useHash: true }
    )
  ],
  exports: [RouterModule],
  providers: [
    HomeResolve,
    HomeObject,
  ]
})
export class AppRoutingModule { }
