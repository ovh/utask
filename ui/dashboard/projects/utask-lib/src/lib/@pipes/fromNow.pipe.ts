import { Pipe, PipeTransform } from '@angular/core';
import * as moment_ from 'moment';
const moment = moment_;

@Pipe({ name: 'utaskFromNow' })
export class FromNowPipe implements PipeTransform {
    transform(value: any, params: any = {}): string {
        const options = {
            ...params,
            type: 'date',
            formatDate: null,
            withoutSuffix: false,
            compareDate: null
        };
        let date = moment();
        let compareDate = null;

        if (options.type === 'timestamp') {
            if (options.compareDate) {
                compareDate = moment(+options.compareDate);
            }
            date = moment(+value);
        } else if (options.type === 'timestamp_microsecond') {
            if (options.compareDate) {
                compareDate = moment(+options.compareDate / 1000);
            }
            date = moment(+value / 1000);
        } else if (options.type === 'timestamp_nanosecond') {
            if (options.compareDate) {
                compareDate = moment(+options.compareDate / 1000000);
            }
            date = moment(+value / 1000000);
        } else if (options.type === 'date') {
            if (options.compareDate) {
                compareDate = moment(options.compareDate, options.formatDate);
            }
            date = moment(value, options.formatDate);
        }

        if (!compareDate) {
            return date.fromNow(options.withoutSuffix);
        }
        return date.from(compareDate, options.withoutSuffix);
    }
}