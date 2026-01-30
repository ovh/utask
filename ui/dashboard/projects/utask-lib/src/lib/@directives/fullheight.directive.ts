import { Directive, ElementRef, OnDestroy, AfterViewInit } from '@angular/core';

@Directive({
    selector: '[fullHeight]',
    standalone: false
})
export class FullHeightDirective implements AfterViewInit, OnDestroy {
    myElement: any;
    subscriptionResize: any;

    constructor(myElement: ElementRef) {
        this.myElement = myElement.nativeElement;
    }

    ngAfterViewInit() {
        this.setHeight();
        this.subscriptionResize = window.addEventListener('resize', () => {
            this.setHeight();
        });
    }

    setHeight() {
        const top = this.myElement.getBoundingClientRect().y
        const windowHeight = document.documentElement.clientHeight;
        this.myElement.style.height = (windowHeight - top) + 'px';
    }

    ngOnDestroy() {
        window.removeEventListener('resize', this.subscriptionResize);
    }
}
