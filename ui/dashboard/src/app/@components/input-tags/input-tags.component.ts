import { Component, EventEmitter, Output, Input, OnInit, OnDestroy, OnChanges, ElementRef, ViewChild, AfterViewInit, SimpleChanges } from '@angular/core';
import * as _ from 'lodash';
import { TaskService } from 'src/app/@services/task.service';
class Word {
    start: number;
    end: number;
    text: string;
}

class Suggestion {
    word: string;
    atStart: boolean;
}

@Component({
    selector: 'app-input-tags',
    template: `
    <span #span>
        <input type="text" #inputPlaceholder class="placeholder" value="" readonly="" autocomplete="off" spellcheck="false" tabindex="-1"/>
        <input type="text" #inputMain class="main" [placeholder]="displayPlaceholder ? placeholder : ''"/>
        <ul class="suggestions" *ngIf="suggestions.length && !suggestionsHide">
            <li *ngFor="let item of suggestions; let index = index;"  [ngClass]="{'selected': suggestionIndex !== -1 && item.word === suggestions[suggestionIndex].word}" (click)="select(index);">
                {{item.word}}
            </li>
        </ul>
    </span>
    `,
    styleUrls: ['./input-tags.sass'],
})
export class InputTagsComponent implements OnChanges, AfterViewInit {
    tags: string[];
    @Input() value: string[] = [];

    @Input() placeholder: string = '';

    @ViewChild('inputMain') inputMain: ElementRef;
    @ViewChild('inputPlaceholder') inputPlaceholder: ElementRef;
    @ViewChild('span') span: ElementRef;

    @Output() public update: EventEmitter<any> = new EventEmitter();

    lastValueUpdated: string;

    suggestions: Suggestion[] = [];
    suggestionIndex: number = 0;
    suggestionsHide: boolean = false;
    displayPlaceholder: boolean = false;

    constructor(private taskService: TaskService) {
        this.tags = this.taskService.tagsRaw.map(t => t + '=');
        this.taskService.tags.asObservable().subscribe((tags: string[]) => {
            this.tags = tags.map(t => t + '=');
        });
    }

    ngOnChanges(diff: SimpleChanges) {
        if (diff.tags) {
            this.tags = this.tags.map(t => t + '=');
        }
        if (diff.value && this.inputMain) {
            this.inputMain.nativeElement.value = this.value.join(' ');
            this.lastValueUpdated = this.inputMain.nativeElement.value;
        }
    }

