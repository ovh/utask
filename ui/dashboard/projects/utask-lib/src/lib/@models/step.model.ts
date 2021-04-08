export default class Step {
    name: string;
    description: string;
    idempotent: boolean;
    action: any;
    output: any;
    metadata: any;
    state: string;
    try_count: number;
    max_retries: number;
    last_run: Date;
    dependencies: string[];
    conditions: any[];
    foreach: string;
    resources: string[];
    tags: { [key: string]: string };
    children: string[];
    children_steps: string[];
    children_steps_map: { [key: string]: boolean };
    custom_states: string[];
    error: string;
    execution_delay: string;
    foreach_stategy: string;
    json_schema: any;
    pre_hook: any;
    retry_pattern: string;
}