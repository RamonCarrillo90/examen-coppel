import { Usuario } from './usuario';

/** Cuerpo de POST /api/auth/login. */
export interface LoginRequest {
  email: string;
  password: string;
}

/** Respuesta de un login exitoso. */
export interface LoginResponse {
  token: string;
  usuario: Usuario;
}