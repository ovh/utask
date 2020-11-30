import {
    Component,
    EventEmitter,
    Output,
    Input,
    ViewChild,
    AfterViewInit,
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    ElementRef
} from '@angular/core';
import { NzSelectComponent } from 'ng-zorro-antd/select';
import { ENTER } from '@angular/cdk/keycodes';

@Component({
    selector: 'lib-utask-input-tags',
    templateUrl: './input-tags.html',
    styleUrls: ['./input-tags.sass'],
    changeDetection: ChangeDetectionStrategy.OnPush
})
export class InputTagsComponent implements AfterViewInit {
    @Input() rawTags: string[];
    @Input() set value(values: string[]) {
        this.values = values;
    }
    @Input() placeholder = '';
    @Output() public update: EventEmitter<Array<string>> = new EventEmitter();

    @ViewChild('selectComponent') public selectComponent: NzSelectComponent;
    @ViewChild('inputTag') public inputTag: ElementRef;

    filteredTags: string[] = [];
    currentTag: string;
    inputTagValid: boolean;
    values: string[] = [];

    constructor(
        private _cd: ChangeDetectorRef
    ) { }

    addItem(): void {
        this.values.push(this.currentTag);
        this._cd.markForCheck();
        this.emitValuesChanged();
    }

    onSelectedTagsChange(data: Array<string>): void {
        this.values = data;
        this._cd.markForCheck();
        this.emitValuesChanged();
    }

    emitValuesChanged(): void {
        this.update.emit(this.values);
    }

    onSelectOpen(state: boolean): void {
        if (!state) { return; }
        this.filteredTags = this.rawTags;
        this.currentTag = '';
        // use a timeout to let the input component appear before executing the callback
        setTimeout(() => { this.inputTag.nativeElement.focus(); }, 0);
    }

    onChangeTagInput(value: string): void {
        this.currentTag = value;
        this.filteredTags = this.rawTags.filter(t => t.toLowerCase().indexOf(value.toLowerCase()) !== -1);
        this.updateInputTagValidity();
    }

    updateInputTagValidity(): void {
        const splittedTagValue = this.currentTag.split('=');
        this.inputTagValid = splittedTagValue.length > 1 && splittedTagValue[1] !== '';
        this._cd.markForCheck();
    }

    onKeyDownTagInput(e: KeyboardEvent): void {
        this.updateInputTagValidity();
        switch (e.keyCode) {
            case ENTER:
                e.preventDefault();
                if (this.inputTagValid) {
                    this.addItem();
                }
                break;
        }
    }

    ngAfterViewInit() {
        this.selectComponent.nzSelectTopControlComponent.nzSelectSearchComponent.disabled = true; // disable text input on the tag selector
        this.selectComponent.onItemClick = () => { }
    }
};