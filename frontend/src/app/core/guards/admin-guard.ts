import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { Auth } from '../services/auth';

/**
 * adminGuard solo deja entrar a admins; un usuario normal es enviado a su
 * propia pantalla. Es solo para la experiencia de uso: la seguridad real
 * la aplica el backend (RequireAdmin).
 */
export const adminGuard: CanActivateFn = () => {
  const auth = inject(Auth);
  return auth.esAdmin() ? true : inject(Router).parseUrl(auth.rutaInicio());
};