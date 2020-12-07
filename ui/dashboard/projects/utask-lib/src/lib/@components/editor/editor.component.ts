import { AfterViewInit, Component, ElementRef, Input, ViewChild, Output, EventEmitter, OnChanges } from '@angular/core';
import { EditorOptions } from 'ng-zorro-antd/code-editor';

@Component({
    selector: 'lib-utask-editor',
    template: `
    <nz-code-editor class="editor" [ngModel]="ngModel" [nzEditorOption]="config"></nz-code-editor> 
    `,
})
export class EditorComponent {
    @ViewChild('editor', {
        static: false
    }) editor: ElementRef;
    @Input() config: EditorOptions;
    @Input() ngModel: string;
    @Output() ngModelChange = new EventEmitter<string>();

    constructor() {
    }
}
