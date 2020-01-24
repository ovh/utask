import { AfterViewInit, Component, ElementRef, Input, ViewChild, Output, EventEmitter, OnChanges } from '@angular/core';
import * as brace from 'brace';
import 'brace/mode/json';
import 'brace/mode/yaml';
import 'brace/mode/json';
import 'brace/theme/monokai';
import * as _ from 'lodash';
import EditorConfig from 'src/app/@models/editorconfig.model';

// const Range = brace.acequire('ace/range').Range;
// require('brace/ext/language_tools');
// const langTools = brace.acequire('ace/ext/language_tools');

@Component({
    selector: 'editor',
    template: `
    <div #editor>
    </div>
    `,
})
export class EditorComponent implements AfterViewInit, OnChanges {
    @ViewChild('editor', null) editor: ElementRef;
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
        this.aceEditor.getSession().setMode(_.get(this, 'config.mode', 'ace/mode/json'));
        this.aceEditor.setOptions({
            tabSize: _.get(this, 'config.tabSize', 2),
            maxLines: _.get(this, 'config.maxLines', Infinity),
        });
        this.aceEditor.setTheme(_.get(this, 'config.theme', 'ace/theme/monokai'));
        this.aceEditor.setValue(this.value);
        this.aceEditor.setReadOnly(_.get(this, 'config.readonly', false));
        this.aceEditor.getSession().setUseWrapMode(_.get(this, 'config.wordwrap', true));
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
