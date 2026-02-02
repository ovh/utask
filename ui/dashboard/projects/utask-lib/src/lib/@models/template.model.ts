export class Template {
    name: string;
    description: string;
    long_description: string;
    blocked: boolean;
    auto_runnable: boolean;
    allow_all_resolver_usernames: boolean;
    allow_task_start_over: boolean;
    allowed_resolver_usernames: string[];
    allowed_resolver_groups: string[];
    doc_link: string;
    inputs: any[];
    resolver_inputs: any[];
    steps?: any[];
}