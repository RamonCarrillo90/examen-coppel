import { inject } from '@angular/core';
import { Routes } from '@angular/router';

import { adminGuard } from './core/guards/admin-guard';
import { authGuard } from './core/guards/auth-guard';
import { Auth } from './core/services/auth';

export const routes: Routes = [
{
    path: 'login',
    loadComponent: () => import('./pages/login/login').then((m) => m.Login),
},
{
    // Todo lo de aquí adentro requiere sesión.
    path: '',
    canActivate: [authGuard],
    children: [
      // "/" lleva a la pantalla de inicio según el rol.
    { path: '', pathMatch: 'full', redirectTo: () => inject(Auth).rutaInicio() },
    {
        path: 'inicio',
        loadComponent: () => import('./pages/inicio/inicio').then((m) => m.Inicio),
    },
    {
        path: 'tareas',
        canActivate: [adminGuard],
        loadComponent: () =>
        import('./pages/tareas-lista/tareas-lista').then((m) => m.TareasLista),
    },
    {
        path: 'tareas/nueva',
        canActivate: [adminGuard],
        loadComponent: () => import('./pages/tarea-form/tarea-form').then((m) => m.TareaForm),
    },
    {
        path: 'disponibles',
        loadComponent: () =>
        import('./pages/disponibles/disponibles').then((m) => m.Disponibles),
    },
    {
        path: 'usuarios',
        canActivate: [adminGuard],
        loadComponent: () =>
        import('./pages/usuarios-lista/usuarios-lista').then((m) => m.UsuariosLista),
    },
    {
        path: 'usuarios/nuevo',
        canActivate: [adminGuard],
        loadComponent: () =>
        import('./pages/usuario-form/usuario-form').then((m) => m.UsuarioForm),
    },
    {
        path: 'usuarios/:id',
        loadComponent: () =>
        import('./pages/usuario-detalle/usuario-detalle').then((m) => m.UsuarioDetalle),
    },
    {
        path: 'usuarios/:id/editar',
        loadComponent: () =>
        import('./pages/usuario-form/usuario-form').then((m) => m.UsuarioForm),
    },
    {
        path: 'usuarios/:usuarioId/tareas/nueva',
        canActivate: [adminGuard],
        loadComponent: () => import('./pages/tarea-form/tarea-form').then((m) => m.TareaForm),
    },
    {
        path: 'tareas/:id/editar',
        canActivate: [adminGuard],
        loadComponent: () => import('./pages/tarea-form/tarea-form').then((m) => m.TareaForm),
    },
    {
        path: 'solicitudes',
        canActivate: [adminGuard],
        loadComponent: () =>
            import('./pages/solicitudes/solicitudes').then((m) => m.SolicitudesPagina),
    },
    ],
},
{ path: '**', redirectTo: '' },
];