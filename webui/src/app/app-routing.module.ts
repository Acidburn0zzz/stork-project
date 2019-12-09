import { NgModule } from '@angular/core'
import { Routes, RouterModule } from '@angular/router'

import { AuthGuard } from './auth.guard'
import { DashboardComponent } from './dashboard/dashboard.component'
import { LoginScreenComponent } from './login-screen/login-screen.component'
import { SwaggerUiComponent } from './swagger-ui/swagger-ui.component'
import { MachinesPageComponent } from './machines-page/machines-page.component'
import { UsersPageComponent } from './users-page/users-page.component'
import { ProfilePageComponent } from './profile-page/profile-page.component'
import { PasswordChangePageComponent } from './password-change-page/password-change-page.component'

const routes: Routes = [
    {
        path: '',
        // component: DashboardComponent,
        pathMatch: 'full',
        // canActivate: [AuthGuard],
        redirectTo: 'machines/all',
    },
    {
        path: 'login',
        component: LoginScreenComponent,
    },
    {
        path: 'logout',
        component: LoginScreenComponent,
    },
    {
        path: 'machines',
        pathMatch: 'full',
        redirectTo: 'machines/all',
    },
    {
        path: 'machines/:id',
        component: MachinesPageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'settings',
        component: ProfilePageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'settings/profile',
        component: ProfilePageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'settings/password',
        component: PasswordChangePageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'users',
        redirectTo: 'users/',
        pathMatch: 'full',
    },
    {
        path: 'users/:id',
        component: UsersPageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'users/new',
        component: UsersPageComponent,
        canActivate: [AuthGuard],
    },
    {
        path: 'swagger-ui',
        component: SwaggerUiComponent,
        canActivate: [AuthGuard],
    },

    // otherwise redirect to home
    { path: '**', redirectTo: '' },
]

@NgModule({
    imports: [RouterModule.forRoot(routes)],
    exports: [RouterModule],
})
export class AppRoutingModule {}
