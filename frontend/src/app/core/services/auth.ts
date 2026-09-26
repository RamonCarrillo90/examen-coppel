import { computed, inject, Service, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';
import { Observable, tap } from 'rxjs';

import { LoginRequest, LoginResponse } from '../models/auth';
import { Usuario } from '../models/usuario';

const CLAVE_TOKEN = 'token';
const CLAVE_USUARIO = 'usuario';

/**
 * Auth guarda la sesión (token + usuario) y la expone como signals,
 * para que cualquier componente reaccione cuando alguien entra o sale.
 */
@Service()
export class Auth {
  private readonly http = inject(HttpClient);
  private readonly router = inject(Router);

  // Estado privado y escribible; hacia afuera solo se expone de lectura.
  private readonly _token = signal<string | null>(localStorage.getItem(CLAVE_TOKEN));
  private readonly _usuario = signal<Usuario | null>(leerUsuarioGuardado());

  readonly token = this._token.asReadonly();
  readonly usuario = this._usuario.asReadonly();
  readonly estaAutenticado = computed(() => this._token() !== null);
  readonly esAdmin = computed(() => this._usuario()?.rol === 'admin');

  /** Inicia sesión y guarda el token y el usuario si las credenciales son correctas. */
  login(datos: LoginRequest): Observable<LoginResponse> {
    return this.http.post<LoginResponse>('/api/auth/login', datos).pipe(
      tap((res) => {
        localStorage.setItem(CLAVE_TOKEN, res.token);
        localStorage.setItem(CLAVE_USUARIO, JSON.stringify(res.usuario));
        this._token.set(res.token);
        this._usuario.set(res.usuario);
      }),
    );
  }

  /** Cierra la sesión y regresa al login. */
  logout(): void {
    localStorage.removeItem(CLAVE_TOKEN);
    localStorage.removeItem(CLAVE_USUARIO);
    this._token.set(null);
    this._usuario.set(null);
    this.router.navigate(['/login']);
  }
  /** Actualiza los datos del usuario en sesión (por ejemplo, si editó su propio perfil). */
actualizarUsuario(usuario: Usuario): void {
  localStorage.setItem(CLAVE_USUARIO, JSON.stringify(usuario));
  this._usuario.set(usuario);
}

  rutaInicio(): string {
    return this._usuario() ? '/inicio' : '/login';
  }
}

/** Lee el usuario guardado en localStorage; si está corrupto, lo ignora. */
function leerUsuarioGuardado(): Usuario | null {
  try {
    const texto = localStorage.getItem(CLAVE_USUARIO);
    return texto ? (JSON.parse(texto) as Usuario) : null;
  } catch {
    return null;
  }
}