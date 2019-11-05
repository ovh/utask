import * as brace from 'brace';

export default class Editor {
    valid: boolean;
    ace: brace.Editor;
    error: any;
    text: string;
    value: any;
    type: string;
    mode: string;
    theme: string;
    minimumSpacing: number;
    spacing: number;
    markerIds: number[];
    selectedMarkerIds: number[];
    changedDelay: number;
    snippets: any[];
};