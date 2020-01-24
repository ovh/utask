import { Component, Input, OnInit } from '@angular/core';
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap';
import jsYaml from 'js-yaml';
import EditorConfig from 'src/app/@models/editorconfig.model';
import { ApiService } from '../../@services/api.service';
import JSToYaml from 'convert-yaml';

@Component({
  selector: 'app-modal-edit-request',
  templateUrl: './modal-edit-request.component.html'
})
export class ModalEditRequestComponent implements OnInit {
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
      this.api.putTask(obj.id, obj).subscribe((data) => {
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
    }
  }
}

/*

            let body = JSON.parse(this.editedTask);
            this.api.apiUtaskTask.put({
                id: this.task.id
            }, body).$promise.then((result) => {
                this.success({
                    item: result.data
                });
            }).catch((err) => {
                this.errors.main = _.get(err, "data.data.error", this.$translate.instant("common.error.message_general"));
            });

                    editRequest (task) {
                        this.$uibModal.open({
                            animation: true,
                            template: `
                            <oui-modal on-dismiss="$ctrl.$uibModalInstance.dismiss();">
                                <utask-task-edit task="$ctrl.task" success="$ctrl.edit(item)"></utask-task-edit>
                            </oui-modal>
                            `,
                            size: this.EnvFactory.config.sizeModal,
                            resolve: {
                                task: () => task,
                            },
                            controllerAs: "$ctrl",
                            controller: class {
                                constructor ($uibModalInstance, task) {
                                    "ngInject";
                                    this.$uibModalInstance = $uibModalInstance;
                                    this.task = task;
                                }
                                edit (item) {
                                    this.$uibModalInstance.close({
                                        $value: item
                                    });
                                }
                            }
                        }).result.then((result) => {
                            this.task = result.$value;
                            this.toast.success("Successfully updated task");
                        }, () => {
                        });
                    }
                  */