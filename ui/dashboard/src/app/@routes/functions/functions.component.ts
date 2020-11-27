import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from 'projects/utask-lib/src/lib/@services/api.service';

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
