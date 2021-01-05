import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { NzTableComponent } from 'ng-zorro-antd/table';

@Component({
  templateUrl: './functions.html',
  styleUrls: ['./functions.sass'],
})
export class FunctionsComponent implements OnInit {
  functions: Function[];
  @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Function>;
  display: { [key: string]: boolean } = {};

  expandSet = new Set<number>();
  codes: string[] = [];

  onExpandChange(id: number, checked: boolean): void {
    if (checked) {
      this.expandSet.add(id);
    } else {
      this.expandSet.delete(id);
    }
  }

  constructor(
    private _route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.functions = this._route.parent.snapshot.data.functions;
    this.codes = [];
    this.functions.forEach((item) => {
      this.codes.push(JSON.stringify(item, null, 4));
    });
  }
}
