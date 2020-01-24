import { Component, Input, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import jsYaml from 'js-yaml';
import EditorConfig from 'src/app/@models/editorconfig.model';
import { ApiService } from '../../@services/api.service';
import JSToYaml from 'convert-yaml';

@Component({
  selector: 'app-modal-edit-resolution',
  templateUrl: './modal-edit-resolution.component.html'
})
export class ModalEditResolutionComponent implements OnInit {
  @Input() public value: any;
  errors: any[];
  public text: string;
  public config: EditorConfig = {
    readonly: false,
    mode: 'ace/mode/yaml',
    theme: 'ace/theme/monokai',
    wordwrap: true
  };
  loading = false;
  error = null;

  constructor(public activeModal: NgbActiveModal, private api: ApiService) {
  }

  ngOnInit() {
    JSToYaml.spacingStart = ' '.repeat(0);
    JSToYaml.spacing = ' '.repeat(4);
    this.text = JSToYaml.stringify(this.value).value;
  }

  textUpdate(text: string) {
    this.text = text;
    try {
      jsYaml.safeLoad(text);
      this.errors = [];
    } catch (err) {
      this.errors = [{
        row: err.mark.line,
        column: 0,
        text: err.message,
        type: 'error'
      }];
    }
  }

  submit() {
    try {
      this.loading = true;
      const obj = jsYaml.safeLoad(this.text);
      this.errors = [];
      this.api.putResolution(obj.id, obj).subscribe((data) => {
        this.error = null;
        this.activeModal.close(data);
      }, (err: any) => {
        this.error = err;
      }).add(() => {
        this.loading = false;
      });
    } catch (err) {
      if (err.mark) {
        this.errors = [{
          row: err.mark.line,
          column: 0,
          text: err.message,
          type: 'error'
        }];
      } else {
        console.log('Error', err);
        this.error = err;
      }
      this.loading = false;
    }
  }
}
