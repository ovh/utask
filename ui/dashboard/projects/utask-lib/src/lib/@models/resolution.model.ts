import Step from "./step.model";

export default class Resolution {
    id: string;
    resolver_username: string;
    state: string;
    instance_id: number;
    last_start: Date;
    last_stop: Date;
    next_retry: Date;
    run_count: number;
    run_max: number;
    base_configurations: any;
    task_id: string;
    task_title: string;
    steps: { [key: string]: Step };
}