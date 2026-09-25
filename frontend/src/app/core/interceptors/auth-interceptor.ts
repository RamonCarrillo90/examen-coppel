import { HttpErrorResponse, HttpInterceptorFn } from '@angular/common/http';
import { inject } from '@angular/core';
import { catchError, throwError } from 'rxjs';

import { Auth } from '../services/auth';

/**
 * authInterceptor agrega el token a cada petición a la API y, si el backend
 * responde 401 (token vencido o inválido), cierra la sesión.
 */
export const authInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(Auth);
  const token = auth.token();

  const peticion = token
    ? req.clone({ setHeaders: { Authorization: `Bearer ${token}` } })
    : req;

  return next(peticion).pipe(
    catchError((error: HttpErrorResponse) => {
      // En el login, un 401 significa "credenciales inválidas": lo maneja la pantalla.
      const esLogin = req.url.endsWith('/api/auth/login');
      if (error.status === 401 && !esLogin) {
        auth.logout();
      }
      return throwError(() => error);
    }),
  );
};