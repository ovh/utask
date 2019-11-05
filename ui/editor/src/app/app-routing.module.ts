import { NgModule, Injectable } from '@angular/core';
// import { Observable } from 'rxjs';
import { Routes, RouterModule/*, ActivatedRouteSnapshot, RouterStateSnapshot, Resolve */} from '@angular/router';

import {EditorComponent} from './editor/editor.component';

import { HomeObject, HomeResolve } from './resolve/home.resolve';

const routes: Routes = [
  { path: 'editor', component: EditorComponent },
  { path: '', redirectTo: '/editor', pathMatch: 'full' }
];

@NgModule({
  imports: [
    RouterModule.forRoot(
    routes,
    { /*enableTracing: true, */useHash: true }
    )
  ],
  exports: [RouterModule],
  providers: [
    HomeResolve,
    HomeObject,
  ]
})
export class AppRoutingModule { }
