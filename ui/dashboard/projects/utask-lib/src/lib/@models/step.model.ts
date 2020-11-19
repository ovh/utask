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
    resources: any;
    tags: any;
}