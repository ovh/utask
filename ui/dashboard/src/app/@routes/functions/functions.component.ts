import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import * as _ from 'lodash';
import { ApiService } from 'utask-lib';
import Function from 'utask-lib/@models/function.model';

@Component({
  templateUrl: './functions.html',
})
export class FunctionsComponent implements OnInit {
  functions: Function[];
  display: { [key: string]: boolean } = {};
  JSON = JSON;

  constructor(private api: ApiService, private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.functions = this.route.parent.snapshot.data.functions;
  }
}
