import { Component, ElementRef, Input, ViewChild, Output, EventEmitter, OnChanges } from '@angular/core';
import { EditorOptions } from 'ng-zorro-antd/code-editor';

@Component({
    selector: 'lib-utask-editor',
    templateUrl: './editor.html',
    styleUrls: ['./editor.sass']
})
export class EditorComponent implements OnChanges {
    @ViewChild('editor', { static: false }) editor: ElementRef;
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

    constructor(private elRef: ElementRef) { }

    ngOnChanges(): void {
        const height = (this.ngModel.split('\n').length + 1) * 18;
        this.elRef.nativeElement.style.height = `${height}px`;
    }
}
