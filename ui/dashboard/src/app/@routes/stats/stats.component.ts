import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { VMTaskState } from 'projects/utask-lib/src/lib/@components/chart-task-states/chart-task-states.component';

@Component({
  templateUrl: './stats.html',
  styleUrls: ['./stats.sass'],
})
export class StatsComponent implements OnInit {
  taskStates: VMTaskState[];

  constructor(
    private route: ActivatedRoute
  ) { }

  ngOnInit() {
    this.taskStates = [
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
    ].sort((a, b) => a.value < b.value ? 1 : -1);
  }
}