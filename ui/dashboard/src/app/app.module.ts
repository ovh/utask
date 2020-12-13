import { NgModule, ErrorHandler } from '@angular/core';
import {
  faUserShield,
  faCheckCircle,
  faTimesCircle,
  faBan,
  faHistory,
  faSync,
  faHourglassHalf,
  faQuestionCircle,
  faCaretDown,
  faCaretUp,
  faAlignJustify
} from '@fortawesome/fontawesome-free-solid';
import fontawesome from '@fortawesome/fontawesome';
fontawesome.library.add(
  faUserShield,
  faCheckCircle,
  faTimesCircle,
  faBan,
  faHistory,
  faSync,
  faHourglassHalf,
  faQuestionCircle,
  faCaretDown,
  faCaretUp,
  faAlignJustify
);
import { AppComponent } from './app.component';
import { MyErrorHandler } from './handlers/error.handler';
import { ThemeService } from './@services/theme.service';
import { routing } from './app.routing';
import { NzSwitchModule } from 'ng-zorro-antd/switch';
import { NzBadgeModule } from 'ng-zorro-antd/badge';
import { NzLayoutModule } from 'ng-zorro-antd/layout';
import { BrowserModule } from '@angular/platform-browser';
import { BaseComponent } from './@routes/base/base.component';
import { HttpClientModule } from '@angular/common/http';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NzAvatarModule } from 'ng-zorro-antd/avatar';
import { NzButtonModule } from 'ng-zorro-antd/button';
import { NzMenuModule } from 'ng-zorro-antd/menu';
import { NzToolTipModule } from 'ng-zorro-antd/tooltip';
import { ApiServiceOptions } from 'projects/utask-lib/src/lib/@services/api.service';
import { ApiServiceOptionsFactory } from 'projects/utask-lib/src/lib/utask-lib.module';
import { environment } from 'src/environments/environment';
import { MetaResolve } from 'projects/utask-lib/src/lib/@resolves/meta.resolve';
import { en_US, NZ_I18N } from 'ng-zorro-antd/i18n';
import { IconsProviderModule } from './icons-provider.module';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

@NgModule({
  declarations: [
    AppComponent,
    BaseComponent,
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,

    NzLayoutModule,
    NzMenuModule,
    NzButtonModule,
    NzAvatarModule,
    NzBadgeModule,
    NzSwitchModule,
    NzToolTipModule,

    IconsProviderModule,

    routing,
  ],
  providers: [
    { provide: NZ_I18N, useValue: en_US },
    { provide: ErrorHandler, useClass: MyErrorHandler },
    {
      provide: ApiServiceOptions,
      useFactory: ApiServiceOptionsFactory(environment.apiBaseUrl),
    },
    ThemeService,
    MetaResolve
  ],
  bootstrap: [AppComponent],
  entryComponents: []
})
export class AppModule { }
