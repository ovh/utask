import { timer } from 'rxjs';

// @Injectable()
export class ActiveIntervalService {
    interval: any;
    inactive: boolean = false;

    constructor() {
    }

    setInterval(foo: any, time: number, callDirectly: boolean) {
        if (callDirectly) {
            foo();
        }
        this.interval = timer(time, time).subscribe((val: any) => {
            if (!document.hidden) {
                this.inactive = false;
                foo();
            }
        });
    }

    stopInterval() {
        this.interval.unsubscribe();
    }
}
