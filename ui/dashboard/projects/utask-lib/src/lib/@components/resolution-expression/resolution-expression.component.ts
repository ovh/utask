import { ChangeDetectionStrategy, Component, Input } from "@angular/core";
import { FormBuilder, FormGroup, Validators } from "@angular/forms";
import { BehaviorSubject } from "rxjs";
import Resolution from "../../@models/resolution.model";
import { ApiService } from "../../@services/api.service";

@Component({
  selector: "lib-utask-resoution-expression",
  templateUrl: "./resolution-expression.html",
  styleUrls: ["./resolution-expression.sass"],
  changeDetection: ChangeDetectionStrategy.OnPush,
})
export class ResolutionExpressionComponent {
  private _resolution: Resolution;
  private _steps$ = new BehaviorSubject<String[]>([]);
  private _result$ = new BehaviorSubject<String | null>(null);
  private _error$ = new BehaviorSubject<String | null>(null);

  readonly formGroup: FormGroup;

  readonly steps$ = this._steps$.asObservable();
  readonly result$ = this._result$.asObservable();
  readonly error$ = this._error$.asObservable();

  @Input("resolution") set resolution(r: Resolution) {
    if ((this._resolution = r)) {
      this._steps$.next(Object.keys(r.steps));
    } else {
      this._steps$.next([]);
    }
  }

  get resolution(): Resolution {
    return this._resolution;
  }

  constructor(private _api: ApiService, _builder: FormBuilder) {
    this.formGroup = _builder.group({
      step: ["", [Validators.required]],
      expression: ["", [Validators.required]],
    });
  }

  reset(): void {
    this.formGroup.reset();
    this._result$.next(null);
    this._error$.next(null);
  }

  submit(): void {
    const { step, expression } = this.formGroup.value;

    this._api.resolution
      .templating(this._resolution, step, expression)
      .subscribe(
        (result) => {
          if (result.error) {
            this._result$.next(null);
            this._error$.next(result.error);
          } else {
            this._result$.next(result.result);
            this._error$.next(null);
          }
        },
        (e) => {
          this._result$.next(null);
          this._error$.next(e.error.error);
        }
      );
  }
}
