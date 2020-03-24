import {NgbModule} from '@ng-bootstrap/ng-bootstrap';

import { BrowserModule } from '@angular/platform-browser';
import { NgModule/*, enableProdMode*/ } from '@angular/core';
// enableProdMode();

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';

import {EditorComponent} from './editor/editor.component';
import {StepsViewerComponent} from './components/stepsviewer.component';

import {FullHeightDirective} from './directives/fullheight.directive';

import {TemplateYamlHelper} from './services/templateyamlhelper.service';
import {WorkflowHelper} from './services/workflowhelper.service';

@NgModule({
  declarations: [
    AppComponent,
    EditorComponent,
    StepsViewerComponent,
    FullHeightDirective
  ],
  imports: [
    BrowserModule,
    NgbModule,
    AppRoutingModule,
  ],
  providers: [
    TemplateYamlHelper,
    WorkflowHelper,
  ],
  entryComponents: [
    EditorComponent,
    StepsViewerComponent,
  ],
  bootstrap: [AppComponent]
})
export class AppModule {}
