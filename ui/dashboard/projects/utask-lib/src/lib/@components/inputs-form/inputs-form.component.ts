import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { FormGroup } from '@angular/forms';
import { ResolverInput } from '../../@models/task.model';
@Component({
	selector: 'lib-utask-inputs-form',
	templateUrl: './inputs-form.html',
	styleUrls: ['./inputs-form.scss'],
	changeDetection: ChangeDetectionStrategy.OnPush
})
export class InputsFormComponent {
	@Input() inputs: Array<ResolverInput>;
	@Input() formGroup: FormGroup;

	public static getInputs(values: any): any {
		const inputs = {};
		Object.keys(values)
			.filter(k => k.startsWith('input_'))
			.forEach(k => inputs[k.split('input_')[1]] = values[k]);
		return inputs;
	}

	trackInput(index: number, input: ResolverInput): string {
		return input.name;
	}
}