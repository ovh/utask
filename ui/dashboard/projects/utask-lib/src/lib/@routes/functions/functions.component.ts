import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { UTaskLibOptions } from '../../@services/api.service';

@Component({
    templateUrl: './functions.html',
    styleUrls: ['./functions.sass'],
    standalone: false
})
export class FunctionsComponent implements OnInit {
  uiBaseUrl: string;
  functions: Function[];

  constructor(
    private _route: ActivatedRoute,
    private _options: UTaskLibOptions
  ) {
    this.uiBaseUrl = this._options.uiBaseUrl;
  }

  ngOnInit() {
    this.functions = this._route.parent.snapshot.data.functions;
  }
}
