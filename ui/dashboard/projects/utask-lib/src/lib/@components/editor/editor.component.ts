import { Component, ElementRef, Input, Output, EventEmitter, OnChanges, OnInit } from '@angular/core';
import { EditorOptions } from 'ng-zorro-antd/code-editor';
import { NzConfigService } from 'ng-zorro-antd/core/config';

@Component({
    selector: 'lib-utask-editor',
    templateUrl: './editor.html',
    styleUrls: ['./editor.sass'],
    standalone: false
})
export class EditorComponent implements OnInit, OnChanges {
    @Input() set config(c: EditorOptions) {
        this._config = {
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            scrollbar: { alwaysConsumeMouseWheel: false },
            ...c
        };
    }
    _config: EditorOptions;
    @Input() ngModel: string;
    @Output() ngModelChange = new EventEmitter<string>();

    constructor(
        private _el: ElementRef,
        private _nzConfigService: NzConfigService
    ) { }

    ngOnInit(): void {
        this.computeTheme();
        this._nzConfigService.getConfigChangeEventForComponent('codeEditor').subscribe(_ => {
            this.computeTheme();
        });
    }

    ngOnChanges(): void {
        const height = (this.ngModel.split('\n').length + 1) * 18;
        this._el.nativeElement.style.height = `${height}px`;
    }

    computeTheme(): void {
        const editorConfig = this._nzConfigService.getConfigForComponent('codeEditor');
        const theme = editorConfig?.defaultEditorOption?.theme;
        this._el.nativeElement.style.borderColor = theme === 'vs-dark' ? '#434343' : '#d9d9d9';
    }
}
