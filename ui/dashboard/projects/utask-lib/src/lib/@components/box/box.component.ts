import { Component, Input, OnChanges } from '@angular/core';
import { Router } from '@angular/router';

class HeaderConfig {
    init: boolean;
    openable: boolean;
    link: string;
    openOnClick: boolean;
    class: string;
}

@Component({
    selector: 'lib-utask-box',
    templateUrl: './box.html',
    styleUrls: ['./box.sass'],
})
export class BoxComponent implements OnChanges {
    @Input('header') header: HeaderConfig;
    display: boolean = true;
    headerConfig: HeaderConfig;

    constructor(
        private router: Router
    ) { }

    ngOnChanges() {
        this.display = this.header.init ?? true;
        this.headerConfig = {
            ...this.header,
            openable: false,
            init: true,
            link: '',
            openOnClick: false,
            class: 'primary',
        };
    }

    headerClick() {
        if (this.header.link) {
            this.router.navigate([this.header.link]);
        }
        else if (this.header.openOnClick) {
            this.display = !this.display;
        }
    }
}
