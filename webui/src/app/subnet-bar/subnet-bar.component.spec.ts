import { async, ComponentFixture, TestBed } from '@angular/core/testing'

import { SubnetBarComponent } from './subnet-bar.component'

describe('SubnetBarComponent', () => {
    let component: SubnetBarComponent
    let fixture: ComponentFixture<SubnetBarComponent>

    beforeEach(async(() => {
        TestBed.configureTestingModule({
            declarations: [SubnetBarComponent],
        }).compileComponents()
    }))

    beforeEach(() => {
        fixture = TestBed.createComponent(SubnetBarComponent)
        component = fixture.componentInstance
        fixture.detectChanges()
    })

    it('should create', () => {
        expect(component).toBeTruthy()
    })
})
