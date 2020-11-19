import { Injectable, Inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { TaskType, TaskState, ResolutionStep } from '../@models/task.model';

export class ParamsListTasks {
    page_size?: number;
    tag?: string[];
    type: TaskType;
    last?: string;
    state?: TaskState;
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
    constructor(private http: HttpClient, private base: string) {
    }

    add(taskId: string, content: string) {
        return this.http.post(
            `${this.base}${taskId}/comment`,
            {
                content
            }
        );
    }
}

export class ApiServiceTask {
    public comment: ApiServiceComment;
    constructor(private http: HttpClient, private base: string) {
        this.comment = new ApiServiceComment(this.http, `${this.base}task/`);
    }

    list(params: ParamsListTasks) {
        return this.http.get(`${this.base}task`, {
            params: params as any,
            observe: 'response',
        });
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

    get(id: string) {
        return this.http.get(`${this.base}task/${id}`);
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
    constructor(private http: HttpClient, private base: string) {
    }

    list(params: ParamsListTemplates) {
        return this.http.get(`${this.base}template`,
            {
                params: params as any,
                observe: 'response',
            });
    }

    get(name: string) {
        return this.http.get(
            `${this.base}template/${name}`
        );
    }
}

export class ParamsListFunctions {
    page_size?: number;
    last?: string;
}

export class ApiServiceFunction {
    constructor(private http: HttpClient, private base: string) {
    }

    list(params: ParamsListFunctions) {
        return this.http.get(`${this.base}function`,
            {
                params: params as any,
                observe: 'response',
            });
    }

    get(name: string) {
        return this.http.get(
            `${this.base}function/${name}`
        );
    }
}

export class ApiServiceMeta {
    constructor(private http: HttpClient, private base: string) {
    }

    get() {
        return this.http.get(`${this.base}meta`);
    }
}

export class ApiServiceStats {
    constructor(private http: HttpClient, private base: string) {
    }

    get() {
        return this.http.get(`${this.base}unsecured/stats`);
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
    ) {

    }
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
        private http: HttpClient,
        private options: ApiServiceOptions,
    ) {
        this.apiBaseUrl = this.options.apiBaseUrl;
        this.meta = new ApiServiceMeta(this.http, this.apiBaseUrl);
        this.task = new ApiServiceTask(this.http, this.apiBaseUrl);
        this.resolution = new ApiServiceResolution(this.http, this.apiBaseUrl);
        this.stats = new ApiServiceStats(this.http, this.apiBaseUrl);
        this.template = new ApiServiceTemplate(this.http, this.apiBaseUrl);
        this.function = new ApiServiceFunction(this.http, this.apiBaseUrl);
    }
}
