import { Component, OnInit, Input } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import EditorConfig from 'src/app/@models/editorconfig.model';
import JSToYaml from 'convert-yaml';

@Component({
  selector: 'app-modal-yaml-preview',
  templateUrl: './modal-yaml-preview.component.html'
})
export class ModalYamlPreviewComponent implements OnInit {
  @Input() public value: any;
  @Input() public title: string;
  public text: string;
  public config: EditorConfig = {
    readonly: true,
    mode: 'ace/mode/yaml',
    theme: 'ace/theme/monokai',
    wordwrap: true
  };

  constructor(public activeModal: NgbActiveModal) {
  }

  ngOnInit() {
    JSToYaml.spacingStart = ' '.repeat(0);
    JSToYaml.spacing = ' '.repeat(4);
    this.text = JSToYaml.stringify(this.value).value;
  }
}
