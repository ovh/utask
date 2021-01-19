import { ChangeDetectionStrategy, ChangeDetectorRef, Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { throwError } from 'rxjs';
import { catchError, finalize } from 'rxjs/operators';
import { ApiService } from '../../@services/api.service';

@Component({
  templateUrl: './function.html',
  styleUrls: ['./function.sass'],
  changeDetection: ChangeDetectionStrategy.OnPush
})
export class FunctionComponent implements OnInit {
  functionName: string;
  loading: boolean;
  error: any;
  function: string = '';

  constructor(
    private _route: ActivatedRoute,
    private _api: ApiService,
    private _cd: ChangeDetectorRef
  ) { }

  ngOnInit() {
    this._route.params.subscribe(params => {
      this.functionName = params.functionName;
      this.load();
    });
  }

  load(): void {
    this.loading = true;
    this._cd.markForCheck();
    this._api.function.getYAML(this.functionName)
      .pipe(
        catchError(err => {
          this.error = err;
          return throwError(err);
        }),
        finalize(() => {
          this.loading = false;
          this._cd.markForCheck();
        })
      )
      .subscribe(data => { this.function = data; });
  }
}