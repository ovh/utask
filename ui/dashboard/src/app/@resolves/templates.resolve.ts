import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { Router } from '@angular/router';
import { catchError } from 'rxjs/operators';
import { /*ActivatedRouteSnapshot, RouterStateSnapshot, */Resolve } from '@angular/router';
import { ApiService } from '../@services/api.service';
import Template from '../@models/template.model';
import { HttpHeaders } from '@angular/common/http';

@Injectable()
export class TemplatesResolve implements Resolve<any> {
    api: ApiService;
    constructor(api: ApiService, private router: Router) {
        this.api = api;
    }

    hasLast(headers: HttpHeaders, pagination: any) {
        const link = headers.get('link');
        if (link) {
            const match = link.match(/last=([^&;\s>]+)/);
            if (match) {
                pagination.last = match[1];
                return true;
            } else {
                pagination.last = '';
                return false;
            }
        } else {
            pagination.last = '';
            return false;
        }
    }

    resolve() {
        return new Promise((resolve, reject) => {
            const pagination = {
                page_size: 1000,
                last: ''
            };
            const load = (p: any, items: Template[] = []) => {
                return this.api.getTemplates(pagination).toPromise().then((data) => {
                    items = items.concat(data.body as Template[]);
                    if (this.hasLast(data.headers, p)) {
                        return load(p, items);
                    } else {
                        return items;
                    }
                }).catch((err) => {
                    throw err;
                });
            };
            load(pagination).then((templates: Template[]) => {
                resolve(templates);
            }).catch((err) => {
                this.router.navigate(['/error']);
                reject(err);
            });
        });
    }
}
