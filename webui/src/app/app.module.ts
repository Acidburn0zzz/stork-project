import { BrowserModule } from '@angular/platform-browser';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';

import { environment } from './../environments/environment';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';

import { ApiModule, BASE_PATH, Configuration } from './backend';

import {ButtonModule} from 'primeng/button';
import {MenubarModule} from 'primeng/menubar';
import {PanelModule} from 'primeng/panel';
import {TableModule} from 'primeng/table';

import { LoginScreenComponent } from './login-screen/login-screen.component';
import { DashboardComponent } from './dashboard/dashboard.component';
import { HostsTableComponent } from './hosts-table/hosts-table.component';
import { SwaggerUiComponent } from './swagger-ui/swagger-ui.component';

export function cfgFactory() {
    return new Configuration();
}

@NgModule({
    declarations: [
        AppComponent,
        LoginScreenComponent,
        DashboardComponent,
        HostsTableComponent,
        SwaggerUiComponent
    ],
    imports: [
        BrowserModule,
        BrowserAnimationsModule,
        HttpClientModule,
        AppRoutingModule,

        ApiModule.forRoot(cfgFactory),

        ButtonModule,
        MenubarModule,
        PanelModule,
        TableModule,
    ],
    providers: [{ provide: BASE_PATH, useValue: environment.apiUrl }],
    bootstrap: [AppComponent]
})
export class AppModule { }
