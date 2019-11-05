import { Component, OnInit, Input } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import JSON2YAML from '../../@services/json2yaml.service';
import EditorConfig from 'src/app/@models/editorconfig.model';

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
    JSON2YAML.setSpacing(0, 4);
    this.text = JSON2YAML.stringify(this.value);
  }
}
