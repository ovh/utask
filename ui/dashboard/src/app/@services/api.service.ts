import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { environment } from '../../environments/environment';

@Injectable()
export class ApiService {
    private base = environment.apiBaseUrl;

    constructor(private http: HttpClient) { }

    getMeta() {
        return this.http.get(`${this.base}meta`);
    }

    getStats() {
        return this.http.get(`${this.base}unsecured/stats`);
    }

    getTemplates(params: any) {
        return this.http.get(
            `${this.base}template`,
            {
                params,
                observe: 'response',
            }
        );
    }

    getTemplate(name: string) {
        return this.http.get(
            `${this.base}template/${name}`
        );
    }

    postResolution(body: any) {
        return this.http.post(
            `${this.base}resolution`,
            body
        );
    }

    postTask(body: any) {
        return this.http.post(
            `${this.base}task`,
            body
        );
    }

    rejectTask(taskId: string) {
        return this.http.post(
            `${this.base}task/${taskId}/wontfix`,
            {}
        );
    }

    tasks(params: any) {
        return this.http.get(
            `${this.base}task`,
            {
                params,
                observe: 'response',
            }
        );
    }

    putTask(id: string, obj: any) {
        return this.http.put(
            `${this.base}task/${id}`,
            obj
        );
    }

    deleteTask(id: string) {
        return this.http.delete(
            `${this.base}task/${id}`
        );
    }

    addComment(taskId: string, content: string) {
        return this.http.post(
            `${this.base}task/${taskId}/comment`,
            {
                content
            }
        );
    }

    putResolution(id: string, obj: any) {
        return this.http.put(
            `${this.base}resolution/${id}`,
            obj
        );
    }

    runResolution(id: string) {
        return this.http.post(`${this.base}resolution/${id}/run`, {});
    }

    extendResolution(id: string) {
        return this.http.post(`${this.base}resolution/${id}/extend`, {});
    }

    pauseResolution(id: string) {
        return this.http.post(`${this.base}resolution/${id}/pause`, {});
    }

    cancelResolution(id: string) {
        return this.http.post(`${this.base}resolution/${id}/cancel`, {});
    }

    task(id: string) {
        return this.http.get(`${this.base}task/${id}`);
    }

    resolution(id: string) {
        return this.http.get(`${this.base}resolution/${id}`);
    }
}
