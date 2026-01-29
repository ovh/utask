import { ChangeDetectionStrategy, ChangeDetectorRef, Component, Input, OnChanges } from '@angular/core';
import Task from '../../@models/task.model';

@Component({
    selector: 'lib-utask-task-status',
    templateUrl: './task-status.html',
    styleUrls: ['./task-status.sass'],
    changeDetection: ChangeDetectionStrategy.OnPush,
    standalone: false
})
export class TaskStatusComponent implements OnChanges {
	@Input() task: Task;
	percentage: number;

	constructor(
		private _cd: ChangeDetectorRef
	) { }

	ngOnChanges(): void {
		if (!this.task) { return; }
		this.percentage = Math.round(this.task.steps_done / this.task.steps_total * 100);
		this._cd.markForCheck();
	}
}