export default class Template {
    name: string;
    description: string;
    long_description: string;
    blocked: boolean;
    auto_runnable: boolean;
    allow_all_resolver_usernames: boolean;
    allowed_resolver_usernames: string[];
    doc_link: string;
    inputs: any[];
    resolver_inputs: any[];
    steps?: any[];
}