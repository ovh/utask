import { Injectable } from '@angular/core';
import { NzConfigService } from 'ng-zorro-antd/core/config';

@Injectable()
export class ThemeService {
	constructor(private nzConfigService: NzConfigService) { }

	getTheme(): string {
		return localStorage.getItem('utask-theme');
	}

	changeTheme(theme: string): void {
		if (theme !== 'default' && theme !== 'dark') {
			return;
		}
		document.body.className = theme;
		localStorage.setItem('utask-theme', theme);
		if (theme === 'dark') {
			const style = document.createElement('link');
			style.type = 'text/css';
			style.rel = 'stylesheet';
			style.id = 'dark-theme';
			style.href = 'assets/ng-zorro-antd.dark.min.css';
			document.body.appendChild(style);
		} else {
			const dom = document.getElementById('dark-theme');
			if (dom) {
				dom.remove();
			}
		}
		const defaultEditorOption = this.nzConfigService.getConfigForComponent('codeEditor')?.defaultEditorOption || {};
		this.nzConfigService.set('codeEditor', {
			defaultEditorOption: {
				...defaultEditorOption,
				theme: theme === 'dark' ? 'vs-dark' : 'vs',
			}
		});
	}
}
