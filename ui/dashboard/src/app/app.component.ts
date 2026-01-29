import { Component, OnInit } from '@angular/core';
import { ThemeService } from './@services/theme.service';

@Component({
    standalone: false,
  selector: 'app-root',
  template: `
    <router-outlet></router-outlet>
  `,
})
export class AppComponent implements OnInit {
  constructor(
    private _themeService: ThemeService
  ) { }

  ngOnInit() {
    const theme = this._themeService.getTheme();
    if (theme) {
      this._themeService.changeTheme(theme);
    }
  }
}
