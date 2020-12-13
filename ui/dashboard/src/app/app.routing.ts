import { ModuleWithProviders } from '@angular/core';
import { Routes, RouterModule, PreloadAllModules } from '@angular/router';
import { BaseComponent } from './@routes/base/base.component';
import { AppModule } from './app.module';
import { MetaResolve } from 'projects/utask-lib/src/lib/@resolves/meta.resolve';

const routes: Routes = [
  {
    path: '',
    component: BaseComponent,
    resolve: {
      meta: MetaResolve
    },
    children: [
      {
        path: '', loadChildren: () => import('../../projects/utask-lib/src/lib/utask-lib.module')
          .then(m => m.UTaskLibModule)
      }
    ]
  }
];

export const routing: ModuleWithProviders<AppModule> = RouterModule.forRoot(routes, {
  initialNavigation: 'enabledNonBlocking',
  preloadingStrategy: PreloadAllModules,
  relativeLinkResolution: 'legacy',
  useHash: true,
  paramsInheritanceStrategy: 'always'
});
