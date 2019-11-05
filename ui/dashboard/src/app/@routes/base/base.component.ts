import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';

@Component({
  templateUrl: './base.html',
})
export class BaseComponent implements OnInit {
  meta: any;
  constructor(private route: ActivatedRoute, private router: Router) {
  }

  ngOnInit() {
    this.meta = this.route.snapshot.data.meta;
  }
}
