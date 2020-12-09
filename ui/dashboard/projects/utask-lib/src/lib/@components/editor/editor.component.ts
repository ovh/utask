import { Component, ElementRef, Input, ViewChild, Output, EventEmitter, AfterViewInit } from '@angular/core';
import { EditorOptions } from 'ng-zorro-antd/code-editor';

@Component({
    selector: 'lib-utask-editor',
    template: `
        <nz-code-editor class="editor" [ngModel]="ngModel" [nzEditorOption]="config" #editor></nz-code-editor>
    `,
    styleUrls: ['./editor.sass']
})
export class EditorComponent implements AfterViewInit {
    @ViewChild('editor', { static: false }) editor: ElementRef;
    @Input() config: EditorOptions;
    @Input() ngModel: string;
    @Output() ngModelChange = new EventEmitter<string>();

    constructor(private elRef: ElementRef) { }

    ngAfterViewInit() {
        let height = (this.ngModel.split('\n').length + 1) * 18;
        height = height > 400 ? 400 : height;
        this.elRef.nativeElement.style.height = `${height}px`;
    }
}
