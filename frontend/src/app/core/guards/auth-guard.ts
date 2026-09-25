import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';

import { Auth } from '../services/auth';

/** authGuard solo deja entrar si hay sesión; si no, manda al login. */
export const authGuard: CanActivateFn = () => {
  const auth = inject(Auth);
  return auth.estaAutenticado() ? true : inject(Router).createUrlTree(['/login']);
};