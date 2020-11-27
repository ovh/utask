export default class Meta {
    application_name: string;
    user_is_admin: boolean;
    username: string;
    
    public constructor(init?: Partial<Meta>) {
        Object.assign(this, init);
    }
}