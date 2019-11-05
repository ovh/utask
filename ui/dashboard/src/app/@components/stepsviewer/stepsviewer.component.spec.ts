import { TestBed, async } from '@angular/core/testing';
import { RouterTestingModule } from '@angular/router/testing';
import { StepsViewerComponent } from './stepsviewer.component';
import { GraphService } from 'src/app/@services/graph.service';

describe('StepsViewerComponent', () => {
    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                RouterTestingModule
            ],
            providers: [
                GraphService
            ],
            declarations: [
                StepsViewerComponent
            ],
        }).compileComponents();
    }));

    it('Create component StepsViewerComponent', () => {
        const fixture = TestBed.createComponent(StepsViewerComponent);
        const app = fixture.debugElement.componentInstance;
        expect(app).toBeTruthy();
    });
});
