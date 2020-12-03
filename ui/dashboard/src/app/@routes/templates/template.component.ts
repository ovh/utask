import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';

@Component({
  templateUrl: './template.html',
  styleUrls: ['./template.sass'],
})
export class TemplateComponent implements OnInit {
  templateName: string;

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.route.params.subscribe(params => {
      this.templateName = params.templateName;
    });
  }
}
