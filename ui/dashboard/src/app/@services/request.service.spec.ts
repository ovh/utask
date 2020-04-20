import { TestBed, async, inject } from '@angular/core/testing';
const d3 = require('d3');
import dagreD3 from 'dagre-d3';
import { Component, ViewChild, ElementRef, AfterViewInit } from '@angular/core';
import { RequestService } from './request.service';
import Task from '../@models/task.model';
import MetaUtask from '../@models/meta-utask.model';
import { NgbModule } from '@ng-bootstrap/ng-bootstrap';

describe('RequestService', () => {
    let requestService: RequestService;

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            imports: [
                NgbModule
            ],
            declarations: [
            ],
            providers: [
                RequestService
            ],
        }).compileComponents();
        requestService = TestBed.get(RequestService);
    }));

    it('Injection request service', () => {
        inject([RequestService], (injectedService: RequestService) => {
            expect(injectedService).toBe(requestService);
        });
    });

    it(`Task with admin role is resolvable`, () => {
        let task: Task = new Task({
        });
        let meta: MetaUtask = new MetaUtask({
            user_is_admin: true,
        });

        expect(true).toEqual(requestService.isResolvable(task, meta, []));
        meta.user_is_admin = false;
    });

    it(`Task without admin role is not resolvable`, () => {
        let task: Task = new Task({
        });
        let meta: MetaUtask = new MetaUtask({
            user_is_admin: false,
        });

        expect(false).toEqual(requestService.isResolvable(task, meta, []));
    });

    it(`Task with resolution is not resolvable`, () => {
        let task: Task = new Task({
            resolution: '2',
        });
        let meta: MetaUtask = new MetaUtask({
            user_is_admin: true,
        });

        expect(false).toEqual(requestService.isResolvable(task, meta, []));
    });

    it(`Task is resolvable if user is in the resolver usernames`, () => {
        let task: Task = new Task({
        });
        let meta: MetaUtask = new MetaUtask({
            user_is_admin: false,
            username: 'login1',
        });
        let resolvers = ['login1'];

        expect(true).toEqual(requestService.isResolvable(task, meta, resolvers));
    });
});
