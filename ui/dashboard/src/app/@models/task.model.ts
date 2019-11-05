export default class Task {
    batch: string;
    comments: any[];
    created: Date;
    errors: any[];
    id: string;
    input: any;
    last_activity: Date;
    last_start: Date;
    last_stop: Date;
    requester_username: string;
    resolution: string;
    resolver_username: string;
    result: any;
    state: string;
    steps_done: number;
    steps_total: number;
    template_name: string;
    title: string;
    watcher_usernames: string[];
}