    ngAfterViewInit() {
        this.inputMain.nativeElement.addEventListener('input', e => {
            setTimeout(() => {
                this.displayPlaceholder = false;
                this.inputPlaceholder.nativeElement.value = '';
                const lastWord = this.getLastWord(e.target.value);
                if (lastWord) {
                    this.doSearch(e.target.value);
                    const suggestion = this.suggestions[this.suggestionIndex];
                    if (_.get(suggestion, 'atStart')) {
                        const fullTextWithoutLastWord = this.getTextWithoutLastWord(e.target.value);
                        this.inputPlaceholder.nativeElement.value = fullTextWithoutLastWord + suggestion.word;
                    }
                } else {
                    this.suggestionsHide = true;
                    this.suggestions.length = 0;
                }

                if (!e.target.value) {
                    this.displayPlaceholder = true;
                }
            }, 0);
        });
        this.inputMain.nativeElement.addEventListener('keydown', key => {
            const value = key.target.value;
            const lastPosition = this.inputMain.nativeElement.selectionStart === value.length;
            switch (key.key) {
                // top
                case 'ArrowUp':
                    if (this.suggestions && this.suggestionIndex > 0) {
                        this.suggestionIndex--;
                        this.updateText(this.suggestionIndex);
                    }
                    key.preventDefault();
                    break;
                // Down 40
                case 'ArrowDown':
                    key.preventDefault();
                    if (this.suggestionsHide) {
                        this.doSearch(value);
                    } else if (this.suggestions && this.suggestionIndex < this.suggestions.length - 1) {
                        this.suggestionIndex++;
                        this.updateText(this.suggestionIndex);
                    }
                    break;
                // Escape 27
                case 'Escape':
                    this.suggestionsHide = true;
                    this.suggestions.length = 0;
                    break;
                // Enter 13 && Tab 9
                case 'Enter':
                case 'Tab':
                    if (!this.suggestionsHide && this.suggestions.length) {
                        if (this.suggestionIndex === -1) {
                            this.suggestionIndex = 0;
                        }
                        this.updateText(this.suggestionIndex);
                        this.suggestionsHide = true;
                        key.preventDefault();
                    } else {
                        this.updateValue(value);
                    }
                    break;
                // Right 39
                case 'ArrowRight':
                    if (this.suggestions.length && lastPosition) {
                        if (this.suggestionIndex === -1) {
                            this.suggestionIndex = 0;
                        }
                        this.updateText(this.suggestionIndex);
                        this.suggestionsHide = true;
                    } else {
                        this.doSearch(value);
                    }
                    break;
            }
        });
        this.inputMain.nativeElement.addEventListener('blur', e => {
            setTimeout(() => {
                const hasFocus = this.inputMain.nativeElement == document.activeElement;
                if (!hasFocus) {
                    this.suggestionsHide = true;
                    this.updateValue(this.inputMain.nativeElement.value);
                    this.inputPlaceholder.nativeElement.value = '';
                }
            }, 150);
        });
        this.inputMain.nativeElement.addEventListener('focus', e => {
            this.doSearch(this.inputMain.nativeElement.value);
        });
        setTimeout(() => {
            this.inputMain.nativeElement.value = this.value.join(' ');
            this.displayPlaceholder = !this.inputMain.nativeElement.value;
        }, 0);
    }

    updateValue(text: string) {
        if (text !== this.lastValueUpdated) {
            this.update.emit(text);
            this.lastValueUpdated = text;
        }
    }

    select(index: number) {
        this.updateText(index);
        this.suggestions.length = 0;
        this.suggestionIndex = 0;
        this.inputMain.nativeElement.focus();
    }

    doSearch(value: string) {
        const inputValue = value;
        const lastWord = this.getLastWord(inputValue);
        const cursorPosition = this.inputMain.nativeElement.selectionStart;
        if (inputValue) {
            if (lastWord && cursorPosition >= lastWord.start) {
                this.suggestionsHide = false;
                this.suggestions = this.getSuggestions(lastWord.text);
                if (!_.filter(this.suggestions, { atStart: true }).length) {
                    this.suggestionIndex = -1;
                } else if (this.suggestions.length) {
                    this.suggestionIndex = 0;
                } else {
                    this.suggestionIndex = 0;
                }
            } else {
                this.suggestions.length = 0;
                this.suggestionIndex = 0;
            }
        }
    }

    updateText(position: number) {
        this.inputPlaceholder.nativeElement.value = '';
        const suggestion = this.suggestions[position];
        const fullText = this.inputMain.nativeElement.value;
        const fullTextWithoutLastWord = this.getTextWithoutLastWord(fullText);
        this.inputMain.nativeElement.value = fullTextWithoutLastWord + suggestion.word;
    }

    getLastWord(text: string): Word {
        const result = text.match(RegExp(/[^\s]+$/));
        if (result && result[0].indexOf('=') === -1) {
            return {
                text: result[0],
                end: result['index'] + result[0].length - 1,
                start: result['index'],
            };
        }
        return null;
    }

    getTextWithoutLastWord(text: string): string {
        const result = text.match(RegExp(/[^\s]+$/));
        if (result) {
            return text.substring(0, result['index']);
        }
        return '';
    }

    getSuggestions(text: string): Suggestion[] {
        let suggestions = [];
        suggestions = suggestions.concat(_.filter(this.tags, (w: string) => w.toLowerCase().startsWith((text.toLowerCase()))).map(t => ({ word: t, atStart: true })));
        suggestions = suggestions.concat(_.filter(this.tags, (w: string) => w.toLowerCase().indexOf(text.toLowerCase()) > 0).map(t => ({ word: t, atStart: false })));
        return suggestions;
    }
};