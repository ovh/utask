import { Injectable, Inject } from '@angular/core';
import Task, { TaskType, TaskState, ResolutionStep, Comment, Stats } from '../@models/task.model';
import Function from '../@models/function.model';
import { Observable } from 'rxjs';
import { HttpClient, HttpResponse } from '@angular/common/http';
import Meta from '../@models/meta.model';
import Template from '../@models/template.model';

export class ParamsListTasks {
    page_size?: number;
    tag?: string[];
    type: TaskType;
    last?: string;
    state?: TaskState;

    public static equals(a: ParamsListTasks, b: ParamsListTasks): boolean {
        return JSON.stringify(a) === JSON.stringify(b);
    }
}

export class NewTask {
    comment: string;
    delay: string;
    input: any;
    tags: { [key: string]: string };
    template_name: string;
    watcher_usernames: string[];
}

export class UpdatedTask {
    input: any;
    tags: { [key: string]: string };
    watcher_usernames: string[];
}

export class ApiServiceComment {
    constructor(
        private _http: HttpClient,
        private _base: string
    ) { }

    add(taskId: string, content: string): Observable<Comment> {
        return this._http.post<Comment>(`${this._base}${taskId}/comment`, {
            content
        });
    }
}

export class ApiServiceTask {
    public comment: ApiServiceComment;

    constructor(
        private http: HttpClient,
        private base: string
    ) {
        this.comment = new ApiServiceComment(this.http, `${this.base}task/`);
    }

    list(params: ParamsListTasks): Observable<HttpResponse<Array<Task>>> {
        return this.http.get<Array<Task>>(`${this.base}task`, {
            params: params as any,
            observe: 'response'
        })
    }

    add(body: NewTask) {
        return this.http.post(
            `${this.base}task`,
            body
        );
    }

    update(id: string, body: UpdatedTask) {
        return this.http.put(
            `${this.base}task/${id}`,
            body
        );
    }

    updateAsYaml(id: string, yaml: string) {
        return this.http.put(
            `${this.base}task/${id}`,
            yaml,
            {
                headers: {
                    accept: 'application/x-yaml',
                },
                responseType: 'text',
                observe: 'body'
            }
        );
    }

    delete(id: string) {
        return this.http.delete(
            `${this.base}task/${id}`
        );
    }

    reject(id: string) {
        return this.http.post(
            `${this.base}task/${id}/wontfix`,
            {}
        );
    }

    get(id: string): Observable<Task> {
        return this.http.get<Task>(`${this.base}task/${id}`);
    }

    getAsYaml(id: string) {
        return this.http.get(`${this.base}task/${id}`, {
            headers: {
                accept: 'application/x-yaml',
            },
            responseType: 'text',
            observe: 'body'
        });
    }
}

export class ParamsListTemplates {
    page_size?: number;
    last?: string;
}

export class ApiServiceTemplate {
    constructor(
        private _http: HttpClient,
        private _base: string
    ) { }

    list(params: ParamsListTemplates): Observable<HttpResponse<Array<Template>>> {
        return this._http.get<Array<Template>>(`${this._base}template`, {
            params: params as any,
            observe: 'response',
        });
    }

    get(name: string): Observable<Template> {
        return this._http.get<Template>(`${this._base}template/${name}`);
    }
}

export class ParamsListFunctions {
    page_size?: number;
    last?: string;
}

export class ApiServiceFunction {
    constructor(
        private _http: HttpClient,
        private _base: string
    ) { }

    list(params: ParamsListFunctions): Observable<HttpResponse<Array<Function>>> {
        return this._http.get<Array<Function>>(`${this._base}function`, {
            params: params as any,
            observe: 'response',
        });
    }

    get(name: string): Observable<Function> {
        return this._http.get<Function>(`${this._base}function/${name}`);
    }
}

export class ApiServiceMeta {
    constructor(
        private _http: HttpClient,
        private _base: string
    ) { }

    get(): Observable<Meta> {
        return this._http.get<Meta>(`${this._base}meta`);
    }
}

export class ApiServiceStats {
    constructor(
        private _http: HttpClient,
        private _base: string
    ) { }

    get(): Observable<Stats> {
        return this._http.get<Stats>(`${this._base}unsecured/stats`);
    }
}

export class NewResolution {
    resolver_inputs: any;
    task_id: string;
}

export class UpdatedResolution {
    resolver_inputs: any;
    steps: { [step: string]: ResolutionStep };
}

export class ApiServiceResolution {
    constructor(private http: HttpClient, private base: string) {
    }

    pause(id: string) {
        return this.http.post(`${this.base}resolution/${id}/pause`, {});
    }

    extend(id: string) {
        return this.http.post(`${this.base}resolution/${id}/extend`, {});
    }

    run(id: string) {
        return this.http.post(`${this.base}resolution/${id}/run`, {});
    }

    cancel(id: string) {
        return this.http.post(`${this.base}resolution/${id}/cancel`, {});
    }

    get(id: string) {
        return this.http.get(`${this.base}resolution/${id}`);
    }

    getAsYaml(id: string) {
        return this.http.get(`${this.base}resolution/${id}`, {
            headers: {
                accept: 'application/x-yaml',
            },
            responseType: 'text',
            observe: 'body'
        });
    }

    update(id: string, resolution: UpdatedResolution) {
        return this.http.put(
            `${this.base}resolution/${id}`,
            resolution
        );
    }

    updateAsYaml(id: string, yaml: string) {
        return this.http.put(
            `${this.base}resolution/${id}`,
            yaml,
            {
                headers: {
                    accept: 'application/x-yaml',
                },
                responseType: 'text',
                observe: 'body'
            }
        );
    }

    add(resolution: NewResolution) {
        return this.http.post(
            `${this.base}resolution`,
            resolution
        );
    }
}

@Injectable({
    providedIn: "root"
})
export class ApiServiceOptions {
    constructor(
        @Inject('apiBaseUrl') public apiBaseUrl: string,
    ) { }
}

@Injectable({
    providedIn: 'root'
})
export class ApiService {
    public meta: ApiServiceMeta;
    public task: ApiServiceTask;
    public resolution: ApiServiceResolution;
    public stats: ApiServiceStats;
    public template: ApiServiceTemplate;
    public function: ApiServiceFunction;
    private apiBaseUrl: string;

    constructor(
        private _http: HttpClient,
        private _options: ApiServiceOptions,
    ) {
        this.apiBaseUrl = this._options.apiBaseUrl;
        this.meta = new ApiServiceMeta(this._http, this.apiBaseUrl);
        this.task = new ApiServiceTask(this._http, this.apiBaseUrl);
        this.resolution = new ApiServiceResolution(this._http, this.apiBaseUrl);
        this.stats = new ApiServiceStats(this._http, this.apiBaseUrl);
        this.template = new ApiServiceTemplate(this._http, this.apiBaseUrl);
        this.function = new ApiServiceFunction(this._http, this.apiBaseUrl);
    }
}
