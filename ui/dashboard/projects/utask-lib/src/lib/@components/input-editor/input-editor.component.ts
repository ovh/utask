import { Component, forwardRef, ChangeDetectionStrategy, ChangeDetectorRef, ElementRef, Input, OnInit } from "@angular/core";
import { ControlValueAccessor, NG_VALUE_ACCESSOR } from "@angular/forms";
import { EditorOptions } from "ng-zorro-antd/code-editor";
import { NzConfigService } from "ng-zorro-antd/core/config";
import { OnChangeType, OnTouchedType } from "ng-zorro-antd/core/types";

@Component({
	selector: 'lib-utask-input-editor',
	templateUrl: './input-editor.html',
	styleUrls: ['./input-editor.sass'],
	providers: [
		{
			provide: NG_VALUE_ACCESSOR,
			useExisting: forwardRef(() => InputEditorComponent),
			multi: true
		}
	],
	changeDetection: ChangeDetectionStrategy.OnPush
})
export class InputEditorComponent implements ControlValueAccessor, OnInit {
	@Input() set config(c: EditorOptions) {
		this._config = {
			minimap: { enabled: false },
			scrollBeyondLastLine: false,
			scrollbar: { alwaysConsumeMouseWheel: false },
			...c
		};
	}
	_config: EditorOptions;

	onChange: OnChangeType = () => { };
	onTouched: OnTouchedType = () => { };
	code: string = '';

	constructor(
		private _cd: ChangeDetectorRef,
		private _el: ElementRef,
		private _nzConfigService: NzConfigService
	) { }

	ngOnInit(): void {
		this.changeCode('');
		this.computeTheme();
		this._nzConfigService.getConfigChangeEventForComponent('codeEditor').subscribe(_ => {
			this.computeTheme();
		})
	}

	writeValue(code: string): void {
		if (!code) {
			return;
		}
		this.code = code;
		this.computeHeight(this.code);
		this._cd.markForCheck();
	}

	registerOnChange(fn: OnChangeType): void {
		this.onChange = fn;
	}

	registerOnTouched(fn: OnTouchedType): void {
		this.onTouched = fn;
	}

	changeCode(code: string): void {
		this.computeHeight(code);
		this.onChange(code);
	}

	computeHeight(code: string): void {
		const height = (code.split('\n').length + 1) * 18;
		this._el.nativeElement.style.height = `${Math.min(height, 400)}px`;
	}

	computeTheme(): void {
		const editorConfig = this._nzConfigService.getConfigForComponent('codeEditor');
		const theme = editorConfig?.defaultEditorOption?.theme;
		this._el.nativeElement.style.borderColor = theme === 'vs-dark' ? '#434343' : '#d9d9d9';
	}
}