/*import { Observable } from 'rxjs';
import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  template: `
    <div>
        Hello WORLD !
        <br/> 
        {{test | json}}
    </div>
  `,
})
export default class HomeComponent implements OnInit {
    test: any;

  constructor(private route: ActivatedRoute
  ) {
      console.log("?,dsq,qdslmkdq", this);
  }

  ngOnInit() {
      this.test = this.route.snapshot.data;
      console.log("NG ON INIT ??", this.route.snapshot.data);
  }
}*/