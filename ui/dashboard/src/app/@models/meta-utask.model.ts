export default class MetaUtask {
    application_name: string;
    user_is_admin: boolean;
    username: string;
    
    public constructor(init?: Partial<MetaUtask>) {
        Object.assign(this, init);
    }
}