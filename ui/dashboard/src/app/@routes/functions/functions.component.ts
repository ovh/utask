import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { NzTableComponent } from 'ng-zorro-antd/table';
import { ApiService } from 'projects/utask-lib/src/lib/@services/api.service';

@Component({
  templateUrl: './functions.html',
  styleUrls: ['./functions.sass'],
})
export class FunctionsComponent implements OnInit {
  functions: Function[];
  @ViewChild('virtualTable') nzTableComponent?: NzTableComponent<Function>;
  display: { [key: string]: boolean } = {};
  JSON = JSON;

  expandSet = new Set<number>();

  onExpandChange(id: number, checked: boolean): void {
    if (checked) {
      this.expandSet.add(id);
    } else {
      this.expandSet.delete(id);
    }
  }

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.functions = this.route.parent.snapshot.data.functions;
  }
}
