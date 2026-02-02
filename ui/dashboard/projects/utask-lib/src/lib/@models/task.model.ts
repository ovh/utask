export class ResolutionStep {
    action: any;
    children: any[];
    children_steps: string[];
    children_steps_map: { [map: string]: boolean };
    conditions: any[];
    custom_states: string[];
    dependencies: string[];
    description: string;
    error: string;
    foreach: string;
    idempotent: boolean;
    json_schema: any[];
    last_run: Date;
    max_retries: number;
    name: string;
    resources: string[];
    retry_pattern: string;
    state: string;
    tags: { [tag: string]: boolean };
    try_count: number;
}

export enum TaskType {
    all = 'all',
    own = 'own',
    resolvable = 'resolvable'
};

export enum TaskState {
    BLOCKED = 'BLOCKED',
    CANCELLED = 'CANCELLED',
    DONE = 'DONE',
    RUNNING = 'RUNNING',
    TODO = 'TODO',
    WONTFIX = 'WONTFIX',
    WAITING = 'WAITING',
    DELAYED = 'DELAYED'
};

export class Comment {
    content: string;
    created: Date;
    id: string;
    updated: Date;
    username: string;
}

export class Error {
    error: string;
    step: string;
}

export class ResolverInput {
    name: string;
    description: string;
    regex: string;
    legal_values: any[];
    collection: boolean;
    type: string;
    optional: boolean;
    default: any;
}

export class Task {
    batch: string;
    comments: Comment[];
    created: Date;
    errors: Error[];
    id: string;
    input: any;
    last_activity: Date;
    last_start: Date;
    last_stop: Date;
    requester_username: string;
    resolver_usernames: string[];
    resolver_groups: string[];
    resolution: string;
    resolver_username: string;
    result: any;
    state: string;
    steps_done: number;
    steps_total: number;
    template_name: string;
    title: string;
    watcher_usernames: string[];
    watcher_groups: string[];
    tags: { [key: string]: string };
    resolver_inputs: ResolverInput[];

    public constructor(init?: Partial<Task>) {
        Object.assign(this, init);
    }
}

export class Stats {
    task_states: { [state: string]: number }
}