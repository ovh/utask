import { AfterViewInit, Component, ElementRef, Input, ViewChild, Output, EventEmitter, OnChanges } from '@angular/core';
import * as brace from 'brace';
import 'brace/mode/json';
import 'brace/mode/yaml';
import 'brace/mode/json';
import 'brace/theme/monokai';
import get from 'lodash-es/get';
import EditorConfig from '../../@models/editorconfig.model';

@Component({
    selector: 'lib-utask-editor',
    template: `
    <div #editor>
    </div>
    `,
})
export class EditorComponent implements AfterViewInit, OnChanges {
    @ViewChild('editor', {
        static: false
    }) editor: ElementRef;
    @Input() value: string;
    @Input() errors: any[];
    @Input() config: EditorConfig;
    @Output() public update: EventEmitter<any> = new EventEmitter();

    private changedDelay: Number;
    private aceEditor: brace.Editor;

    constructor() {
    }

    ngOnChanges(diff) {
        if (diff.errors && this.aceEditor) {
            this.aceEditor.getSession().clearAnnotations();
            this.aceEditor.getSession().setAnnotations(this.errors);
        }
    }

    ngAfterViewInit() {
        this.aceEditor = brace.edit(this.editor.nativeElement);
        this.aceEditor.getSession().setMode(get(this, 'config.mode', 'ace/mode/json'));
        this.aceEditor.setOptions({
            tabSize: get(this, 'config.tabSize', 2),
            maxLines: get(this, 'config.maxLines', Infinity),
        });
        this.aceEditor.setTheme(get(this, 'config.theme', 'ace/theme/monokai'));
        this.aceEditor.setValue(this.value);
        this.aceEditor.setReadOnly(get(this, 'config.readonly', false));
        this.aceEditor.getSession().setUseWrapMode(get(this, 'config.wordwrap', true));
        this.aceEditor.getSession().on('change', () => {
            this.changedDelay = +new Date();
            const lastChanged = this.changedDelay;
            setTimeout(() => {
                if (lastChanged === this.changedDelay) {
                    this.update.emit(this.aceEditor.getValue());
                }
            }, 250);
        });
        this.aceEditor.clearSelection();
    }
}
