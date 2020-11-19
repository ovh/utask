import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { ToastrModule } from 'ngx-toastr';
import { TemplateDetailsComponent } from './@components/template-details/template-details.component';
import { ChartTaskStatesComponent } from './@components/chart-task-states/chart-task-states.component';
import { AutofocusDirective } from './@directives/autofocus.directive';
import { FullHeightDirective } from './@directives/fullheight.directive';
import { FormsModule } from '@angular/forms';
import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';
import { TagInputModule } from 'ngx-chips';
import { InfiniteScrollModule } from 'ngx-infinite-scroll';
import { environment } from 'src/environments/environment';
import { UTaskLibModule } from 'projects/utask-lib/src/lib/utask-lib.module';
import { BoxComponent } from './@components/box/box.component';

const components: any[] = [
  TemplateDetailsComponent,
  ChartTaskStatesComponent,
  BoxComponent,

  FullHeightDirective,
  AutofocusDirective,
];

@NgModule({
  declarations: components,
  imports: [
    HttpClientModule,
    ToastrModule.forRoot({
      positionClass: 'toast-bottom-right',
    }),
    BrowserAnimationsModule,
    BrowserModule,
    FormsModule,
    InfiniteScrollModule,
    TagInputModule,
    NgbModule,

    UTaskLibModule.forRoot({
      apiBaseUrl: environment.apiBaseUrl,
    }),
  ],
  exports: components,
  providers: [
  ],
  bootstrap: [],
  entryComponents: [
  ]
})
export class UTaskModule { }
