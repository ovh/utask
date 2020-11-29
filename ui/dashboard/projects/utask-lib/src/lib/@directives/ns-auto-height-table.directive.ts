// Forked code from https://github.com/1-2-3/zorro-sharper

import {
  Directive,
  ElementRef,
  Input,
  SimpleChange,
  HostListener,
  ChangeDetectorRef,
} from '@angular/core';
import { NzTableComponent } from 'ng-zorro-antd/table';
import { off } from 'process';

@Directive({
  selector: '[nsAutoHeightTable]'
})
export class NsAutoHeightTableDirective {
  @Input('nsAutoHeightTable')
  offset: number;

  constructor(
    private element: ElementRef,
    private table: NzTableComponent,
    private cd: ChangeDetectorRef
  ) {
    if (this.table && this.table.nzPageIndexChange) {
      this.table.nzPageIndexChange.subscribe((index) => {
        let tableBody = this.element.nativeElement.querySelector(
          '.ant-table-body'
        );
        if (tableBody && tableBody.scrollTop) {
          tableBody.scrollTop = 0;
        }
      });
    }
  }

  @HostListener('window:resize', ['$event'])
  onResize() {
    this.doAutoSize();
  }

  ngOnInit() { }

  ngAfterViewInit() {
    this.doAutoSize();
  }

  private doAutoSize() {
    setTimeout(() => {
      let offset = this.offset || 0;
      if (
        this.element &&
        this.element.nativeElement &&
        this.element.nativeElement.parentElement &&
        this.element.nativeElement.parentElement.offsetHeight
      ) {
        if (this.table && this.table.nzScroll && this.table.nzScroll.x) {
          let originNzScroll = this.table.nzScroll
            ? { ...this.table.nzScroll }
            : null;
          this.table.nzScroll = {
            y: (this.element.nativeElement.parentElement.offsetHeight - offset).toString() + 'px',
            x: this.table.nzScroll.x,
          };
          this.table.ngOnChanges({
            nzScroll: new SimpleChange(
              { originNzScroll },
              this.table.nzScroll,
              false
            ),
          });
          this.cd.detectChanges();
        } else {
          let originNzScroll = this.table.nzScroll
            ? { ...this.table.nzScroll }
            : null;
          this.table.nzScroll = {
            ...{
              y: (this.element.nativeElement.parentElement.offsetHeight - offset).toString() + 'px'
            },
          };

          this.table.ngOnChanges({
            nzScroll: new SimpleChange(
              { originNzScroll },
              this.table.nzScroll,
              false
            ),
          });
          this.cd.detectChanges();
        }
      }
    }, 10);
  }
}