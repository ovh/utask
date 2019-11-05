import * as brace from 'brace';

export default class EditorConfig {
    readonly?: boolean;
    mode?: string;
    theme?: string;
    tabSize?: number;
    wordwrap?: boolean;
    maxLines?: number;
}
