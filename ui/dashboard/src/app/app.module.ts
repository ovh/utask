import { NgModule, ErrorHandler } from '@angular/core';
import { AppComponent } from './app.component';
import { MyErrorHandler } from './handlers/error.handler';
import { ThemeService } from './@services/theme.service';
import { routing } from './app.routing';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzBadgeModule } from 'ng-zorro-antd/badge';
import { NzElementPatchModule } from 'ng-zorro-antd/core/element-patch';

import { NzGraphModule } from 'ng-zorro-antd/graph';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { BrowserModule } from '@angular/platform-browser';
import { BaseComponent } from './@routes/base/base.component';
import { HttpClientModule } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { UTaskLibOptions } from 'projects/utask-lib/src/lib/@services/api.service';
import { UTaskLibOptionsFactory } from 'projects/utask-lib/src/lib/utask-lib.module';
import { environment } from 'src/environments/environment';
import { MetaResolve } from 'projects/utask-lib/src/lib/@resolves/meta.resolve';
import { en_US, NZ_I18N } from 'ng-zorro-antd/i18n';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { NotFoundComponent } from './@routes/not-found/not-found.component';
import { NzResultModule } from 'ng-zorro-antd/result';
import { NzIconModule } from 'ng-zorro-antd/icon';

@NgModule({
  declarations: [
    AppComponent,
    BaseComponent,
    NotFoundComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,

    NzLayoutModule,
    NzGraphModule,
    NzMenuModule,
    NzButtonModule,
    NzAvatarModule,
    NzBadgeModule,
    NzElementPatchModule,
    NzSwitchModule,
    NzToolTipModule,
    NzResultModule,
    NzIconModule,

    routing
  ],
  providers: [
    { provide: NZ_I18N, useValue: en_US },
    { provide: ErrorHandler, useClass: MyErrorHandler },
    {
      provide: UTaskLibOptions,
      useFactory: UTaskLibOptionsFactory(environment.apiBaseUrl, '/', environment.refresh),
    },
    ThemeService,
    MetaResolve
  ],
  bootstrap: [AppComponent],
})
export class AppModule { }
