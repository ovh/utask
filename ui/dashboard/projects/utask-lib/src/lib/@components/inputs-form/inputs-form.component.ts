import { ChangeDetectionStrategy, Component, Input } from '@angular/core';
import { FormGroup } from '@angular/forms';

@Component({
	selector: 'lib-utask-inputs-form',
	templateUrl: './inputs-form.html',
	styleUrls: ['./inputs-form.sass'],
	changeDetection: ChangeDetectionStrategy.OnPush
})
export class InputsFormComponent {
	@Input() inputs: any;
	@Input() formGroup: FormGroup;

	public static getInputs(values: any): any {
		const inputs = {};
		Object.keys(values)
			.filter(k => k.startsWith('input_'))
			.forEach(k => inputs[k.split('input_')[1]] = values[k]);
		return inputs;
	}

	submitForm(): void { }

	trackInput(index: number, input: any) {
		return input.name;
	}
}