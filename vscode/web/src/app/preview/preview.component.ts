import { AfterViewInit, ChangeDetectorRef, Component, HostListener, ViewChild } from '@angular/core';
import { Resolution, StepsViewerComponent } from '@ovhcloud/utask-lib';

import { parse } from 'yaml';

@Component({
    selector: 'app-preview',
    templateUrl: './preview.component.html',
    styleUrls: ['./preview.component.scss'],
    standalone: false
})
export class PreviewComponent implements AfterViewInit {
  // @ts-ignore
  private _vscode = acquireVsCodeApi();

  resolution: Resolution | null = null;
  invalid = false;

  @ViewChild(StepsViewerComponent, { static: false }) viewer!: StepsViewerComponent;

  constructor(private _changeDetectorRef: ChangeDetectorRef){}

  ngAfterViewInit(): void {
    this._vscode.postMessage({
      type: 'initialized',
      value: true,
    });
  }

  private setResolution(resolution: Resolution | null): void {
    this.resolution = resolution;
    this._changeDetectorRef.detectChanges();
  }

  private clearResolution(): void {
    this.setResolution(null);
  }

  private setInvalid(invalid: boolean): void {
    this.invalid = invalid;
    this._changeDetectorRef.detectChanges();
  }

  @HostListener('window:message', ['$event'])
  onRefresh(e: MessageEvent) {
    if (e.data.type === 'refresh') {
      this.setInvalid(false);
      this.clearResolution();

      try {
        this.setResolution(parse(e.data.value) as Resolution);
        this.viewer.fit();
        this.viewer.draw();
      } catch (e) {
        this.setInvalid(true);
        console.error(e);
      }
    }
  }
}
