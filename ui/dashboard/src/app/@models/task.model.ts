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

export default class Task {
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
    resolution: string;
    resolver_username: string;
    result: any;
    state: string;
    steps_done: number;
    steps_total: number;
    template_name: string;
    title: string;
    watcher_usernames: string[];
    tags: { [key: string]: string };
    
    public constructor(init?: Partial<Task>) {
        Object.assign(this, init);
    }
}
