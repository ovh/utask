import { AfterViewInit, Component, Input, OnChanges, ViewChild } from '@angular/core';
import { Router } from '@angular/router';
import { NzCollapsePanelComponent } from 'ng-zorro-antd/collapse';

class HeaderConfig {
    init: boolean;
    openable: boolean;
    class: string;
    color: string;
    fontColor: string;
}

@Component({
    selector: 'lib-utask-box',
    templateUrl: './box.html',
    styleUrls: ['./box.sass'],
    standalone: false
})
export class BoxComponent implements OnChanges {
    @ViewChild('panel') panel: NzCollapsePanelComponent;

    @Input() header: HeaderConfig;
    display = true;
    headerConfig: HeaderConfig;

    customStyle: any
    customStylePanel: any
    customStyleHeader: any

    constructor() { }

    ngOnChanges() {
        this.display = this.header.init ?? true;
        this.headerConfig = {
            openable: false,
            init: true,
            class: 'primary',
            ...this.header
        };
        if (this.headerConfig.color && this.headerConfig.fontColor) {
            this.customStyle = {
                'border-color': this.headerConfig.color,
                'background-color': this.headerConfig.color
            };
            this.customStylePanel = {
                'border-color': this.headerConfig.color
            };
            this.customStyleHeader = {
                color: this.headerConfig.fontColor
            }
        }
    }
}
