import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import * as _ from 'lodash';

class VMTaskState {
  label: string;
  color: string;
  value: number;
  key: string;
}

@Component({
  templateUrl: './stats.html',
})
export class StatsComponent implements OnInit {
  taskStates: VMTaskState[];

  constructor(private route: ActivatedRoute) {
  }

  ngOnInit() {
    this.taskStates = _.orderBy([
      {
        color: '#dc3545',
        key: 'BLOCKED',
        label: 'blocked',
        value: this.route.snapshot.data.stats.task_states['BLOCKED'] || 0
      },
      {
        color: '#fd7e14',
        key: 'CANCELLED',
        label: 'cancelled',
        value: this.route.snapshot.data.stats.task_states['CANCELLED'] || 0
      },
      {
        color: '#28a745',
        key: 'DONE',
        label: 'done',
        value: this.route.snapshot.data.stats.task_states['DONE'] || 0
      },
      {
        color: '#007bff',
        key: 'RUNNING',
        label: 'running',
        value: this.route.snapshot.data.stats.task_states['RUNNING'] || 0
      },
      {
        color: '#005ff0',
        key: 'TODO',
        label: 'todo',
        value: this.route.snapshot.data.stats.task_states['TODO'] || 0
      },
      {
        color: '#f06e20',
        key: 'WONTFIX',
        label: 'wontfix',
        value: this.route.snapshot.data.stats.task_states['WONTFIX'] || 0
      },
    ], ['value'], ['desc']);
  }
}