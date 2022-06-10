import { AfterViewInit, Component, HostListener, ViewChild } from '@angular/core';
import { StepsViewerComponent } from '@ovhcloud/utask-lib';
import Step from '@ovhcloud/utask-lib/lib/@models/step.model';
import { parse } from 'yaml';

@Component({
  selector: 'app-preview',
  templateUrl: './preview.component.html',
  styleUrls: ['./preview.component.scss']
})
export class PreviewComponent implements AfterViewInit {
  // @ts-ignore
  private _vscode = acquireVsCodeApi();

  @ViewChild(StepsViewerComponent, { static: true }) viewer!: StepsViewerComponent;

  ngAfterViewInit(): void {
    this._vscode.postMessage({
      type: 'initialized',
      value: true,
    })
  }

  @HostListener('window:message', ['$event'])
  onRefresh(e: MessageEvent) {
    if (e.data.type === 'refresh') {
      try {
        this.viewer.resolution = parse(e.data.value);
        this.viewer.clear();
        this.viewer.draw();
      } catch (e) {
        console.error(e);
      }
    }
  }
}
