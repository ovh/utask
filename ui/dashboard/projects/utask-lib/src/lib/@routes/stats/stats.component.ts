import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { VMTaskState } from '../../@components/chart-task-states/chart-task-states.component';
import { UTaskLibOptions } from '../../@services/api.service';

@Component({
  templateUrl: './stats.html',
  styleUrls: ['./stats.sass'],
})
export class StatsComponent implements OnInit {
  taskStates: VMTaskState[];
  uiBaseUrl: string;

  constructor(
    private _route: ActivatedRoute,
    private _options: UTaskLibOptions
  ) {
    this.uiBaseUrl = this._options.uiBaseUrl;
  }

  ngOnInit() {
    this.taskStates = [
      {
        color: '#dc3545',
        key: 'BLOCKED',
        label: 'blocked',
        value: this._route.snapshot.data.stats.task_states['BLOCKED'] || 0
      },
      {
        color: '#fd7e14',
        key: 'CANCELLED',
        label: 'cancelled',
        value: this._route.snapshot.data.stats.task_states['CANCELLED'] || 0
      },
      {
        color: '#28a745',
        key: 'DONE',
        label: 'done',
        value: this._route.snapshot.data.stats.task_states['DONE'] || 0
      },
      {
        color: '#007bff',
        key: 'RUNNING',
        label: 'running',
        value: this._route.snapshot.data.stats.task_states['RUNNING'] || 0
      },
      {
        color: '#005ff0',
        key: 'TODO',
        label: 'todo',
        value: this._route.snapshot.data.stats.task_states['TODO'] || 0
      },
      {
        color: '#f06e20',
        key: 'WONTFIX',
        label: 'wontfix',
        value: this._route.snapshot.data.stats.task_states['WONTFIX'] || 0
      },
      {
        color: '#8a2be2',
        key: 'WAITING',
        label: 'waiting',
        value: this._route.snapshot.data.stats.task_states['WAITING'] || 0
      },
    ].sort((a, b) => a.value < b.value ? 1 : -1);
  }
}